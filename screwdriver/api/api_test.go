package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	retryhttp "github.com/hashicorp/go-retryablehttp"
	"github.com/screwdriver-cd/sd-cmd/util"
	"github.com/stretchr/testify/assert"
)

const (
	binaryFormat      = "binary"
	dockerFormat      = "docker"
	habitatFormat     = "habitat"
	habitatModeRemote = "remote"
	habitatModeLocal  = "local"
)

const (
	fakeAPIURL       = "http://fake.com/v4/"
	fakeSDToken      = "fake-sd-token"
	dummyNamespace   = "foo-dummy"
	dummyName        = "name-dummy"
	dummyVersion     = "1.0.1"
	dummyDescription = "dummy description"
	dummyBinaryFile  = "/dummy/sd-step"
	dummyDockerImage = "chefdk:1.2.3"
	dummyHabitatMode = habitatModeRemote
	dummyHart        = "dummy/dummy.hart"
	dummyHabitatPkg  = "core/git/2.14.1"
	dummyCmd         = "dummy-command"
	dummyTag         = "stable"
)

var (
	binaryFilePath     = "../../testdata/binary/hello"
	habitatPackagePath = "../../testdata/binary/hello.hart"
	errorResponses     = []struct {
		code    int
		message string
	}{
		{200, `{{"namespace":"invalid","name":"json","version":"1.0","format":"binary","binary":{"file":"./foobar.sh"}}`},
		{403, `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`},
		{403, `{"statusCode": 403,"error": "Forbidden","message": {"This error message json is broken"}`},
		{500, `{"statusCode": 500,"error": "InternalServerError","message": "server error"}`},
		{200, `{"statusCode": 200,"error": "JsonBroken",{"message"}: "This json is broken"}`},
		{600, `{"statusCode": 403,"error": "Unknown","message": "Unknown"}`},
	}
)

func makeFakeHTTPClient(t *testing.T, code int, body, endpoint string) *retryhttp.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client := retryhttp.NewClient()
	client.HTTPClient.Transport = tr
	return client
}

func dummySmallSpec() (cmd *util.CommandSpec) {
	return &util.CommandSpec{
		Namespace: dummyNamespace,
		Name:      dummyName,
		Version:   dummyVersion,
	}
}

func dummySpec(format string) (cmd *util.CommandSpec) {
	cmd = &util.CommandSpec{
		Namespace:   dummyNamespace,
		Name:        dummyName,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}
	switch format {
	case binaryFormat:
		cmd.Binary = new(util.Binary)
		cmd.Binary.File = dummyBinaryFile
	case dockerFormat:
		cmd.Docker = new(util.Docker)
		cmd.Docker.Command = dummyCmd
		cmd.Docker.Image = dummyDockerImage
	case habitatFormat:
		cmd.Habitat = new(util.Habitat)
		cmd.Habitat.Command = dummyCmd
		cmd.Habitat.Mode = dummyHabitatMode
		cmd.Habitat.File = dummyHart
		cmd.Habitat.Package = dummyHabitatPkg
	}
	return cmd
}

func TestNew(t *testing.T) {
	client := New(fakeAPIURL, fakeSDToken)
	if _, ok := client.(API); !ok {
		t.Errorf("Client should have API interface")
	}
}

