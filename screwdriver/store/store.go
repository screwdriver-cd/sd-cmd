package store

import (
	"net/http"
	"time"

	sdapi "github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const (
	timeoutSec = 180
)

// API is a Store API endpoint
type API interface {
	GetCommand() (*Command, error)
}

type api struct {
	baseURL string
	client  *http.Client
	meta    *sdapi.Command
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
	Meta *sdapi.Command
}

// New return API object
func New(meta *sdapi.Command) (API, error) {
	api := &api{
		baseURL: "http://store/base/",
		client:  &http.Client{Timeout: timeoutSec * time.Second},
		meta:    meta,
	}
	return API(api), nil
}

// GetCommand return Command from Store API
func (api api) GetCommand() (*Command, error) {
	return nil, nil
}
