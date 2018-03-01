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
	PostCommand(specPath string, commandSpec *util.CommandSpec) (*util.CommandSpec, error)
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
	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on GET: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands", smallSpec.Namespace, smallSpec.Name, smallSpec.Version)

	// Request to api
	const contentType = "application/json"
	// No payload
	payload := bytes.NewBuffer([]byte(""))

	responseSpec, err := c.httpRequest("GET", uri.String(), contentType, payload)
	if err != nil {
		return nil, err
	}

	return responseSpec, nil
}

func specToPayloadBytes(commandSpec *util.CommandSpec) (specBytes []byte, err error) {
	// Generate yaml payload bytes
	commandSpecBytes, err := json.Marshal(&commandSpec)
	if err != nil {
		return nil, err
	}
	var payload util.PayloadYaml
	payload.Yaml = string(commandSpecBytes)
	specBytes, err = json.Marshal(&payload)
	if err != nil {
		return nil, err
	}
	return
}

func writeMultipartYaml(writer *multipart.Writer, specPath string, commandSpec *util.CommandSpec) error {
	// Convert command spec to payload
	specBytes, err := specToPayloadBytes(commandSpec)
	if err != nil {
		return fmt.Errorf("Failed to convert spec to bytes:%v", err)
	}

	// Set yaml filename
	fileNameYaml := filepath.Base(specPath)
	yamlPart, err := writer.CreateFormFile("file", fileNameYaml)
	if err != nil {
		return fmt.Errorf("Failed to load spec file:%v", err)
	}
	// Write yaml part
	yamlPart.Write(specBytes)

	return nil
}

func writeMultipartBin(writer *multipart.Writer, commandSpec *util.CommandSpec) error {
	// Load binary from file
	fileContentsBin, err := util.LoadByte(commandSpec.Binary.File)
	if err != nil {
		return fmt.Errorf("Failed to load binary file:%v", err)
	}

	// Set filename of binary
	fileNameBin := filepath.Base(commandSpec.Binary.File)
	binPart, err := writer.CreateFormFile("file", fileNameBin)
	if err != nil {
		return fmt.Errorf("Failed to create form of binary:%v", err)
	}

	// Write binary part
	binPart.Write(fileContentsBin)

	return nil
}

func (c client) httpRequest(httpMethod string, uri string, contentType string, body *bytes.Buffer) (*util.CommandSpec, error) {
	// Send request
	responseBytes, statusCode, err := c.sendHTTPRequest(httpMethod, uri, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("Post request failed: %v", err)
	}

	// Check response
	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return nil, err
	}

	// Get response
	responseSpec := util.CommandSpec{}
	err = json.Unmarshal(responseBytes, &responseSpec)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json response %v", err)
	}

	return &responseSpec, err
}

func (c client) PostCommand(specPath string, commandSpec *util.CommandSpec) (*util.CommandSpec, error) {
	// Generate payload
	// Prepare body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writeMultipartYaml(writer, specPath, commandSpec)

	var err error
	switch commandSpec.Format {
	case "binary":
		err = writeMultipartBin(writer, commandSpec)
	case "habitat":
		return nil, fmt.Errorf("Posting habitat is not implemented yet")
	case "docker":
		return nil, fmt.Errorf("Posting docker is not implemented yet")
	}

	if err != nil {
		return nil, err
	}

	writer.Close()

	// Generate URL
	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands")

	// Send request
	contentType := writer.FormDataContentType()
	responseSpec, err := c.httpRequest("POST", uri.String(), contentType, body)
	if err != nil {
		return nil, err
	}

	return responseSpec, nil
}

func (c client) sendHTTPRequest(method, url, contentType string, payload *bytes.Buffer) ([]byte, int, error) {
	req, err := http.NewRequest(
		method,
		url,
		payload,
	)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))

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
