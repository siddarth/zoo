package zoo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var mux *web.Mux

func init() {
	Path = "example"
	goji.Get("/hello/:name", hello)
	mux = goji.DefaultMux
}

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func TestZoo(t *testing.T) {
	err := Run(mux, func(req *http.Request) {})
	if err != nil {
		t.Errorf("%+v", err)
	}
}
