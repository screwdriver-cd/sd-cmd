package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
)

const (
	timeoutSec = 10
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(namespace, command, version string) (*Command, error)
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
	Command     string `json:"command"`
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
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Screwdriver API %d %s: %s", e.StatusCode, e.Reason, e.Message)
}

// New returns API object
func New() (API, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}
	return API(c), nil
}

func newClient() (*client, error) {
	c := &client{
		baseURL: config.SDAPIURL,
		jwt:     config.SDToken,
		client:  &http.Client{Timeout: timeoutSec * time.Second},
	}
	return c, nil
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
func (c client) GetCommand(namespace, command, version string) (*Command, error) {
	cmd := new(Command)

	url := fmt.Sprintf("%scommands/%s/%s", c.baseURL, namespace+"%2F"+command, version)
	req, err := http.NewRequest("GET", url, strings.NewReader(""))
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
	body, err := handleResponse(res)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, cmd)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal command: %v", err)
	}
	return cmd, nil
}
