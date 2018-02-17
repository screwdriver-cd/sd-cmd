package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	timeoutSec = 10
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error)
	PostCommand(commandSpec *util.CommandSpec) error
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

func handleResponse(bodyBytes []byte, statusCode int) ([]byte, error) {
	switch statusCode / 100 {
	case 2:
		res := new(util.CommandSpec)
		err := json.Unmarshal(bodyBytes, res)
		if err != nil {
			return nil, fmt.Errorf("Screwdriver API Response unparseable: status=%d, err=%v", statusCode, err)
		}
		return bodyBytes, nil
	case 4:
		res := new(ResponseError)
		err := json.Unmarshal(bodyBytes, res)
		if err != nil {
			return nil, fmt.Errorf("Screwdriver API Response unparseable: status=%d, err=%v", statusCode, err)
		}
		return nil, res
	case 5:
		return nil, fmt.Errorf("Screwdriver API has internal server error: statusCode=%d", statusCode)
	default:
		return nil, fmt.Errorf("Unknown error happen while communicate with Screwdriver API: Statuscode=%d", statusCode)
	}
}

// GetCommand returns Command from Screwdriver API
func (c client) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	uri, err := parseURL(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on GET: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands", smallSpec.Namespace, smallSpec.Name, smallSpec.Version)

	// Request to api
	bodyBytes, statusCode, err := c.httpRequest("GET", uri.String(), c.jwt, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to get http response: %v", err)
	}

	// Get response body
	bodyBytes, err = handleResponse(bodyBytes, statusCode)

	if err != nil {
		return nil, err
	}

	cmd := new(util.CommandSpec)
	err = json.Unmarshal(bodyBytes, cmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal command: %v", err)
	}
	return cmd, nil
}

func (c client) PostCommand(commandSpec *util.CommandSpec) error {
	uri, err := parseURL(c.baseURL)
	if err != nil {
		return fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands")

	payload := util.CommandSpecToJsonBytes(*commandSpec)
	bodyBytes, statusCode, err := c.httpRequest("POST", uri.String(), c.jwt, payload)
	if err != nil {
		return fmt.Errorf("Post request failed: %q", err)
	}

	bodyBytes, err = handleResponse(bodyBytes, statusCode)
	if err != nil {
		return err
	}

	return nil
}

func parseURL(urlstr string) (*url.URL, error) {
	uri, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL of Screwdriver API. URL is %q", urlstr)
	}
	return uri, nil
}

func (c client) httpRequest(method, url, token string, payload []byte) ([]byte, int, error) {
	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer([]byte(payload)),
	)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Faild to get http response: %q", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	statusCode := resp.StatusCode

	defer resp.Body.Close()

	return body, statusCode, err
}
