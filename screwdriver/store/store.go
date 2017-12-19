package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
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
	client  *http.Client
	meta    *api.Command
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
	Meta *api.Command
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("Store API %d %s", e.StatusCode, e.Reason)
}

func commandURL(meta *api.Command) (string, error) {
	if meta.Format == "binary" {
		url := fmt.Sprintf("%scommands/%s%%2F%s/%s", config.SDStoreURL,
			meta.Namespace, meta.Command, meta.Version)
		return url, nil
	}
	return "", fmt.Errorf("The format is not binary")
}

// New returns Store object
func New(meta *api.Command) (Store, error) {
	c, err := newClient(meta)
	if err != nil {
		return nil, err
	}
	return Store(c), nil
}

func newClient(meta *api.Command) (*client, error) {
	baseURL, err := commandURL(meta)
	if err != nil {
		return nil, err
	}
	c := &client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeoutSec * time.Second},
		meta:    meta,
	}
	return c, nil
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

func reviseContentType(ct string) string {
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
	command.Meta = c.meta
	request, err := http.NewRequest("GET", c.baseURL, strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("Failed to create reqeust about command to Store API: %v", err)
	}
	res, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Faied to get command from Store API: %v", err)
	}
	defer res.Body.Close()
	body, err := handleResponse(res)
	if err != nil {
		return nil, err
	}
	command.Type = reviseContentType(res.Header.Get("Content-type"))
	command.Body = body
	return command, nil
}
