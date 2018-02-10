package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	fakeSDToken = "fake-sd-token"
	fakeJWT     = "fake-jwt"
)

var fakeHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello HTTP Test")
})

func TestHttpPost(t *testing.T) {
	ts := httptest.NewServer(fakeHandler)
	defer ts.Close()
	jsonStr := `{"test":foo","data":"bar"}`
	fakePayload := []byte(jsonStr)
	HttpPost(ts.URL, fakeSDToken, fakePayload)
}
