package store

import (
	"net/http"
	"time"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const (
	timeoutSec = 180
)

// API is a Store API endpoint
type API interface {
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

// New returns API object
func New(meta *api.Command) (API, error) {
	c := &client{
		baseURL: "http://store/base/",
		client:  &http.Client{Timeout: timeoutSec * time.Second},
		meta:    meta,
	}
	return API(c), nil
}

// GetCommand returns Command from Store API
func (c client) GetCommand() (*Command, error) {
	return nil, nil
}
