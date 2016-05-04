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

// The oryx http package provides standard request and response in json.
//			Error, when error, use this handler.
//			Data, when no error, use this handler.
//			SystemError, application level error code.
// The standard server response:
//			code, an int error code.
//			data, specifies the data.
// The global variable Server is the header["Server"].
package http

import (
	"encoding/json"
	"fmt"
	ologger "github.com/ossrs/go-oryx-lib/logger"
	"net/http"
)

// header["Content-Type"] in response.
const (
	HttpJson = "application/json"
	HttpJavaScript = "application/javascript"
)


// header["Server"] in response.
var Server = "Oryx"

// system int error.
type SystemError int

func (v SystemError) Error() string {
	return fmt.Sprintf("System error=%d", int(v))
}

// system conplex error.
type SystemComplexError struct {
	// the system error code.
	Code SystemError
	// the description for this error.
	Message string
}

func (v SystemComplexError) Error() string {
	return fmt.Sprintf("%v, %v", v.Code.Error(), v.Message)
}

// http standard error response.
func Error(ctx ologger.Context, err error) http.Handler {
	// for complex error, use code instead.
	if v, ok := err.(SystemComplexError); ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ologger.Error.Println(ctx, "Serve", r.URL, "failed. err is", err.Error())
			jsonHandler(ctx, map[string]interface{}{"code": v.Code, "data": v.Message}).ServeHTTP(w, r)
		})
	}

	// for int error, use code instead.
	if v, ok := err.(SystemError); ok {
		return jsonHandler(ctx, map[string]int{"code": int(v)})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		w.Header().Set("Content-Type", HttpJson)

		// unknown error, log and response detail
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ologger.Error.Println(ctx, "Serve", r.URL, "failed. err is", err.Error())
	})
}

// http normal response.
func Data(ctx ologger.Context, v interface{}) http.Handler {
	rv := map[string]interface{}{
		"code": 0,
		"data": v,
	}

	// for string, directly use it without convert,
	// for the type covert by golang maybe modify the content.
	if v, ok := v.(string); ok {
		rv["data"] = v
	}

	return jsonHandler(ctx, rv)
}

// set http header.
func SetHeader(w http.ResponseWriter) {
	w.Header().Set("Server", Server)
}

// response json directly.
func jsonHandler(ctx ologger.Context, rv interface{}) http.Handler {
	var err error
	var b []byte
	if b, err = json.Marshal(rv); err != nil {
		return Error(ctx, err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)

		q := r.URL.Query()
		if cb := q.Get("callback"); cb != "" {
			w.Header().Set("Content-Type", HttpJavaScript)
			fmt.Fprintf(w, "%s(%s)", cb, string(b))
		} else {
			w.Header().Set("Content-Type", HttpJson)
			w.Write(b)
		}
	})
}
