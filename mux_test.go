package pat

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPatMatch(t *testing.T) {
	for i, tt := range []struct {
		pat   string
		u     string
		match bool
		vals  url.Values
	}{
		{"/", "/", true, nil},
		{"/", "/wrong_url", false, nil},
		{"/foo/:name", "/foo/bar", true, url.Values{":name": {"bar"}}},
		{"/foo/:name/baz", "/foo/bar", false, nil},
		{"/foo/:name/bar/", "/foo/keith/bar/baz", true, url.Values{":name": {"keith"}}},
		{"/foo/:name/bar/", "/foo/keith/bar/", true, url.Values{":name": {"keith"}}},
		{"/foo/:name/bar/", "/foo/keith/bar", false, nil},
		{"/foo/:name/baz", "/foo/bar/baz", true, url.Values{":name": {"bar"}}},
		{"/foo/:name/baz/:id", "/foo/bar/baz", false, nil},
		{"/foo/:name/baz/:id", "/foo/bar/baz/123", true, url.Values{":name": {"bar"}, ":id": {"123"}}},
		{"/foo/:name/baz/:name", "/foo/bar/baz/123", true, url.Values{":name": {"bar", "123"}}},
		{"/foo/:name.txt", "/foo/bar.txt", true, url.Values{":name": {"bar"}}},
		{"/foo/:name", "/foo/:bar", true, url.Values{":name": {":bar"}}},
		{"/foo/:a:b", "/foo/val1:val2", true, url.Values{":a": {"val1"}, ":b": {":val2"}}},
		{"/foo/:a.", "/foo/.", true, url.Values{":a": {""}}},
		{"/foo/:a:b", "/foo/:bar", true, url.Values{":a": {""}, ":b": {":bar"}}},
		{"/foo/:a:b:c", "/foo/:bar", true, url.Values{":a": {""}, ":b": {""}, ":c": {":bar"}}},
		{"/foo/::name", "/foo/val1:val2", true, url.Values{":": {"val1"}, ":name": {":val2"}}},
		{"/foo/:name.txt", "/foo/bar/baz.txt", false, nil},
		{"/foo/x:name", "/foo/bar", false, nil},
		{"/foo/x:name", "/foo/xbar", true, url.Values{":name": {"bar"}}},
	} {
		params, ok := (&patHandler{tt.pat, nil}).try(tt.u)
		if !tt.match {
			if ok {
				t.Errorf("[%d] url %q matched pattern %q", i, tt.u, tt.pat)
			}
			continue
		}
		if !ok {
			t.Errorf("[%d] url %q did not match pattern %q", i, tt.u, tt.pat)
			continue
		}
		if tt.vals != nil {
			if !reflect.DeepEqual(params, tt.vals) {
				t.Errorf(
					"[%d] for url %q, pattern %q: got %v; want %v",
					i, tt.u, tt.pat, params, tt.vals,
				)
			}
		}
	}
}

func TestPatRoutingHit(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.Query())
		if got, want := r.URL.Query().Get(":name"), "keith"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}))

	p.ServeHTTP(nil, newRequest("GET", "/foo/keith?a=b", nil))
	if !ok {
		t.Error("handler not called")
	}
}

func TestPatRoutingMethodNotAllowed(t *testing.T) {
	p := New()

	var ok bool
	p.Post("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	p.Put("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	r := httptest.NewRecorder()
	p.ServeHTTP(r, newRequest("GET", "/foo/keith", nil))

	if ok {
		t.Fatal("handler called when it should have not been allowed")
	}
	if r.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d; expected %d", r.Code, http.StatusMethodNotAllowed)
	}

	got := strings.Split(r.Header().Get("Allow"), ", ")
	sort.Strings(got)
	want := []string{"POST", "PUT"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got Allow header %v; want %v", got, want)
	}
}

// Check to make sure we don't pollute the Raw Query when we have no parameters
func TestPatNoParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		if r.URL.RawQuery != "" {
			t.Errorf("RawQuery was %q; should be empty", r.URL.RawQuery)
		}
	}))

	p.ServeHTTP(nil, newRequest("GET", "/foo/", nil))
	if !ok {
		t.Error("handler not called")
	}
}

// Check to make sure we don't pollute the Raw Query when there are parameters but no pattern variables
func TestPatOnlyUserParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		if got, want := r.URL.RawQuery, "a=b"; got != want {
			t.Errorf("for RawQuery: got %q; want %q", got, want)
		}
	}))

	p.ServeHTTP(nil, newRequest("GET", "/foo/?a=b", nil))
	if !ok {
		t.Error("handler not called")
	}
}

func TestPatImplicitRedirect(t *testing.T) {
	p := New()
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r := newRequest("GET", "/foo", nil)
	res := httptest.NewRecorder()
	p.ServeHTTP(res, r)

	if res.Code != 301 {
		t.Errorf("expected Code 301, was %d", res.Code)
	}

	if loc := res.Header().Get("Location"); loc != "/foo/" {
		t.Errorf("expected %q, got %q", "/foo/", loc)
	}

	p = New()
	p.Get("/foo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r = newRequest("GET", "/foo", nil)
	res = httptest.NewRecorder()
	res.Code = 200
	p.ServeHTTP(res, r)

	if res.Code != 200 {
		t.Errorf("expected Code 200, was %d", res.Code)
	}
}

func TestTail(t *testing.T) {
	for i, test := range []struct {
		pat    string
		path   string
		expect string
	}{
		{"/:a/", "/x/y/z", "y/z"},
		{"/:a/", "/x", ""},
		{"/:a/", "/x/", ""},
		{"/:a", "/x/y/z", ""},
		{"/b/:a", "/x/y/z", ""},
		{"/hello/:title/", "/hello/mr/mizerany", "mizerany"},
		{"/:a/", "/x/y/z", "y/z"},
	} {
		tail := Tail(test.pat, test.path)
		if tail != test.expect {
			t.Errorf("failed test %d: Tail(%q, %q) == %q (!= %q)",
				i, test.pat, test.path, tail, test.expect)
		}
	}
}

func newRequest(method, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err)
	}
	return req
}

func TestMethodMatch(t *testing.T) {
	p := New()
	p.Patch("/foo/bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Test to see if we get a 405 Method Not Allowed errors from trying to
	// issue a GET request to a handler that only supports the PATCH method.
	req, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()
	resp.Code = http.StatusMethodNotAllowed
	p.ServeHTTP(resp, req)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected code 405, got %d", resp.Code)
	}

	// Now, test to see if we get a 200 OK from issuing a PATCH request to
	// the same handler.
	req, err = http.NewRequest("PATCH", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp = httptest.NewRecorder()
	resp.Code = http.StatusOK
	p.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected code %d, got %d", http.StatusOK, resp.Code)
	}
}
