package pat

import (
	"net/http"
	"net/url"
)

type PatternServeMux struct {
	handlers map[string][]*patHandler
}

func New() *PatternServeMux {
	return &PatternServeMux{make(map[string][]*patHandler)}
}

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

func (p *PatternServeMux) Get(pat string, h http.Handler) {
	p.Add("GET", pat, h)
}

func (p *PatternServeMux) Post(pat string, h http.Handler) {
	p.Add("POST", pat, h)
}

func (p *PatternServeMux) Put(pat string, h http.Handler) {
	p.Add("PUT", pat, h)
}

func (p *PatternServeMux) Del(pat string, h http.Handler) {
	p.Add("DELETE", pat, h)
}

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
		case j > len(ph.pat):
			return nil, false
		case j == len(ph.pat) && ph.pat[j-1] == '/':
			return p, true
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
