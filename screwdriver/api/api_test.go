package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/datafortest"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	fakeAPIURL  = "http://fake.com/v4/"
	fakeSDToken = "fake-sd-token"
	fakeJWT     = "fake-jwt"
)

var commandSpecYamlPath = datafortest.TestDataRootPath + "/sd-command.yaml"

func makeFakeHTTPClient(t *testing.T, code int, body, endpoint string) *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, body)
	}))
	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	return &http.Client{Transport: tr}
}

func createSmallSpec(namespace, name, version string) *util.CommandSpec {
	return &util.CommandSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}
}

func createSpecBinary(namespace, name, version, filePath string) *util.CommandSpec {
	b := &util.Binary{
		File: filePath,
	}

	return &util.CommandSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
		Format:    "binary",
		Binary:    b,
	}
}

func TestNew(t *testing.T) {
	client := New(fakeAPIURL, fakeSDToken)
	if _, ok := client.(API); !ok {
		t.Errorf("Client should have API interface")
	}
}

func TestGetCommand(t *testing.T) {
	// case success
	c := newClient(fakeAPIURL, fakeSDToken)
	ns, name, ver := "foo", "bar", "1.0"
	jsonMsg := fmt.Sprintf(`{"namespace":"%s","name":"%s","version":"%s","format":"binary","binary":{"file":"./foobar.sh"}}`, ns, name, ver)
	c.client = makeFakeHTTPClient(t, 200, jsonMsg, fmt.Sprintf("/v4/commands/%s/%s/%s", ns, name, ver))
	api := API(c)

	// request
	command, err := api.GetCommand(createSmallSpec(ns, name, ver))
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if command.Namespace != ns {
		t.Errorf("command.Namespace=%q, want %q", command.Namespace, ns)
	}

	// case failure. check 4xx error message
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)

	// request
	_, err = api.GetCommand(createSmallSpec(ns, name, ver))
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	response := []struct {
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

	// request
	for _, res := range response {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.GetCommand(createSmallSpec(ns, name, ver))
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

func TestSendHTTPRequest(t *testing.T) {
	ns, name, ver := "foo", "bar", "1.0"
	jsonResponse := fmt.Sprintf(`{"namespace":"%s","name":"%s","version":"%s","format":"binary","binary":{"file":"./foobar.sh"}}`, ns, name, ver)
	var fakeHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, jsonResponse)
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
	// case success
	c := newClient(fakeAPIURL, fakeSDToken)
	ns, name, ver := "foo", "bar", "1.0"
	binaryFilePath := datafortest.TestDataRootPath + "/binary/hello"
	responseMsg := fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"binary","binary":{"file":"%s"},"pipelineId":250270}`,
		ns, name, ver, binaryFilePath)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api := API(c)

	// request
	specPath := datafortest.TestDataRootPath + "/yaml/sd-command.yaml"
	_, err := api.PostCommand(specPath, createSpecBinary(ns, name, ver, binaryFilePath))
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case failure. check 4xx error message
	errMsg := `{"statusCode": 403,"error": "Forbidden","message": "Access Denied"}`
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)

	// request
	_, err = api.PostCommand(specPath, createSpecBinary(ns, name, ver, binaryFilePath))
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	response := []struct {
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

	// request
	for _, res := range response {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err = api.PostCommand("dummy", createSpecBinary(ns, name, ver, binaryFilePath))
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}
func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}
