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
	"regexp"
	"sort"
	"strings"

	"github.com/zenazn/goji/web"
)

// Config describes the global zoo config. These can be overridden
// on a per-test-level in TestConf.
type Config struct {
	MatchMode
	MungeRequest func(*http.Request)
	TestConf     map[string]TestConfig
}

// TestConfig describes a zoo-directory-level description of config.
type TestConfig struct {
	MatchMode
	MungeRequest func(*http.Request)
}

// Run is the main method for zoo tests: it takes a gojimux,
// and a map of test name (i.e. the directory in the zoo dir).
func Run(mux *web.Mux, config Config) error {
	tests, err := getTests()
	if err != nil {
		return fmt.Errorf("%+v", err)
	}

	for _, test := range tests {
		mungeRequest := config.MungeRequest
		verificationMode := config.MatchMode

		conf, confExists := config.TestConf[filepath.Base(test)]
		if confExists {
			mungeRequest = conf.MungeRequest
			verificationMode = conf.MatchMode
		}

		in, err := os.Open(path.Join(test, requestFn))
		if err != nil {
			return fmt.Errorf("error opening input %q: %v", test, err)
		}
		req, err := http.ReadRequest(bufio.NewReader(in))
		if err != nil {
			return fmt.Errorf("error parsing request for %q: %v", test, err)
		}
		if mungeRequest != nil {
			log.Printf("munging request")
			mungeRequest(req)
		}

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

		if err := verify(test, repBytes, verificationMode); err != nil {
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
			continue
		}
		dirs = append(dirs, filepath.Join(absPath, fi.Name()))
	}

	sort.Strings(dirs)
	return dirs, nil
}

func normalize(buf []byte) []byte {
	return []byte(strings.TrimSpace(string(buf)))
}

func verify(test string, actual []byte, verificationMode MatchMode) error {
	expected, err := ioutil.ReadFile(path.Join(test, expectedRepFn))
	if err != nil {
		return err
	}
	normalizedExpected := normalize(expected)
	normalizedActual := normalize(actual)

	fmt.Printf("Test %q running in mode %+v", test, verificationMode)
	errNoMatch := fmt.Errorf("responses for %q don't match", test)

	switch verificationMode {
	case Exact:
		if bytes.Compare(normalizedExpected, normalizedActual) != 0 {
			return errNoMatch
		}
	case Regexp:
		matched, err := regexp.Match(string(normalizedExpected), normalizedActual)
		if err != nil {
			return err
		}

		if !matched {
			return errNoMatch
		}
	default:
		return fmt.Errorf("unknown verificationMode: %+v", verificationMode)
	}

	return nil
}
