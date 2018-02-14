package api

import (
	"fmt"
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

func makeFakeHTTPClient(t *testing.T, code int, body, endpoint string) *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/json")
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

func createSmallSpec(namespace, name, version string) *util.CommandSpec {
	return &util.CommandSpec{
		Namespace: namespace,
		Name:      name,
		Version:   version,
	}
}

func TestNew(t *testing.T) {
	client := New(fakeAPIURL, fakeSDToken)
	if _, ok := client.(API); !ok {
		t.Errorf("Client should have API interface.")
	}
}

func TestGetCommand(t *testing.T) {
	// success
	c := newClient(fakeAPIURL, fakeSDToken)
	api := API(c)
	ns, name, ver := "foo", "bar", "1.0"
	jsonMsg := fmt.Sprintf("{\"namespace\":\"%s\",\"name\":\"%s\",\"version\":\"%s\",\"format\":\"binary\",\"binary\":{\"file\":\"./foobar.sh\"}}", ns, name, ver)
	c.client = makeFakeHTTPClient(t, 200, jsonMsg, fmt.Sprintf("/v4/commands/%s/%s/%s", ns, name, ver))
	command, err := api.GetCommand(createSmallSpec(ns, name, ver))
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if command.Namespace != ns {
		t.Errorf("command.Namespace=%q, want %q", command.Namespace, ns)
	}

	// failure. check 4xx error message
	errMsg := "{\"statusCode\": 403,\"error\": \"Forbidden\",\"message\": \"Access Denied\"}"
	ansMsg := "Screwdriver API 403 Forbidden: Access Denied"
	c.client = makeFakeHTTPClient(t, 403, errMsg, "")
	api = API(c)
	_, err = api.GetCommand(createSmallSpec(ns, name, ver))
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// failure. check some api response error
	response := []struct {
		code    int
		message string
	}{
		{200, "{{\"namespace\":\"invalid\",\"name\":\"json\",\"version\":\"1.0\",\"format\":\"binary\",\"binary\":{\"file\":\"./foobar.sh\"}}"},
		{403, "{\"statusCode\": 403,\"error\": \"Forbidden\",\"message\": \"Access Denied\"}"},
		{403, "{\"statusCode\": 403,\"error\": \"Forbidden\",\"message\": {\"This error message json is broken\"}"},
		{500, "{\"statusCode\": 500,\"error\": \"InternalServerError\",\"message\": \"server error\"}"},
		{200, "{\"statusCode\": 200,\"error\": \"JsonBroken\",{\"message\"}: \"This json is broken\"}"},
		{600, "{\"statusCode\": 403,\"error\": \"Unknown\",\"message\": \"Unknown\"}"},
	}
	for _, res := range response {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err := api.GetCommand(createSmallSpec(ns, name, ver))
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}
