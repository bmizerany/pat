package pat

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPatMatch(t *testing.T) {
	params, ok := (&patHandler{"/", nil}).try("/")
	expect(t, true, ok)

	params, ok = (&patHandler{"/", nil}).try("/wrong_url")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/:name", nil}).try("/foo/bar")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar/baz")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar/")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil}).try("/foo/keith/bar")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar/baz")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz/123")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar"}, ":id": {"123"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:name", nil}).try("/foo/bar/baz/123")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar", "123"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil}).try("/foo/bar.txt")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name", nil}).try("/foo/:bar")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil}).try("/foo/val1:val2")
	expect(t, true, ok)
	deepExpect(t, url.Values{":a": {"val1"}, ":b": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:a.", nil}).try("/foo/.")
	expect(t, true, ok)
	deepExpect(t, url.Values{":a": {""}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil}).try("/foo/:bar")
	expect(t, true, ok)
	deepExpect(t, url.Values{":a": {""}, ":b": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b:c", nil}).try("/foo/:bar")
	expect(t, true, ok)
	deepExpect(t, url.Values{":a": {""}, ":b": {""}, ":c": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/::name", nil}).try("/foo/val1:val2")
	expect(t, true, ok)
	deepExpect(t, url.Values{":": {"val1"}, ":name": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil}).try("/foo/bar/baz.txt")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/bar")
	expect(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/xbar")
	expect(t, true, ok)
	deepExpect(t, url.Values{":name": {"bar"}}, params)
}

func TestPatRoutingHit(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.Query())
		expect(t, "keith", r.URL.Query().Get(":name"))
	}))

	r, err := http.NewRequest("GET", "/foo/keith?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	expect(t, ok, true)
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

	p.Patch("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	r, err := http.NewRequest("GET", "/foo/keith", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, r)

	expect(t, ok, false)
	expect(t, http.StatusMethodNotAllowed, rr.Code)

	allowed := strings.Split(rr.Header().Get("Allow"), ", ")
	sort.Strings(allowed)
	deepExpect(t, allowed, []string{"PATCH", "POST", "PUT"})
}

// Check to make sure we don't pollute the Raw Query when we have no parameters
func TestPatNoParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		expect(t, "", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	expect(t, ok, true)
}

// Check to make sure we don't pollute the Raw Query when there are parameters but no pattern variables
func TestPatOnlyUserParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		expect(t, "a=b", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	expect(t, ok, true)
}

func TestPatImplicitRedirect(t *testing.T) {
	p := New()
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

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

	r, err = http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

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

func TestCustomNotFound(t *testing.T) {
	p := New()
	p.SetNotFoundHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299)
		w.Write([]byte("Custom message here"))
	})

	r, err := http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	p.ServeHTTP(res, r)

	expect(t, res.Code, 299)
	expect(t, res.Body.String(), "Custom message here")

	// Handler version.
	p = New()
	p.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Custom message here"))
	}))

	r, err = http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	res = httptest.NewRecorder()
	p.ServeHTTP(res, r)

	expect(t, res.Code, 404)
	expect(t, res.Body.String(), "Custom message here")

	// Normal not found.
	p = New()
	r, err = http.NewRequest("GET", "/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	res = httptest.NewRecorder()
	p.ServeHTTP(res, r)

	expect(t, res.Code, 404)
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func deepExpect(t *testing.T, a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Expected [%v] (type %v) - Got [%v] (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
