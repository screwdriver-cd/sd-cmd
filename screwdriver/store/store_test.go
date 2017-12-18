package store

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

var (
	cacheEnv map[string]string
)

const (
	dummyNameSpace   = "foo-dummy"
	dummyCommand     = "cmd-dummy"
	dummyVersion     = "1.1.1"
	dummyFormat      = "binary"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
	fakeArtifactsDir = "http://fake.store/v1/"
	fakeAPIURL       = "http://fake.com/v4/"
	fakeAPIToken     = "fake-api-token"
)

func setup() {
	config.SDStoreURL = fakeArtifactsDir
	config.SDAPIURL = fakeAPIURL
	config.SDAPIToken = fakeAPIToken
}

func makeFakeHTTPClient(t *testing.T, code int, body, endpoint, cType string) *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", cType)
		fmt.Fprintf(w, body)
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
	return &http.Client{Transport: tr}
}

func dummySDCommand() (meta *api.Command) {
	meta = &api.Command{
		Namespace:   dummyNameSpace,
		Command:     dummyCommand,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      dummyFormat,
	}
	meta.Binary.File = dummyFile
	return meta
}

func TestNew(t *testing.T) {
	var sdCommand *api.Command
	var store Store
	var err error

	// success
	sdCommand = dummySDCommand()
	store, err = New(sdCommand)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := store.(Store); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// failure
	sdCommand.Format = "docker"
	store, err = New(sdCommand)
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestCommand(t *testing.T) {
	// success
	sdCommand := dummySDCommand()
	c, _ := newClient(sdCommand)
	store := Store(c)
	dummyURL := fmt.Sprintf("/v1/commands/%s/%s/%s", dummyNameSpace, dummyCommand, dummyVersion)
	c.client = makeFakeHTTPClient(t, 200, "Hello World", dummyURL, "text/plain")
	binary, err := store.GetCommand()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if binary.Type != "text/plain" {
		t.Errorf("binary.Type=%q, want %q", binary.Type, "text/plain")
	}
	if string(binary.Body) != "Hello World" {
		t.Errorf("binary.Body=%q, want %q", string(binary.Body), "Hello World")
	}

	// failure. check 4xx error message
	sdCommand = dummySDCommand()
	c, _ = newClient(sdCommand)
	c.client = makeFakeHTTPClient(t, 404, "{\"statusCode\": 404, \"error\": \"Not Found\"}", "", "text/plain")
	store = Store(c)
	_, err = store.GetCommand()
	if err.Error() != "Store API 404 Not Found" {
		t.Errorf("err=%q, want %q", err.Error(), "Store API 404 Not Found")
	}

	// failure. check some api response error
	sdCommand = dummySDCommand()
	c, _ = newClient(sdCommand)
	clients := []*http.Client{
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
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	os.Exit(ret)
}
