package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
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

func TestNew(t *testing.T) {
	client, err := New(fakeAPIURL, fakeSDToken)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := client.(API); !ok {
		t.Errorf("New does not fulfill API interface")
	}
}

func TestGetCommand(t *testing.T) {
	// success
	c, _ := newClient(fakeAPIURL, fakeSDToken)
	api := API(c)
	ns, cmd, ver := "foo", "bar", "1.0"
	jsonMsg := fmt.Sprintf("{\"namespace\":\"%s\",\"command\":\"%s\",\"version\":\"%s\",\"format\":\"binary\",\"binary\":{\"file\":\"./foobar.sh\"}}", ns, cmd, ver)
	c.client = makeFakeHTTPClient(t, 200, jsonMsg, fmt.Sprintf("/v4/commands/%s/%s/%s", ns, cmd, ver))
	command, err := api.GetCommand(ns, cmd, ver)
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
	_, err = api.GetCommand(ns, cmd, ver)
	if err.Error() != ansMsg {
		t.Errorf("err=%q, want %q", errMsg, ansMsg)
	}

	// failure. check some api response error
	response := []struct {
		code    int
		message string
	}{
		{200, "{{\"namespace\":\"invalid\",\"command\":\"json\",\"version\":\"1.0\",\"format\":\"binary\",\"binary\":{\"file\":\"./foobar.sh\"}}"},
		{403, "{\"statusCode\": 403,\"error\": \"Forbidden\",\"message\": \"Access Denied\"}"},
		{403, "{\"statusCode\": 403,\"error\": \"Forbidden\",\"message\": {\"This error messeg json is broken\"}"},
		{500, "{\"statusCode\": 500,\"error\": \"InternalServerError\",\"message\": \"server error\"}"},
		{200, "{\"statusCode\": 200,\"error\": \"JsonBroken\",{\"message\"}: \"This json is broken\"}"},
		{600, "{\"statusCode\": 403,\"error\": \"Unknown\",\"message\": \"Unknown\"}"},
	}
	for _, res := range response {
		c.client = makeFakeHTTPClient(t, res.code, res.message, "")
		api = API(c)
		_, err := api.GetCommand(ns, cmd, ver)
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

func TestMain(m *testing.M) {
	ret := m.Run()
	os.Exit(ret)
}
