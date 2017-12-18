package executor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	dummyNameSpace   = "foo-dummy"
	dummyCommand     = "cmd-dummy"
	dummyVersion     = "1.0.1"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
)

func setup() {
	config.SDAPIURL = "http://fake.com/v4/"
	config.SDStoreURL = "http://fake.store/v1/"
	b, _ := ioutil.ReadFile("testdata/validShell.sh")
	validShell = string(b)
	b, _ = ioutil.ReadFile("testdata/invalidShell.sh")
	invalidShell = string(b)
}

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

type dummySDAPIBinary struct{}

func (d *dummySDAPIBinary) SetJWT() error { return nil }

func (d *dummySDAPIBinary) GetCommand(namespace, command, version string) (*api.Command, error) {
	return dummySDCommand(binaryFormat), nil
}

type dummySDAPIBroken struct{}

func (d *dummySDAPIBroken) SetJWT() error { return nil }
func (d *dummySDAPIBroken) GetCommand(namespace, command, version string) (*api.Command, error) {
	return nil, fmt.Errorf("Something error happen")
}

func dummySDCommand(format string) (cmd *api.Command) {
	cmd = &api.Command{
		Namespace:   dummyNameSpace,
		Command:     dummyCommand,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}
	cmd.Binary.File = dummyFile
	return cmd
}

func TestNew(t *testing.T) {
	// success
	sdapi := api.API(new(dummySDAPIBinary))
	executor, err := New(sdapi, []string{"ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := executor.(Executor); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// success
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"exec", "ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// failure. no command
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid command
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"ns@cmd/ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. Screwdriver API error
	sdapi = api.API(new(dummySDAPIBroken))
	_, err = New(sdapi, []string{"ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	os.Exit(ret)
}
