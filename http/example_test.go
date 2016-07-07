// The MIT License (MIT)
//
// Copyright (c) 2013-2016 Oryx(ossrs)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http_test

import (
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"fmt"
	"net/http"
)

func ExampleHttpTest_Global() {
	ohttp.Server = "Test"
	fmt.Println("Server:", ohttp.Server)

	// Output:
	// Server: Test
}

func ExampleHttpTest_RawResponse() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		// Set the common response header when need to write RAW message.
		ohttp.SetHeader(w)

		// Write RAW message only, or should use the Error() or Data() functions.
		w.Write([]byte("Hello, World!"))
	})
}

func ExampleHttpTest_JsonData() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		// Response data which can be marshal to json.
		ohttp.Data(nil, map[string]interface{}{
			"version": "1.0",
			"count": 100,
		}).ServeHTTP(w, r)
	})
}

func ExampleHttpTest_Error() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		// Response unknown error with HTTP/500
		ohttp.Error(nil, fmt.Errorf("System error")).ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		// Response known error {code:xx}
		ohttp.Error(nil, ohttp.SystemError(100)).ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		// Response known complex error {code:xx,data:"xxx"}
		ohttp.Error(nil, ohttp.SystemComplexError{ohttp.SystemError(100), "Error description"}).ServeHTTP(w, r)
	})
}
