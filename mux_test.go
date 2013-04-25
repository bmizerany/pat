package pat

import (
	"github.com/bmizerany/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
)

func TestPatMatch(t *testing.T) {
	params, ok := (&patHandler{"/", nil, false}).try("/")
	assert.Equal(t, true, ok)

	params, ok = (&patHandler{"/", nil, false}).try("/wrong_url")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name", nil, false}).try("/foo/bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz", nil, false}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/bar/", nil, false}).try("/foo/keith/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil, false}).try("/foo/keith/bar/")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"keith"}}, params)

	params, ok = (&patHandler{"/foo/:name/bar/", nil, false}).try("/foo/keith/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz", nil, false}).try("/foo/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil, false}).try("/foo/bar/baz")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil, false}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}, ":id": {"123"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:name", nil, false}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar", "123"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil, false}).try("/foo/bar.txt")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name", nil, false}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil, false}).try("/foo/val1:val2")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {"val1"}, ":b": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:a.", nil, false}).try("/foo/.")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}}, params)

	params, ok = (&patHandler{"/foo/:a:b", nil, false}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}, ":b": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/:a:b:c", nil, false}).try("/foo/:bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":a": {""}, ":b": {""}, ":c": {":bar"}}, params)

	params, ok = (&patHandler{"/foo/::name", nil, false}).try("/foo/val1:val2")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":": {"val1"}, ":name": {":val2"}}, params)

	params, ok = (&patHandler{"/foo/:name.txt", nil, false}).try("/foo/bar/baz.txt")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil, false}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil, false}).try("/foo/xbar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)
}

func TestPatRoutingHit(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.Query())
		assert.Equal(t, "keith", r.URL.Query().Get(":name"))
	}))

	r, err := http.NewRequest("GET", "/foo/keith?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
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

	r, err := http.NewRequest("GET", "/foo/keith", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, r)

	assert.T(t, !ok)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	allowed := strings.Split(rr.Header().Get("Allow"), ", ")
	sort.Strings(allowed)
	assert.Equal(t, allowed, []string{"POST", "PUT"})
}

// Check to make sure we don't pollute the Raw Query when we have no parameters
func TestPatNoParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		assert.Equal(t, "", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
}

// Check to make sure we don't pollute the Raw Query when there are parameters but no pattern variables
func TestPatOnlyUserParams(t *testing.T) {
	p := New()

	var ok bool
	p.Get("/foo/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
		t.Logf("%#v", r.URL.RawQuery)
		assert.Equal(t, "a=b", r.URL.RawQuery)
	}))

	r, err := http.NewRequest("GET", "/foo/?a=b", nil)
	if err != nil {
		t.Fatal(err)
	}

	p.ServeHTTP(nil, r)

	assert.T(t, ok)
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

func TestPatImplicitRedirectWithMatcher(t *testing.T) {
	p := New()
	p.Get("/foo/:bar/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Should redirect to /foo/fluffy/.
	r, err := http.NewRequest("GET", "/foo/fluffy", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()
	p.ServeHTTP(res, r)

	if res.Code != 301 {
		t.Errorf("expected Code 301, was %d", res.Code)
	}

	if loc := res.Header().Get("Location"); loc != "/foo/fluffy/" {
		t.Errorf("expected %q, got %q", "/foo/fluffy/", loc)
	}

	// Should redirect and maintain passed parameters.
	r, err = http.NewRequest("GET", "/foo/fluffy?extra=bits", nil)
	if err != nil {
		t.Fatal(err)
	}

	res = httptest.NewRecorder()
	p.ServeHTTP(res, r)

	if res.Code != 301 {
		t.Errorf("expected Code 301, was %d", res.Code)
	}

	if loc := res.Header().Get("Location"); loc != "/foo/fluffy/?extra=bits" {
		t.Errorf("expected %q, got %q", "/foo/fluffy/?extra=bits", loc)
	}

	// Test adding both handlers manually (no redirect).
	p = New()
	p.Get("/foo/:bar", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	p.Get("/foo/:bar/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	r, err = http.NewRequest("GET", "/foo/fluffly", nil)
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
	if g := Tail("/:a/", "/x/y/z"); g != "y/z" {
		t.Fatalf("want %q, got %q", "y/z", g)
	}

	if g := Tail("/:a/", "/x"); g != "" {
		t.Fatalf("want %q, got %q", "", g)
	}

	if g := Tail("/:a/", "/x/"); g != "" {
		t.Fatalf("want %q, got %q", "", g)
	}

	if g := Tail("/:a", "/x/y/z"); g != "" {
		t.Fatalf("want: %q, got %q", "", g)
	}

	if g := Tail("/b/:a", "/x/y/z"); g != "" {
		t.Fatalf("want: %q, got %q", "", g)
	}
}

func TestBuildRedirectUrl(t *testing.T) {
	r, _ := http.NewRequest("GET", "/foo", nil)
	if buildUrlWithSlash(r) != "/foo/" {
		t.Fatalf("want %q, got %q", "/foo/", buildUrlWithSlash(r))
	}

	r, _ = http.NewRequest("GET", "/foo#bar", nil)
	if buildUrlWithSlash(r) != "/foo/#bar" {
		t.Fatalf("want %q, got %q", "/foo/#bar", buildUrlWithSlash(r))
	}

	r, _ = http.NewRequest("GET", "/foo?bar=wow", nil)
	if buildUrlWithSlash(r) != "/foo/?bar=wow" {
		t.Fatalf("want %q, got %q", "/foo/?bar=wow", buildUrlWithSlash(r))
	}

	r, _ = http.NewRequest("GET", "/foo?bar=wow#hello", nil)
	if buildUrlWithSlash(r) != "/foo/?bar=wow#hello" {
		t.Fatalf("want %q, got %q", "/foo/?bar=wow#hello", buildUrlWithSlash(r))
	}
}
