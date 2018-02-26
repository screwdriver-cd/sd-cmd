package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	timeoutSec = 10
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error)
	PostCommand(specPath string, commandSpec *util.CommandSpec) (string, error)
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
	const contentType = "application/json"
	// No payload
	payload := bytes.NewBuffer([]byte(""))

	bodyBytes, statusCode, err := c.httpRequest("GET", uri.String(), c.jwt, contentType, payload)

	if err != nil {
		return nil, fmt.Errorf("Failed to get http response: %v", err)
	}

	// Check response body
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

func specToPayloadBytes(commandSpec *util.CommandSpec) (specBytes []byte, err error) {
	// Generate yaml payload bytes
	commandSpecBytes, err := json.Marshal(&commandSpec)
	var payload util.PayloadYaml
	payload.Yaml = string(commandSpecBytes)
	specBytes, err = json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	return
}

func (c client) PostCommand(specPath string, commandSpec *util.CommandSpec) (string, error) {
	// Generate URL
	uri, err := parseURL(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands")

	// Generate payload

	// Prepare body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Convert command spec to payload
	specBytes, err := specToPayloadBytes(commandSpec)
	if err != nil {
		return "", fmt.Errorf("Failed to convert spec to bytes:%v", err)
	}

	// Set yaml filename
	fileNameYaml := filepath.Base(specPath)
	yamlPart, err := writer.CreateFormFile("file", fileNameYaml)
	if err != nil {
		return "", fmt.Errorf("Failed to load spec file:%v", err)
	}

	// Write yaml part
	yamlPart.Write(specBytes)

	if commandSpec.Format == "binary" {
		// Load binary from file
		fileContentsBin, err := util.LoadByte(commandSpec.Binary.File)
		if err != nil {
			return "", fmt.Errorf("Failed to load binary file:%v", err)
		}

		// Set filename of binary
		fileNameBin := filepath.Base(commandSpec.Binary.File)
		binPart, err := writer.CreateFormFile("file", fileNameBin)
		if err != nil {
			return "", fmt.Errorf("Failed to create form of binary:%v", err)
		}

		// Write binary part
		binPart.Write(fileContentsBin)
	}
	writer.Close()

	// Send request
	contentType := writer.FormDataContentType()
	responseBytes, statusCode, err := c.httpRequest("POST", uri.String(), c.jwt, contentType, body)
	if err != nil {
		return "", fmt.Errorf("Post request failed: %v", err)
	}

	// Check response
	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return "", err
	}

	// Get version of posted command
	responseSpec := util.CommandSpec{}
	err = json.Unmarshal(responseBytes, &responseSpec)
	if err != nil {
		return "", fmt.Errorf("Failed to parse json response %v", err)
	}

	return responseSpec.Version, nil
}

func parseURL(urlstr string) (*url.URL, error) {
	uri, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL of Screwdriver API. URL is %v", urlstr)
	}
	return uri, nil
}

func (c client) httpRequest(method, url, token string, contentType string, payload *bytes.Buffer) ([]byte, int, error) {
	req, err := http.NewRequest(
		method,
		url,
		payload,
	)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get http response: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to read response body: %v", err)
	}

	statusCode := resp.StatusCode

	defer resp.Body.Close()

	return body, statusCode, err
}
