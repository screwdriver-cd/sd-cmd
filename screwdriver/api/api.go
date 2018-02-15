package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	timeoutSec = 10
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error)
	PostCommand(commandSpec []byte) error
}

type client struct {
	baseURL string
	jwt     string
	client  *http.Client
}

// ResponseError is an error response from the Screwdriver API
type ResponseError struct {
	StatusCode int    `json:"statusCode"`
	Reason     string `json:"error"`
	Message    string `json:"message"`
}

// Command is a Screwdriver Command
type Command struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Format      string `json:"format"`
	Habitat     struct {
		Mode    string `json:"mode"`
		Package string `json:"package"`
		Command string `json:"command"`
	} `json:"habitat"`
	Docker struct {
		Image   string `json:"image"`
		Command string `json:"command"`
	} `json:"docker"`
	Binary struct {
		File string `json:"file"`
	} `json:"binary"`
	PipelineId string `json:"pipelineId"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Screwdriver API %d %s: %s", e.StatusCode, e.Reason, e.Message)
}

// New returns API object
func New(sdAPI, sdToken string) API {
	c := newClient(sdAPI, sdToken)
	return API(c)
}

func newClient(sdAPI, sdToken string) *client {
	return &client{
		baseURL: sdAPI,
		jwt:     sdToken,
		client:  &http.Client{Timeout: timeoutSec * time.Second},
	}
}

func handleResponse(res *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading response Body from Screwdriver API: %v", err)
	}

	switch res.StatusCode / 100 {
	case 2:
		return body, nil
	case 4:
		res := new(ResponseError)
		err = json.Unmarshal(body, res)
		if err != nil {
			return nil, fmt.Errorf("Unparseable error response from Screwdriver API: %v", err)
		}
		return nil, res
	case 5:
		return nil, fmt.Errorf("%v: Screwdriver API has internal server error", res.StatusCode)
	default:
		return nil, fmt.Errorf("Unknown error happen while communicate with Screwdriver API")
	}
}

// GetCommand returns Command from Screwdriver API
func (c client) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	uri, err := parseUrl(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on GET: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands", smallSpec.Namespace, smallSpec.Name, smallSpec.Version)

	// Request to api
	req, err := http.NewRequest("GET", uri.String(), strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request about command to Screwdriver API: %v", err)
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to get command from Screwdriver API: %v", err)
	}
	defer res.Body.Close()

	// TODO: refactor
	// res, err := httpRequest("GET", uri.String(), c.jwt, nil)

	// Get response body
	body, err := handleResponse(res)

	if err != nil {
		return nil, err
	}

	cmd := new(util.CommandSpec)
	err = json.Unmarshal(body, cmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal command: %v", err)
	}
	return cmd, nil
}

func (c client) PostCommand(commandSpec []byte) error {
	uri, err := parseUrl(c.baseURL)
	if err != nil {
		return fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands")

	_, err = httpRequest("POST", uri.Path, c.jwt, commandSpec)
	if err != nil {
		return fmt.Errorf("Post request failed: %q", err)
	}

	return nil
}

func parseUrl(urlstr string) (*url.URL, error) {
	uri, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL of Screwdriver API. URL is %q", urlstr)
	}
	return uri, nil
}

func httpRequest(method, url, token string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer([]byte(payload)),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return resp, err
}
