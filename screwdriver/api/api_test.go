package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	fakeAPIURL  = "http://fake.com/v4/"
	fakeSDToken = "fake-sd-token"
	fakeJWT     = "fake-jwt"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	habitatModeRemote = "remote"
	habitatModeLocal  = "local"
)

const (
	dummyNamespace      = "foo-dummy"
	dummyName           = "name-dummy"
	dummyVersion        = "1.0.1"
	dummyDescription    = "dummy description"
	dummyBinaryFileName = "sd-step"
	dummyBinaryFile     = "/dummy/" + dummyBinaryFileName
	dummyDockerImage    = "chefdk:1.2.3"
	dummyDockerCmd      = "knife"
	dummyHabitatMode    = habitatModeRemote
	dummyHabitatPkg     = "core/git/2.14.1"
	dummyHabitatCmd     = "git"
)

var (
	binarySpecYamlPath        = "../testdata/yaml/binary-sd-command.yaml"
	habitatRemoteSpecYamlPath = "../testdata/yaml/habitat-remote-sd-command.yaml"
	habitatLocalSpecYamlPath  = "../testdata/yaml/habitat-local-sd-command.yaml"
	dockerSpecYamlPath        = "../testdata/yaml/docker-sd-command.yaml"
	binaryFilePath            = "../../testdata/binary/hello"
	habitatPackagePath        = "../../testdata/binary/hello.hart"
)

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
		cmd.Docker.Command = dummyDockerCmd
		cmd.Docker.Image = dummyDockerImage
	case habitatFormat:
		cmd.Habitat = new(util.Habitat)
		cmd.Habitat.Command = dummyHabitatCmd
		cmd.Habitat.Mode = dummyHabitatMode
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
		_, err = api.GetCommand(dummySmallSpec())
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
	_, err := api.PostCommand(binarySpecYamlPath, spec)
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
	_, err = api.PostCommand(dockerSpecYamlPath, spec)
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

	_, err = api.PostCommand(habitatRemoteSpecYamlPath, spec)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case success habitat local
	spec = dummySpec(habitatFormat)
	spec.Habitat.Mode = habitatModeLocal
	spec.Habitat.Package = habitatPackagePath
	responseMsg = fmt.Sprintf(`{"id":76,"namespace":"%s","name":"%s",
		"version":"%s","description":"foobar","maintainer":"foo@yahoo-corp.jp",
		"format":"habitat","habitat":{"mode":"%s", "package":"%s", "command":"%s"},"pipelineId":250271}`,
		spec.Namespace, spec.Name, spec.Version, spec.Habitat.Mode, spec.Habitat.Package, spec.Habitat.Command)
	c.client = makeFakeHTTPClient(t, 200, responseMsg, "/v4/commands")
	api = API(c)

	_, err = api.PostCommand(habitatLocalSpecYamlPath, spec)
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
	_, err = api.PostCommand(binarySpecYamlPath, spec)
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// case failure. check some api response error
	spec = dummySpec(binaryFormat)
	spec.Binary.File = binaryFilePath
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
		_, err = api.PostCommand("dummy", spec)
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}

	// case failure. unknown format
	spec = dummySpec("unknown")
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)
	_, err = api.PostCommand(binarySpecYamlPath, spec)
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestValidateCommand(t *testing.T) {
	// case success
	c := newClient(fakeAPIURL, fakeSDToken)
	responseMsg := fmt.Sprintf(`{"errors": [] }`)
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
		_, err = api.ValidateCommand("dummy")
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}
func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}
