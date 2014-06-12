# Pat
 A Sinatra style pattern muxer for Go's net/http library.

This was originally created by [Blake Mizerany](https://github.com/bmizerany). This fork is just some improvements/tweaks to some issues that I've encountered.

## Installation

	$ go get github.com/unrolled/pat

## Usage

~~~ go
// main.go
package main

import (
  "net/http"

  "github.com/unrolled/pat"
)

func main() {
  mux := pat.New()
  mux.GetFunc("/hello/:name", func(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello, " + req.URL.Query().Get(":name") + "!\n"))
  })

  // Register this pat with the default serve mux so that other packages
  // may also be exported. (i.e. /debug/pprof/*)
  http.Handle("/", mux)
  http.ListenAndServe(":3000", nil)
}
~~~

It's that simple.

### 404 - Not Found

You can also define a custom 404 handler:

~~~ go
// main.go
package main

import (
  "net/http"

  "github.com/unrolled/pat"
)

func main() {
  mux := pat.New()
  mux.GetFunc("/hello/:name", func(w http.ResponseWriter, req *http.Request) {
    w.Write([]byte("Hello, " + req.URL.Query().Get(":name") + "!\n"))
  })

  mux.SetNotFoundHandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("Oops! You're lost!"))
  })

  http.Handle("/", mux)
  http.ListenAndServe(":3000", nil)
}
~~~

For more information, see:
http://godoc.org/github.com/unrolled/pat

## Additions for this fork

- Added PATCH method.
- Removed dependecy on bmizerany/assert for testing.
- Added ability to set custom not found handler.

## Contributors

* Keith Rarick (@krarick) - github.com/kr
* Blake Mizerany (@bmizerany) - github.com/bmizerany
* Evan Shaw
* George Rogers
* Cory Jacobsen (@coryjacobsen) - github.com/unrolled

## License

Copyright (C) 2012 by Keith Rarick, Blake Mizerany

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
