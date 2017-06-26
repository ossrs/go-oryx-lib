// The MIT License (MIT)
//
// Copyright (c) 2013-2017 Oryx(ossrs)
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
//		logger.I(Context, ...)
//		logger.T(Context, ...)
//		logger.W(Context, ...)
//		logger.E(Context, ...)
// Or use format:
//		logger.If(Context, format, ...)
//		logger.Tf(Context, format, ...)
//		logger.Wf(Context, format, ...)
//		logger.Ef(Context, format, ...)
// @remark the Context is optional thus can be nil.
// @remark From 1.7+, the ctx could be context.Context from std library,
//	however, we must use logger.WithContext to wrap it, for example,
//		ctx,cancel := context.WithCancel(context.Background())
//		ctx := logger.WithContext(ctx)
//		ol.T(ctx, "log with context")
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

// The context for current goroutine.
// It maybe a cidContext or context.Context from GO1.7.
// @remark Use logger.WithContext(ctx) to wrap the context.
type Context interface{}

// The context to get current coroutine cid.
type cidContext interface {
	Cid() int
}

// the LOG+ which provides connection-based log.
type loggerPlus struct {
	logger *log.Logger
}

func NewLoggerPlus(l *log.Logger) Logger {
	return &loggerPlus{logger: l}
}

func (v *loggerPlus) format(ctx Context, a ...interface{}) []interface{} {
	if ctx == nil {
		return append([]interface{}{fmt.Sprintf("[%v]", os.Getpid())}, a...)
	} else if ctx, ok := ctx.(cidContext); ok {
		return append([]interface{}{fmt.Sprintf("[%v][%v]", os.Getpid(), ctx.Cid())}, a...)
	}
	return a
}

// Info, the verbose info level, very detail log, the lowest level, to discard.
var Info Logger

// Alias for Info level println.
func I(ctx Context, a ...interface{}) {
	Info.Println(ctx, a...)
}

// Printf for Info level log.
func If(ctx Context, format string, a ...interface{}) {
	Info.Printf(ctx, format, a...)
}

// Trace, the trace level, something important, the default log level, to stdout.
var Trace Logger

// Alias for Trace level println.
func T(ctx Context, a ...interface{}) {
	Trace.Println(ctx, a...)
}

// Printf for Trace level log.
func Tf(ctx Context, format string, a ...interface{}) {
	Trace.Printf(ctx, format, a...)
}

// Warn, the warning level, dangerous information, to stderr.
var Warn Logger

// Alias for Warn level println.
func W(ctx Context, a ...interface{}) {
	Warn.Println(ctx, a...)
}

// Printf for Warn level log.
func Wf(ctx Context, format string, a ...interface{}) {
	Warn.Printf(ctx, format, a...)
}

// Error, the error level, fatal error things, ot stderr.
var Error Logger

// Alias for Error level println.
func E(ctx Context, a ...interface{}) {
	Error.Println(ctx, a...)
}

// Printf for Error level log.
func Ef(ctx Context, format string, a ...interface{}) {
	Error.Printf(ctx, format, a...)
}

// The logger for oryx.
type Logger interface {
	// Println for logger plus,
	// @param ctx the connection-oriented context,
	// 	or context.Context from GO1.7, or nil to ignore.
	Println(ctx Context, a ...interface{})
	Printf(ctx Context, format string, a ...interface{})
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
