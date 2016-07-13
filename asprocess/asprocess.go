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

// The oryx asprocess package provides associated process, which fork by parent
// process and die when parent die, for example, BMS server use asprocess to
// transcode audio, resolve DNS, bridge protocol(redis, kafka e.g.), and so on.
package asprocess

import (
	ocore "github.com/ossrs/go-oryx-lib/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// The recomment interval to check the parent pid.
const CheckParentInterval = time.Second * 1

// The cleanup function.
type Cleanup func()

// Watch the parent process, quit when parent quit.
// @remark optional ctx the logger context. nil to ignore.
// @reamrk check interval, user can use const CheckParentInterval
// @remark optional callback cleanup callback function. nil to ignore.
func Watch(ctx ocore.Context, interval time.Duration, callback Cleanup) {
	v := &aspContext{
		ctx:      ctx,
		interval: interval,
		callback: callback,
	}

	v.InstallSignals()

	v.WatchParent()
}

type aspContext struct {
	ctx      ocore.Context
	interval time.Duration
	callback Cleanup
}

func (v *aspContext) InstallSignals() {
	// install signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for s := range sigs {
			ocore.Trace.Println(v.ctx, "go signal", s)

			if v.callback != nil {
				v.callback()
			}

			os.Exit(0)
		}
	}()
	ocore.Trace.Println(v.ctx, "signal watched")
}

func (v *aspContext) WatchParent() {
	ppid := os.Getppid()

	go func() {
		for {
			if pid := os.Getppid(); pid == 1 || pid != ppid {
				ocore.Error.Println(v.ctx, "quit for parent problem, ppid is", pid)

				if v.callback != nil {
					v.callback()
				}

				os.Exit(0)
			}
			//ocore.Trace.Println(v.ctx, "parent pid", ppid, "ok")

			time.Sleep(v.interval)
		}
	}()
	ocore.Trace.Println(v.ctx, "parent process watching, ppid is", ppid)
}
