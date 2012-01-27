package pat

import (
	"testing"
	"net/http"
)

func BenchmarkPatternMatching(b *testing.B) {
	p := New()
	p.Get("/hello/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){}))
	r, err := http.NewRequest("GET", "/hello/blake", nil)
	if err != nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		p.ServeHTTP(nil, r)
	}
}
