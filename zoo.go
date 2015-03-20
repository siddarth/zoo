package zoo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"

	"github.com/zenazn/goji/web"
)

const (
	requestFn     = "request"
	expectedRepFn = "expected_response"
	actualRepFn   = "actual_response"
)

// Path can be set to the directory
var Path = "zoo"

// Run is the main method for zoo tests: it takes a testing.T, a gojimux,
// and a func that is called on every request.
func Run(mux *web.Mux, mungeRequest func(*http.Request)) error {
	tests, err := getTests()
	if err != nil {
		return fmt.Errorf("%+v", err)
	}

	for _, test := range tests {
		in, err := os.Open(path.Join(test, requestFn))
		if err != nil {
			return fmt.Errorf("error opening input %q: %v", test, err)
			continue
		}
		req, err := http.ReadRequest(bufio.NewReader(in))
		if err != nil {
			return fmt.Errorf("error parsing request for %q: %v", test, err)
			continue
		}
		mungeRequest(req)

		rep := httptest.NewRecorder()
		mux.ServeHTTP(rep, req)
		if rep.Code != 200 {
			log.Printf("%q returned non-200: %v", test, rep.Body)
		}

		// Get the body of the response out to a buffer.
		httpRep := http.Response{
			StatusCode: rep.Code,
			Header:     rep.Header(),
			Body:       ioutil.NopCloser(rep.Body),
		}
		repBytes, err := httputil.DumpResponse(&httpRep, true)
		if err != nil {
			return fmt.Errorf("failed to dump response %q: %v", test, err)
		}

		// Dump the actual response to disk for ease of debugging
		err = ioutil.WriteFile(path.Join(test, actualRepFn), repBytes, 0666)
		if err != nil {
			return fmt.Errorf("error writing actual_response %q: %v", test, err)
		}

		if err := verify(test, repBytes); err != nil {
			return err
		}
	}

	return nil
}

func getTests() ([]string, error) {
	var dirs []string
	absPath, _ := filepath.Abs(Path)

	fis, err := ioutil.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			return nil, fmt.Errorf("unexpected non-dir file in zoo path: %s", fi.Name())
		}
		dirs = append(dirs, filepath.Join(absPath, fi.Name()))
	}

	return dirs, nil
}

func verify(test string, actual []byte) error {
	expected, err := ioutil.ReadFile(path.Join(test, expectedRepFn))
	if err != nil {
		return err
	}

	if bytes.Compare(expected, actual) != 0 {
		return fmt.Errorf("responses for %q don't match", test)
	}

	return nil
}
