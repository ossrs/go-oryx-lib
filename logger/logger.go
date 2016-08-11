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

// The oryx logger package provides connection-oriented log service.
//		logger.Info.Println(Context, ...)
//		logger.Trace.Println(Context, ...)
//		logger.Warn.Println(Context, ...)
//		logger.Error.Println(Context, ...)
// @remark the Context is optional thus can be nil.
package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// default level for logger.
const (
	logInfoLabel  = "[info] "
	logTraceLabel = "[trace] "
	logWarnLabel  = "[warn] "
	logErrorLabel = "[error] "
)

// the context for current goroutine.
type Context interface {
	// get current goroutine cid.
	Cid() int
}

// the LOG+ which provides connection-based log.
type loggerPlus struct {
	logger *log.Logger
}

func NewLoggerPlus(l *log.Logger) Logger {
	return &loggerPlus{logger: l}
}

func (v *loggerPlus) Println(ctx Context, a ...interface{}) {
	if ctx == nil {
		a = append([]interface{}{fmt.Sprintf("[%v]", os.Getpid())}, a...)
	} else {
		a = append([]interface{}{fmt.Sprintf("[%v][%v]", os.Getpid(), ctx.Cid())}, a...)
	}
	v.logger.Println(a...)
}

// Info, the verbose info level, very detail log, the lowest level, to discard.
var Info Logger

// Alias for Info level println.
func I(ctx Context, a ...interface{}) {
	Info.Println(ctx, a...)
}

// Trace, the trace level, something important, the default log level, to stdout.
var Trace Logger

// Alias for Trace level println.
func T(ctx Context, a ...interface{}) {
	Trace.Println(ctx, a...)
}

// Warn, the warning level, dangerous information, to stderr.
var Warn Logger

// Alias for Warn level println.
func W(ctx Context, a ...interface{}) {
	Warn.Println(ctx, a...)
}

// Error, the error level, fatal error things, ot stderr.
var Error Logger

// Alias for Error level println.
func E(ctx Context, a ...interface{}) {
	Error.Println(ctx, a...)
}

// The logger for oryx.
type Logger interface {
	// Println for logger plus,
	// @param ctx the connection-oriented context, or nil to ignore.
	Println(ctx Context, a ...interface{})
}

func init() {
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(os.Stdout, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(os.Stderr, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(os.Stderr, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
}

// Switch the underlayer io.
// @remark user must close previous io for logger never close it.
func Switch(w io.Writer) {
	// TODO: support level, default to trace here.
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(w, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(w, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(w, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))

	if w, ok := w.(io.Closer); ok {
		previousIo = w
	}
}

// The previous underlayer io for logger.
var previousIo io.Closer

// The interface io.Closer
// Cleanup the logger, discard any log util switch to fresh writer.
func Close() (err error) {
	Info = NewLoggerPlus(log.New(ioutil.Discard, logInfoLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Trace = NewLoggerPlus(log.New(ioutil.Discard, logTraceLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Warn = NewLoggerPlus(log.New(ioutil.Discard, logWarnLabel, log.Ldate|log.Ltime|log.Lmicroseconds))
	Error = NewLoggerPlus(log.New(ioutil.Discard, logErrorLabel, log.Ldate|log.Ltime|log.Lmicroseconds))

	if previousIo != nil {
		err = previousIo.Close()
		previousIo = nil
	}

	return
}