func TestGetCommand(t *testing.T) {
	c := newClient(fakeAPIURL, fakeSDToken)
	// case success
	smallSpec := dummySmallSpec()
	jsonMsg := fmt.Sprintf(`{"namespace":"%s","name":"%s","version":"%s","format":"binary","binary":{"file":"./foobar.sh"}}`,
		smallSpec.Namespace, smallSpec.Name, smallSpec.Version)
	c.client = makeFakeHTTPClient(t, 200, jsonMsg, fmt.Sprintf("/v4/commands/%s/%s/%s", smallSpec.Namespace, smallSpec.Name, smallSpec.Version))
	api := API(c)

	// request
	command, err := api.GetCommand(smallSpec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if command.Namespace != smallSpec.Namespace {
		t.Errorf("command.Namespace=%q, want %q", command.Namespace, smallSpec.Namespace)
	}

	// case failure. check 4xx error message
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)

	// request
	_, err = api.GetCommand(dummySmallSpec())
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	for _, res := range errorResponses {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.GetCommand(dummySmallSpec())
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

// TestRetryCommand - makes three attempts to make a http request
// the first two will fail and the final one will succeed
func TestRetryCommand(t *testing.T) {
	c := newErrorRetryClient(fakeAPIURL, fakeSDToken)

	// case success
	smallSpec := dummySmallSpec()
	api := API(c)

	// request
	_, err := api.GetCommand(smallSpec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestGetBinPath(t *testing.T) {
	testCases := []struct {
		specPath     string
		filePath     string
		expectedPath string
	}{
		{"sd-command.yaml", "hello", "hello"},
		{"sd-command.yaml", "./hello", "hello"},
		// Note: allow an absolute path.
		{"sd-command.yaml", "/usr/local/bin/hello", "/usr/local/bin/hello"},
		{"./sd-command.yaml", "hello", "hello"},
		{"../../testdata/yaml/sd-command.yaml", "bin/hello", "../../testdata/yaml/bin/hello"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedPath, getBinPath(tc.specPath, tc.filePath))
	}
}

func TestSendHTTPRequest(t *testing.T) {
	ns, name, ver := "foo", "bar", "1.0"
	jsonResponse := fmt.Sprintf(`{"namespace":"%s","name":"%s","version":"%s","format":"binary","binary":{"file":"./foobar.sh"}}`, ns, name, ver)
	var fakeHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, jsonResponse)
	})

	testServer := httptest.NewServer(fakeHandler)
	defer testServer.Close()

	c := newClient(fakeAPIURL, fakeSDToken)
	c.client = makeFakeHTTPClient(t, 403, "foo", "")

	// GET
	method := "GET"
	contentType := "application/json"
	// No payload
	payload := bytes.NewBuffer([]byte(""))

	_, _, err := c.sendHTTPRequest(method, testServer.URL, contentType, payload)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestPostCommand(t *testing.T) {
	c := newClient(fakeAPIURL, fakeSDToken)

	// case success binary
	spec := dummySpec(binaryFormat)
	spec.Binary.File = binaryFilePath
	responseMsg := fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"binary","binary":{"file":"%s"},"pipelineId":250270}`,
		spec.Namespace, spec.Name, spec.Version, spec.Binary.File)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api := API(c)
	_, err := api.PostCommand(spec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case success docker
	spec = dummySpec(dockerFormat)
	responseMsg = fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"docker","docker":{"image":"%s", "command":"%s"},"pipelineId":250271}`,
		spec.Namespace, spec.Name, spec.Version, spec.Docker.Image, spec.Docker.Command)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api = API(c)
	_, err = api.PostCommand(spec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case success habitat remote
	spec = dummySpec(habitatFormat)
	responseMsg = fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"habitat","habitat":{"mode":"%s", "package":"%s", "command":"%s"},"pipelineId":250271}`,
		spec.Namespace, spec.Name, spec.Version, spec.Habitat.Mode, spec.Habitat.Package, spec.Habitat.Command)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api = API(c)

	_, err = api.PostCommand(spec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case success habitat local
	spec = dummySpec(habitatFormat)
	spec.Habitat.Mode = habitatModeLocal
	spec.Habitat.File = habitatPackagePath
	responseMsg = fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"habitat","habitat":{"mode":"%s", "package":"%s", "command":"%s"},"pipelineId":250271}`,
		spec.Namespace, spec.Name, spec.Version, spec.Habitat.Mode, spec.Habitat.Package, spec.Habitat.Command)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api = API(c)

	_, err = api.PostCommand(spec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case failure. check 4xx error message
	spec = dummySpec(binaryFormat)
	spec.Binary.File = binaryFilePath
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)
	_, err = api.PostCommand(spec)
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	spec = dummySpec(binaryFormat)
	spec.Binary.File = binaryFilePath
	for _, res := range errorResponses {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.PostCommand(spec)
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}

	// case failure. unknown format
	spec = dummySpec("unknown")
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)
	_, err = api.PostCommand(spec)
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestValidateCommand(t *testing.T) {
	// case success
	c := newClient(fakeAPIURL, fakeSDToken)
	responseMsg := `{"errors": [] }`
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/validate/command")
	api := API(c)

	// request
	yamlPath := "../../testdata/yaml/sd-command.yaml"
	yamlByte, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		t.Errorf("err=:%q", err)
	}
	yamlString := string(yamlByte)

	_, err = api.ValidateCommand(yamlString)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case failure. check 4xx error message
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)

	// request
	_, err = api.ValidateCommand(yamlString)
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	for _, res := range errorResponses {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.ValidateCommand("dummy")
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

func TestTagCommand(t *testing.T) {
	// case success
	c := newClient(fakeAPIURL, fakeSDToken)
	responseMsg := `{"errors": [] }`
	dummyEndpoint := path.Join("/v4/commands/", dummyNamespace, dummyName, "tags", dummyTag)
	spec := dummySpec(binaryFormat)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, dummyEndpoint)
	api := API(c)

	// request
	_, err := api.TagCommand(spec, dummyVersion, dummyTag)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case failure. check 4xx error message
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)

	// request
	_, err = api.TagCommand(spec, dummyVersion, dummyTag)
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	for _, res := range errorResponses {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.TagCommand(spec, dummyVersion, dummyTag)
		if err == nil {
			println(res.code, res.message)
			t.Errorf("err=nil, want error")
		}
	}
}

// RoundTripFunc .
type RoundTripTest struct {
	Attempts int
}

// RoundTrip - will fail twice and succeed once
// this exercises the retryable http client
func (f *RoundTripTest) RoundTrip(req *http.Request) (*http.Response, error) {
	ns, name, ver := "foo", "bar", "1.0"
	jsonResponse := fmt.Sprintf(`{"namespace":"%s","name":"%s","version":"%s","format":"binary","binary":{"file":"./foobar.sh"}}`, ns, name, ver)
	if f.Attempts > 1 {
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString(jsonResponse))}, nil
	}

	e := &url.Error{
		Op:  "Get",
		URL: req.URL.RawPath,
		Err: &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: fmt.Errorf("%s", "i/o timeout"),
		},
	}
	f.Attempts++
	return nil, e
}

func newErrorRetryClient(sdAPI, sdToken string) *client {
	rt := &RoundTripTest{}
	tr := rt

	rc := retryhttp.NewClient()
	rc.Logger = log.New(os.Stderr, "DEBUG", log.Lshortfile)
	rc.HTTPClient = &http.Client{
		Transport: tr,
	}

	return &client{
		baseURL: sdAPI,
		jwt:     sdToken,
		client:  rc,
	}
}

func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}
