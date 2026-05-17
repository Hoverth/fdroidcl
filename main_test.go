package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"mvdan.cc/fdroidcl/adb"
)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) != "fdroidcl" {
		// start the static http server once
		path := filepath.Join("testdata", "staticrepo")
		fs := http.FileServer(http.Dir(path))
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The files are static, so add a unique etag for each file.
			w.Header().Set("Etag", strconv.Quote(r.URL.Path))
			fs.ServeHTTP(w, r)
		})
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
		server := &http.Server{Handler: handler}
		go server.Serve(ln)
		// Save it to a global, which will be added as an env var to be
		// picked up by the children processes.
		staticRepoHost = ln.Addr().String()
	} else {
		httpClient.Transport = repoTransport{os.Getenv("REPO_HOST")}
	}

	os.Exit(testscript.RunMain(m, map[string]func() int{
		"fdroidcl": main1,
	}))
}

type repoTransport struct {
	repoHost string
}

func (t repoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// replace https://f-droid.org/repo/foo with http://localhost:1234/foo
	req.URL.Scheme = "http"
	req.URL.Host = t.repoHost
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/repo")
	return http.DefaultClient.Do(req)
}

var staticRepoHost string

func TestScripts(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Setup: func(e *testscript.Env) error {
			home := e.WorkDir + "/home"
			if err := os.MkdirAll(home, 0o777); err != nil {
				return err
			}
			e.Vars = append(e.Vars, "HOME="+home)
			e.Vars = append(e.Vars, "AppData="+home)
			e.Vars = append(e.Vars, "LocalAppData="+home)
			e.Vars = append(e.Vars, "REPO_HOST="+staticRepoHost)
			return nil
		},
		Condition: func(cond string) (bool, error) {
			switch cond {
			case "device":
				devices, err := adb.Devices()
				return err == nil && len(devices) == 1, nil
			}
			return false, fmt.Errorf("unknown condition %q", cond)
		},
	})
}
