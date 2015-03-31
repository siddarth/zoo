package zoo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"

	"github.com/mgutz/ansi"
	"github.com/zenazn/goji/web"
)

type Request struct {
	Name               string
	Req                *http.Request
	MungeResponseBytes func([]byte) []byte
}

func Record(mux *web.Mux, requests []*Request) error {
	for _, test := range requests {
		req := test.Req
		testName := test.Name

		zoolog(fmt.Sprintf("[zoo] [%s] generating request and expected response for test", testName))
		reqBytes, err := httputil.DumpRequestOut(req, true)
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
			zoolog(fmt.Sprintf("[zoo] [%s] creating dir: %q", testName, testDir))
			if err := os.Mkdir(testDir, 0755); err != nil {
				return err
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("was going to create a directory at %q but there is a file in its place", testDir)
		}

		requestPath := filepath.Join(testDir, requestFn)
		zoolog(fmt.Sprintf("[zoo] [%s] writing request out to %q", testName, requestPath))

		if err := ioutil.WriteFile(requestPath, reqBytes, 0644); err != nil {
			return err
		}

		if test.MungeResponseBytes != nil {
			zoolog(fmt.Sprintf("[zoo] [%s] munging response bytes", testName))
			repBytes = test.MungeResponseBytes(repBytes)
		}

		expectedResponsePath := filepath.Join(testDir, expectedRepFn)
		zoolog(fmt.Sprintf("[zoo] [%s] writing expected_response out to %q", testName, expectedResponsePath))
		if err := ioutil.WriteFile(expectedResponsePath, repBytes, 0644); err != nil {
			return err
		}
	}

	zoolog(fmt.Sprintf("[zoo] all done"))
	return nil
}

func zoolog(msg string) {
	phosphorize := ansi.ColorFunc("green+h:black")
	fmt.Println(phosphorize(msg))
}
