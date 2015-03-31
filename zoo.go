// Package zoo provides a series utilities to run tests against a Goji web server.
// The Goji source is available at https://github.com/zenazn/goji, and its godocs
// are available at https://godoc.org/github.com/zenazn/goji
package zoo

const (
	requestFn     = "request"
	expectedRepFn = "expected_response"
	actualRepFn   = "actual_response"
)

// MatchMode describes the different ways Zoo supports verifying matches.
type MatchMode int

const (
	// Exact refers to a verbatim match of response bodies.
	Exact MatchMode = iota

	// Regexp refers to a regexp-compiled match of response bodies.
	Regexp
)

// Path can be set to the directory
var Path = "zoo"
