package pat

import (
	"github.com/bmizerany/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPatMatch(t *testing.T) {
	params, ok := (&patHandler{"/foo/:name", nil}).try("/foo/bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz", nil}).try("/foo/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/:name/baz/:id", nil}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}, ":id": {"123"}}, params)

	params, ok = (&patHandler{"/foo/:name/baz/:name", nil}).try("/foo/bar/baz/123")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar", "123"}}, params)

	params, ok = (&patHandler{"/foo/::name", nil}).try("/foo/bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{"::name": {"bar"}}, params)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/bar")
	assert.Equal(t, false, ok)

	params, ok = (&patHandler{"/foo/x:name", nil}).try("/foo/xbar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":name": {"bar"}}, params)

<<<<<<< HEAD
	params, ok = (&patHandler{"/foo/", nil}).try("/foo/bar/baz")
	assert.Equal(t, true, ok)
=======
	params, ok = (&patHandler{"/foo/*", nil}).try("/foo/bar/baz")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":splat": {"bar/baz"}}, params)

	params, ok = (&patHandler{"/foo/*", nil}).try("/foo/bar")
	assert.Equal(t, true, ok)
	assert.Equal(t, url.Values{":splat": {"bar"}}, params)
>>>>>>> ce97829f22552d81ed5cc297590f43d4481a658b
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

func TestPatRoutingNoHit(t *testing.T) {
	p := New()

	var ok bool
	p.Post("/foo/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	r, err := http.NewRequest("GET", "/foo/keith", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, r)

	assert.T(t, !ok)
	assert.Equal(t, http.StatusNotFound, rr.Code)
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
