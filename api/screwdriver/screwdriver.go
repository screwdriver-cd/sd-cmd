package screwdriver

import (
	"net/http"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
)

const (
	timeoutSec = 10
)

// API is a Screwdriver API endpoint
type API interface {
	SetJWT() error
	GetCommand(namespace, command, version string) (*Command, error)
}

type api struct {
	baseURL  string
	apiToken string
	jwt      string
	client   *http.Client
}

// JWT is a Screwdriver JWT
type JWT struct {
	Token string `json:"token"`
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
		Binary  string `json:"binary"`
	} `json:"habitat"`
	Docker struct {
		Image string `json:"image"`
	} `json:"docker"`
	Binary struct {
		File string `json:"file"`
	} `json:"binary"`
}

// New return API object
func New() (API, error) {
	api := &api{
		baseURL:  config.SdAPIURL,
		apiToken: config.SdAPIToken,
		client:   &http.Client{Timeout: timeoutSec * time.Second},
	}
	return API(api), nil
}

// SetJWT set jwt to API
func (api *api) SetJWT() error {
	return nil
}

// GetCommand return Command from Screwdriver API
func (api api) GetCommand(namespace, command, version string) (*Command, error) {
	cmd := new(Command)
	cmd.Namespace = namespace
	cmd.Command = command
	cmd.Version = version
	cmd.Format = "binary"
	return cmd, nil
}
