package zoo

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
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
	testConf := map[string]TestConfig{
		"regexp":      TestConfig{Regexp, nil},
		"test_record": TestConfig{Regexp, nil},
	}

	config := Config{
		MatchMode:    Exact,
		MungeRequest: func(req *http.Request) {},
		TestConf:     testConf,
	}
	err := Run(mux, config)
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func TestRecord(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost/random", nil)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	tests := []*Request{
		&Request{
			Name: "test_record",
			Req:  req,
			MungeResponseBytes: func(in []byte) []byte {
				pattern := "rand_\\w+"
				reg := regexp.MustCompile(pattern)
				return []byte(reg.ReplaceAllString(string(in), pattern))
			},
		},
	}
	if err := Record(mux, tests); err != nil {
		t.Fatalf("%+v", err)
	}

	if err := Run(mux, Config{
		MatchMode:    Regexp,
		MungeRequest: func(req *http.Request) {},
	}); err != nil {
		t.Fatalf("%+v", err)
	}
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
