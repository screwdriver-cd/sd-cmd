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

	retryhttp "github.com/hashicorp/go-retryablehttp"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	timeoutSec         = 180
	defaultContentType = "application/json"
)

// API is a Screwdriver API endpoint
type API interface {
	GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error)
	PostCommand(commandSpec *util.CommandSpec) (*util.CommandSpec, error)
	ValidateCommand(yamlString string) (*util.ValidateResponse, error)
	TagCommand(commandSpec *util.CommandSpec, targetVersion, tag string) (*util.TagResponse, error)
	RemoveTagCommand(commandSpec *util.CommandSpec, tag string) (*util.TagResponse, error)
	SetVerbose(isVerbose bool)
}

type client struct {
	baseURL   string
	jwt       string
	client    *retryhttp.Client
	isVerbose bool
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
	c := retryhttp.NewClient()
	c.HTTPClient = &http.Client{Timeout: timeoutSec * time.Second}
	return &client{
		baseURL:   sdAPI,
		jwt:       sdToken,
		client:    c,
		isVerbose: false,
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

	// No payload
	payload := new(bytes.Buffer)

	responseSpec, err := c.httpRequest("GET", uri.String(), defaultContentType, payload)
	if err != nil {
		return nil, err
	}

	return responseSpec, nil
}

func writeValidateBody(yamlString string) (bodyBuff *bytes.Buffer, err error) {
	var payload util.PayloadYaml
	payload.Yaml = yamlString
	validateBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	bodyBuff = bytes.NewBuffer(validateBytes)
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
	if err != nil {
		return nil, err
	}
	bodyBuff = bytes.NewBuffer(specBytes)
	return
}

func versionToPayLoadBuf(targetVersion string) (bodyBuff *bytes.Buffer, err error) {
	body := util.TagTargetVersion{
		Version: targetVersion,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyBuff = bytes.NewBuffer(bodyBytes)
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

func getBinPath(specPath string, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}
	return filepath.Join(filepath.Dir(specPath), filePath)
}

func writeMultipartBin(writer *multipart.Writer, commandSpec *util.CommandSpec) error {
	var filePath string
	switch commandSpec.Format {
	case "binary":
		filePath = commandSpec.Binary.File
	case "habitat":
		filePath = commandSpec.Habitat.File
	}

	// normalize the binary file path as a relative path from the spec yaml.
	filePath = getBinPath(commandSpec.SpecYamlPath, filePath)
	fileContents, err := util.LoadByte(filePath)
	if err != nil {
		return fmt.Errorf("Failed to load file:%v", err)
	}

	fileName := filepath.Base(filePath)
	filePart, err := writer.CreateFormFile("file", fileName)
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
		contentType = defaultContentType
	default:
		return nil, fmt.Errorf(`Unknown "Format" value of command spec: %v`, err)
	}
	if err != nil {
		return nil, err
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
	body, err := writeValidateBody(yamlString)
	if err != nil {
		return nil, err
	}

	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL on POST: %v", err)
	}
	uri.Path = path.Join(uri.Path, "validator/command")

	responseBytes, statusCode, err := c.sendHTTPRequest("POST", uri.String(), defaultContentType, body)
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

func (c client) TagCommand(commandSpec *util.CommandSpec, targetVersion, tag string) (res *util.TagResponse, err error) {
	body, err := versionToPayLoadBuf(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("Failed to create payload: %v", err)
	}

	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands", commandSpec.Namespace, commandSpec.Name, "tags", tag)

	responseBytes, statusCode, err := c.sendHTTPRequest("PUT", uri.String(), defaultContentType, body)
	if err != nil {
		return
	}

	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return
	}

	res = new(util.TagResponse)
	err = json.Unmarshal(responseBytes, res)
	return
}

func (c client) RemoveTagCommand(commandSpec *util.CommandSpec, tag string) (res *util.TagResponse, err error) {
	// No payload
	payload := new(bytes.Buffer)

	uri, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse URL: %v", err)
	}
	uri.Path = path.Join(uri.Path, "commands", commandSpec.Namespace, commandSpec.Name, "tags", tag)

	responseBytes, statusCode, err := c.sendHTTPRequest("DELETE", uri.String(), defaultContentType, payload)
	if err != nil {
		return
	}

	responseBytes, err = handleResponse(responseBytes, statusCode)
	if err != nil {
		return
	}

	res = new(util.TagResponse)
	err = json.Unmarshal(responseBytes, res)
	return
}

func (c client) sendHTTPRequest(method, url, contentType string, payload *bytes.Buffer) ([]byte, int, error) {
	req, err := retryhttp.NewRequest(method, url, payload)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.jwt))

	if !c.isVerbose {
		// Disable http debug output on non verbose mode
		c.client.Logger = nil
	}

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

func (c *client) SetVerbose(isVerbose bool) {
	c.isVerbose = isVerbose
}
