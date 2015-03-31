dpackage zoo

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/zenazn/goji/web"
)

func Record(mux *web.Mux, tests map[string]*http.Request) error {
	for testName, req := range tests {
		log.Printf("generating request and expected response for test: %q", testName)
		reqBytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			return err
		}

		repRecorder := httptest.NewRecorder()
		mux.ServeHTTP(repRecorder, req)
		rep := http.Response{
			StatusCode: repRecorder.Code,
			Header:     repRecorder.Header(),
			Body:       ioutil.NopCloser(repRecorder.Body),
		}

		repBytes, err := httputil.DumpResponse(&rep, true)
		if err != nil {
			return err
		}

		absPath, err := filepath.Abs(Path)
		if err != nil {
			return err
		}

		testDir := filepath.Join(absPath, testName)

		if fi, err := os.Stat(testDir); err != nil {
			log.Printf("creating dir: %q", testDir)
			if err := os.Mkdir(testDir, 0755); err != nil {
				return err
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("was going to create a directory at %q but there is a file in its place", testDir)
		}

		requestPath := filepath.Join(testDir, requestFn)
		log.Printf("writing request out to %q", requestPath)

		if err := ioutil.WriteFile(requestPath, reqBytes, 0644); err != nil {
			return err
		}

		expectedResponsePath := filepath.Join(testDir, expectedRepFn)
		log.Printf("writing expected_response out to %q", expectedResponsePath)
		if err := ioutil.WriteFile(expectedResponsePath, repBytes, 0644); err != nil {
			return err
		}
	}

	log.Printf("all done")
	return nil
}
