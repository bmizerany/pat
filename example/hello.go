package main

import (
	"github.com/bmizerany/pat.go"
	"io"
	"net/http"
)

func main() {
	m := pat.New()
	m.Get("/hello/:name", http.HandlerFunc(hello))
	http.ListenAndServe("localhost:5000", m)
}

func hello(w http.ResponseWriter, r *http.Request) {
	// Path variable names are in the URL.Query() and start with ':'.
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
}
