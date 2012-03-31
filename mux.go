// Package pat implements a simple URL pattern muxer
package pat

import (
	"net/http"
	"net/url"
)

// PatternServeMux is an HTTP request multiplexer. It matches the URL of each
// incoming request against a list of registered patterns with their associated
// methods and calls the handler for the pattern that most closely matches the
// URL.
//
// Pattern matching attempts each pattern in the order in which they were
// registered.
//
// Patterns may contain literals or captures. Captures start with a colon and
// end with the last character or the first slash encountered.
//
// Example pattern with one capture:
//   /hello/:name
// Will match:
//   /hello/blake
//   /hello/keith
// Will not match:
//   /hello/blake/
//   /hello/blake/foo
//   /foo
//   /foo/bar
//
// Example 2:
//    /hello/:name/
// Will match:
//   /hello/blake/
//   /hello/keith/foo
// Will not match:
//   /hello/blake
//   /hello/keith
//   /foo
//   /foo/bar
//
// Retrieve the capture from the r.URL.Query().Get(":name") in a handler (note
// the colon). If a capture name appears more than once, the additional values
// are appended to the previous values (see
// http://golang.org/pkg/net/url/#Values)
//
// A trivial example server is:
//
//      package main
//
//      import (
//		"io"
//		"net/http"
//		"github.com/bmizerany/pat"
//		"log"
//      )
//
//	// hello world, the web server
//	func HelloServer(w http.ResponseWriter, req *http.Request) {
//		io.WriteString(w, "hello, "+r.URL.Query().Get(":name")+"!\n")
//	}
//
//	func main() {
//		m := pat.New()
//		m.Handle("/hello/:name", http.HandlerFunc(HelloServer))
//		err := http.ListenAndServe(":12345", m)
//		if err != nil {
//			log.Fatal("ListenAndServe: ", err)
//		}
//	}
type PatternServeMux struct {
	handlers map[string][]*patHandler
}

// New returns a new PatternServeMux.
func New() *PatternServeMux {
	return &PatternServeMux{make(map[string][]*patHandler)}
}

// ServeHTTP matched r.URL.Path against its routing table.
func (p *PatternServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ph := range p.handlers[r.Method] {
		if params, ok := ph.try(r.URL.Path); ok {
			if len(params) > 0 {
				r.URL.RawQuery = url.Values(params).Encode() + "&" + r.URL.RawQuery
			}
			ph.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// Get will register a pattern with a handler for GET requests.
func (p *PatternServeMux) Get(pat string, h http.Handler) {
	p.Add("GET", pat, h)
}

// Post will register a pattern with a handler for POST requests.
func (p *PatternServeMux) Post(pat string, h http.Handler) {
	p.Add("POST", pat, h)
}

// Put will register a pattern with a handler for PUT requests.
func (p *PatternServeMux) Put(pat string, h http.Handler) {
	p.Add("PUT", pat, h)
}

// Del will register a pattern with a handler for DELETE requests.
func (p *PatternServeMux) Del(pat string, h http.Handler) {
	p.Add("DELETE", pat, h)
}

// Add will register a pattern with a handler for meth requests.
func (p *PatternServeMux) Add(meth, pat string, h http.Handler) {
	p.handlers[meth] = append(p.handlers[meth], &patHandler{pat, h})
}

type patHandler struct {
	pat string
	http.Handler
}

func (ph *patHandler) try(path string) (url.Values, bool) {
	p := make(url.Values)
	var i, j int
	for i < len(path) {
		switch {
		case j >= len(ph.pat):
			if ph.pat[len(ph.pat)-1] == '/' {
				return p, true
			} else {
				return nil, false
			}
		case ph.pat[j] == ':':
			var name, val string
			name, j = find(ph.pat, '/', j)
			val, i = find(path, '/', i)
			p.Add(name, val)
		case path[i] == ph.pat[j]:
			i++
			j++
		default:
			return nil, false
		}
	}
	if j != len(ph.pat) {
		return nil, false
	}
	return p, true
}

func find(s string, c byte, i int) (string, int) {
	j := i
	for j < len(s) && s[j] != c {
		j++
	}
	return s[i:j], j
}
