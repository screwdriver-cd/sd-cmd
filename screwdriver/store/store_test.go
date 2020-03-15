package store

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	retryhttp "github.com/hashicorp/go-retryablehttp"
	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	dummyNameSpace   = "foo-dummy"
	dummyName        = "name-dummy"
	dummyVersion     = "1.1.1"
	dummyFormat      = "dummy-format"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
	dummyMode        = "dummy_mode"
	dummyPackage     = "dummy_org/dummy"
	dummyHart        = "dummy.hart"
	dummyCommand     = "dummy_get"
	fakeArtifactsDir = "http://fake.store/v1/"
	fakeAPIURL       = "http://fake.com/v4/"
	fakeSDToken      = "fake-sd-token"
)

func setup() {
	config.SDStoreURL = fakeArtifactsDir
	config.SDAPIURL = fakeAPIURL
	config.SDToken = fakeSDToken
}

func makeFakeHTTPClient(t *testing.T, code int, body, endpoint, cType string) *retryhttp.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", cType)
		fmt.Fprint(w, body)
	}))
	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			if endpoint == "" {
				return url.Parse(server.URL)
			} else if req.URL.Path == endpoint {
				return url.Parse(server.URL)
			}
			return req.URL, nil
		},
	}
	client := retryhttp.NewClient()
	client.HTTPClient.Transport = tr
	return client
}

func dummyCommandSpec(format string) (spec *util.CommandSpec) {
	spec = &util.CommandSpec{
		Namespace:   dummyNameSpace,
		Name:        dummyName,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}

	switch format {
	case binaryFormat:
		spec.Binary = &util.Binary{
			File: dummyFile,
		}
	case habitatFormat:
		spec.Habitat = &util.Habitat{
			Mode:    dummyMode,
			File:    dummyHart,
			Package: dummyPackage,
			Command: dummyCommand,
		}
	}
	return spec
}

func TestNew(t *testing.T) {
	var spec *util.CommandSpec
	var store Store

	// success new
	spec = dummyCommandSpec(dummyFormat)
	store = New(config.SDStoreURL, spec, config.SDToken)

	if _, ok := store.(Store); !ok {
		t.Errorf("New does not fulfill API interface")
	}
}

func TestGetCommand(t *testing.T) {
	// success
	spec := dummyCommandSpec(binaryFormat)
	c := newClient(config.SDStoreURL, spec, config.SDToken)
	store := Store(c)
	dummyURL := fmt.Sprintf("/v1/commands/%s/%s/%s", dummyNameSpace, dummyName, dummyVersion)
	c.client = makeFakeHTTPClient(t, 200, "Hello World", dummyURL, "text/plain")
	cmd, err := store.GetCommand()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if cmd.Type != "text/plain" {
		t.Errorf("cmd.Type=%q, want %q", cmd.Type, "text/plain")
	}
	if string(cmd.Body) != "Hello World" {
		t.Errorf("cmd.Body=%q, want %q", string(cmd.Body), "Hello World")
	}

	// failure. check 4xx error message
	spec = dummyCommandSpec(binaryFormat)
	c = newClient(config.SDStoreURL, spec, config.SDToken)
	c.client = makeFakeHTTPClient(t, 404, "{\"statusCode\": 404, \"error\": \"Not Found\"}", "", "text/plain")
	store = Store(c)
	_, err = store.GetCommand()
	if err.Error() != "Store API 404 Not Found" {
		t.Errorf("err=%q, want %q", err.Error(), "Store API 404 Not Found")
	}

	// failure. check some api response error
	spec = dummyCommandSpec(binaryFormat)
	c = newClient(config.SDStoreURL, spec, config.SDToken)
	clients := []*retryhttp.Client{
		makeFakeHTTPClient(t, 404, "{\"statusCode\": 404, \"error\": \"Not Found\"}", "", "text/plain"),
		makeFakeHTTPClient(t, 500, "ERROR", "", "text/plain"),
		makeFakeHTTPClient(t, 600, "ERROR", "", "text/plain"),
		makeFakeHTTPClient(t, 404, "{{\"statusCode\": 404, \"error\": \"Not Found\"}", "", "text/plain"),
	}
	for _, client := range clients {
		c.client = client
		store = Store(c)
		_, err := store.GetCommand()
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}

	// failure.
	spec = dummyCommandSpec(dockerFormat)
	c = newClient(config.SDStoreURL, spec, config.SDToken)
	store = Store(c)
	_, err = store.GetCommand()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	os.Exit(ret)
}
