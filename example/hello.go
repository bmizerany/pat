package main

import (
	"github.com/bmizerany/pat.go"
	"io"
	"net/http"
)

func main() {
	m := pat.New()
	m.Get("/hello/:name", http.HandlerFunc(hello))
	m.Get("/splat/*", http.HandlerFunc(splat))
	http.ListenAndServe("localhost:5000", m)
}

func hello(w http.ResponseWriter, r *http.Request) {
	// Path variable names are in the URL.Query() and start with ':'.
	name := r.URL.Query().Get(":name")
	io.WriteString(w, "Hello, "+name)
}

func splat(w http.ResponseWriter, r *http.Request) {
	// Path variable names are in the URL.Query() and start with ':'.
	s := r.URL.Query().Get(":splat")
	io.WriteString(w, "Splat: "+s)
}
