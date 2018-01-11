package executor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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
	dummyName        = "name-dummy"
	dummyVersion     = "1.0.1"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
)

var (
	validShell   string
	invalidShell string
)

func setup() {
	config.SDAPIURL = "http://fake.com/v4/"
	config.SDStoreURL = "http://fake.store/v1/"
	config.BaseCommandPath = filepath.Join(os.TempDir(), "sd")
	b, _ := ioutil.ReadFile("testdata/validShell.sh")
	validShell = string(b)
	b, _ = ioutil.ReadFile("testdata/invalidShell.sh")
	invalidShell = string(b)
}

func teardown() {
	os.RemoveAll(config.BaseCommandPath)
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

func (d *dummySDAPIBinary) GetCommand(namespace, name, version string) (*api.Command, error) {
	return dummyAPICommand(binaryFormat), nil
}

type dummySDAPIBroken struct{}

func (d *dummySDAPIBroken) GetCommand(namespace, name, version string) (*api.Command, error) {
	return nil, fmt.Errorf("Something error happen")
}

func dummyAPICommand(format string) (cmd *api.Command) {
	cmd = &api.Command{
		Namespace:   dummyNameSpace,
		Name:        dummyName,
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
	_, err = New(sdapi, []string{"sd-cmd", "ns@cmd/ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. Screwdriver API error
	sdapi = api.API(new(dummySDAPIBroken))
	_, err = New(sdapi, []string{"sd-cmd", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. exec command is invalid
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"sd-cmd", "NO_TYPE", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. exec command is invalid
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"sd-cmd", "exec", "arg", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
