package zoo

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var mux *web.Mux

func init() {
	Path = "example"
	initGoji()
	rand.Seed(time.Now().UTC().UnixNano())
	mux = goji.DefaultMux
}

func initGoji() {
	goji.Get("/hello/:name", func(c web.C, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
	})
	goji.Get("/random", func(c web.C, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Your random string is: rand_%s!", randSeq(32))
	})
}

func TestZoo(t *testing.T) {
	config := map[string]Config{
		"regexp": Config{
			MatchMode:    Regexp,
			MungeRequest: func(req *http.Request) {},
		},
	}
	err := Run(mux, config)
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestRecord(t *testing.T) {
	uri, _ := url.Parse("/hello/siddarth")

	tests := map[string]*http.Request{
		"test_record": &http.Request{
			Header: map[string][]string{
				"Test-Header": []string{"Test value 1", "Test Value 2"},
			},
			Method: "GET",
			URL:    uri,
		},
	}
	ZooRecord(mux, tests)
}

// h/t: http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
