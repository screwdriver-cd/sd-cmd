package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	retryhttp "github.com/hashicorp/go-retryablehttp"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	timeoutSec = 180
)

// Store is a Store API endpoint
type Store interface {
	GetCommand() (*Command, error)
}

type client struct {
	baseURL string
	client  *retryhttp.Client
	spec    *util.CommandSpec
	jwt     string
}

// ResponseError is an error response from the Store API
type ResponseError struct {
	StatusCode int    `json:"statusCode"`
	Reason     string `json:"error"`
}

// Command is a Store Command
type Command struct {
	Type string
	Body []byte
	Spec *util.CommandSpec
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Store API %d %s", e.StatusCode, e.Reason)
}

func (c *client) commandURL() (string, error) {
	if c.spec != nil && (c.spec.Format == "binary" || c.spec.Format == "habitat") {
		uri, err := url.Parse(c.baseURL)
		if err != nil {
			return "", fmt.Errorf("The base Screwdriver Store API is invalid %q", c.baseURL)
		}
		uri.Path = path.Join(uri.Path, "commands", c.spec.Namespace, c.spec.Name, c.spec.Version)
		return uri.String(), nil
	}
	return "", fmt.Errorf("The format %s is not expected", c.spec.Format)
}

// New returns Store object
func New(baseURL string, spec *util.CommandSpec, sdToken string) Store {
	c := newClient(baseURL, spec, sdToken)
	return Store(c)
}

func newClient(baseURL string, spec *util.CommandSpec, sdToken string) *client {
	cli := retryhttp.NewClient()
	cli.HTTPClient.Timeout = timeoutSec * time.Second
	return &client{
		baseURL: baseURL,
		client:  cli,
		spec:    spec,
		jwt:     sdToken,
	}
}

func handleResponse(res *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Reading response Body from Store API: %v", err)
	}

	switch res.StatusCode / 100 {
	case 2:
		return body, nil
	case 4:
		res := new(ResponseError)
		err = json.Unmarshal(body, res)
		if err != nil {
			return nil, fmt.Errorf("Unparseable error response from Store API: %v", err)
		}
		return nil, res
	case 5:
		return nil, fmt.Errorf("%v: Store API has internal server error", res.StatusCode)
	default:
		return nil, fmt.Errorf("Unknown error happen while communicate with Store API")
	}
}

// parseContentType returns content type of Command from Store API
// ex("text/plain; charset=UTF-8" => "text/plain")
func parseContentType(ct string) string {
	vers := strings.Split(ct, ";")
	for _, str := range vers {
		if strings.Contains(str, "/") {
			return strings.Trim(str, " ")
		}
	}
	return ""
}

// GetCommand returns Command from Store API
func (c client) GetCommand() (*Command, error) {
	command := new(Command)
	command.Spec = c.spec
	uri, err := c.commandURL()
	if err != nil {
		return nil, err
	}
	request, err := retryhttp.NewRequest("GET", uri, strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("Failed to create request about command to Store API: %v", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))
	res, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to get command from Store API: %v", err)
	}
	defer res.Body.Close()
	body, err := handleResponse(res)
	if err != nil {
		return nil, err
	}
	command.Type = parseContentType(res.Header.Get("Content-type"))
	command.Body = body
	return command, nil
}
