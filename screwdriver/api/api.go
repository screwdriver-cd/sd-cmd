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
	timeoutSec = 180
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error)
	PostCommand(commandSpec *util.CommandSpec) (*util.CommandSpec, error)
	ValidateCommand(yamlString string) (*util.ValidateResponse, error)
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

	const contentType = "application/json"
	// No payload
	payload := new(bytes.Buffer)

	responseSpec, err := c.httpRequest("GET", uri.String(), contentType, payload)
	if err != nil {
		return nil, err
	}

	return responseSpec, nil
}

func writeValidateBody(yamlString string) (bodyBuff *bytes.Buffer, err error) {
	var payload util.PayloadYaml
	payload.Yaml = yamlString
	validateBytes, err := json.Marshal(payload)
	bodyBuff = bytes.NewBuffer(validateBytes)
	if err != nil {
		return nil, err
	}
	return
}

func specToPayloadBuf(commandSpec *util.CommandSpec) (bodyBuff *bytes.Buffer, err error) {
	commandSpecBytes, err := json.Marshal(commandSpec)
	if err != nil {
		return nil, err
	}
	var payload util.PayloadYaml
	payload.Yaml = string(commandSpecBytes)
	specBytes, err := json.Marshal(payload)
	bodyBuff = bytes.NewBuffer(specBytes)
	if err != nil {
		return nil, err
	}
	return
}

func writeMultipartYaml(writer *multipart.Writer, commandSpec *util.CommandSpec) error {
	specBytes, err := json.Marshal(commandSpec)
	if err != nil {
		return fmt.Errorf("Failed to convert spec to bytes:%v", err)
	}

	yamlPart, err := writer.CreateFormField("spec")
	if err != nil {
		return fmt.Errorf("Failed to load spec file:%v", err)
	}

	_, err = yamlPart.Write(specBytes)
	if err != nil {
		return fmt.Errorf("Failed to write yaml part to the request body:%v", err)
	}

	return nil
}

func writeMultipartBin(writer *multipart.Writer, commandSpec *util.CommandSpec) error {
	var filePath string
	switch commandSpec.Format {
	case "binary":
		filePath = commandSpec.Binary.File
	case "habitat":
		filePath = commandSpec.Habitat.File
	}

	fileContents, err := util.LoadByte(filePath)
	if err != nil {
		return fmt.Errorf("Failed to load file:%v", err)
	}

	fileName := filepath.Base(filePath)
	filePart, err := writer.CreateFormFile(commandSpec.Format, fileName)
	if err != nil {
		return fmt.Errorf("Failed to create form of %s:%v", commandSpec.Format, err)
	}

	_, err = filePart.Write(fileContents)
	if err != nil {
		return fmt.Errorf("Failed to write file contents to form:%v", err)
	}

	return nil
}

func (c client) httpRequest(httpMethod string, uri string, contentType string, body *bytes.Buffer) (*util.CommandSpec, error) {
	responseBytes, statusCode, err := c.sendHTTPRequest(httpMethod, uri, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("Post request failed: %v", err)
	}

	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return nil, err
	}

	responseSpec := &util.CommandSpec{}
	err = json.Unmarshal(responseBytes, responseSpec)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse json response %v", err)
	}

	return responseSpec, err
}

func writeMultipart(commandSpec *util.CommandSpec) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writeMultipartYaml(writer, commandSpec)
	if err != nil {
		return nil, "", err
	}

	err = writeMultipartBin(writer, commandSpec)
	if err != nil {
		return nil, "", err
	}
	writer.Close()

	contentType := writer.FormDataContentType()

	return body, contentType, nil
}

func (c client) PostCommand(commandSpec *util.CommandSpec) (*util.CommandSpec, error) {
	var (
		body        *bytes.Buffer
		contentType string
		err         error
	)
	switch commandSpec.Format {
	case "binary":
		body, contentType, err = writeMultipart(commandSpec)
	case "habitat":
		if commandSpec.Habitat.Mode == "local" {
			body, contentType, err = writeMultipart(commandSpec)
		} else {
			body, err = specToPayloadBuf(commandSpec)
			contentType = "application/json"
		}
	case "docker":
		body, err = specToPayloadBuf(commandSpec)
		contentType = "application/json"
	default:
		return nil, fmt.Errorf(`Unknown "Format" value of command spec: %v`, err)
	}

	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands")

	responseSpec, err := c.httpRequest("POST", uri.String(), contentType, body)
	if err != nil {
		return nil, err
	}

	return responseSpec, nil
}

func (c client) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	contentType := "application/json"
	body, err := writeValidateBody(yamlString)

	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "validator/command")

	responseBytes, statusCode, err := c.sendHTTPRequest("POST", uri.String(), contentType, body)
	if err != nil {
		return nil, fmt.Errorf("Post request failed: %v", err)
	}

	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return nil, err
	}
	res := new(util.ValidateResponse)
	err = json.Unmarshal(responseBytes, res)
	if err != nil {
		return nil, fmt.Errorf("Screwdriver API Response unparseable: status=%d, err=%v", statusCode, err)
	}
	return res, nil
}

func (c client) sendHTTPRequest(method, url, contentType string, payload *bytes.Buffer) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, payload)
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
