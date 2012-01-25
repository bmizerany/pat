package main

import (
	"github.com/bmizerany/pat.go"
	"net/http"
	"io"
)

func main() {
	m := pat.NewPatternServeMux()
	m.Get("/hello/:name", http.HandlerFunc(hello))
	http.ListenAndServe("localhost:5000", m)
}

func hello(w http.ResponseWriter, r *http.Request) {
	// Path variable names are in the URL.Query() and start with ':'.
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
}
