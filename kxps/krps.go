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

// The oryx kxps package provides some kxps, for example:
//	N kbps, N k bits per seconds
//	N krps, N k requests per seconds
// over some duration for instance 10s, 30s, 5m, average.
package kxps

import (
	"fmt"
	ol "github.com/ossrs/go-oryx-lib/logger"
	"io"
	"sync"
	"time"
)

// The source to stat the requests.
type KrpsSource interface {
	// Get total number of requests.
	NbRequests() uint64
}

// The object to calc the krps.
type Krps interface {
	// Start the krps sample goroutine.
	Start() (err error)

	// Get the rps in last 10s.
	Rps10s() float64
	// Get the rps in last 30s.
	Rps30s() float64
	// Get the rps in last 300s.
	Rps300s() float64
	// Get the rps in average
	Average() float64

	// When closed, this krps should never use again.
	io.Closer
}

// sample for krps.
type sample struct {
	rps        float64
	nbRequests uint64
	create     time.Time
	lastSample time.Time
	// Duration in seconds.
	interval time.Duration
}

func (v *sample) initialize(now time.Time, nbRequests uint64) {
	v.nbRequests = nbRequests
	v.lastSample = now
	v.create = now
}

func (v *sample) sample(now time.Time, nbRequests uint64) bool {
	if v.lastSample.Add(v.interval).After(now) {
		return false
	}

	diff := int64(nbRequests - v.nbRequests)
	v.nbRequests = nbRequests
	v.lastSample = now
	if diff <= 0 {
		v.rps = 0
		return true
	}

	interval := int(v.interval / time.Millisecond)
	v.rps = float64(diff) * 1000 / float64(interval)

	return true
}

var krpsClosed = fmt.Errorf("krps closed")

// The implementation object.
type krps struct {
	// internal objects.
	source  KrpsSource
	ctx     ol.Context
	closed  bool
	started bool
	lock    *sync.Mutex
	// samples
	r10s  sample
	r30s  sample
	r300s sample
	// for average
	average uint64
	create  time.Time
}

func NewKrps(ctx ol.Context, s KrpsSource) Krps {
	v := &krps{
		lock:   &sync.Mutex{},
		source: s,
		ctx:    ctx,
	}

	v.r10s.interval = time.Duration(10) * time.Second
	v.r30s.interval = time.Duration(30) * time.Second
	v.r300s.interval = time.Duration(300) * time.Second

	return v
}

func (v *krps) Close() (err error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.closed = true
	v.started = false
	return
}

func (v *krps) Rps10s() float64 {
	if !v.started {
		panic("should start krps first.")
	}

	return v.r10s.rps
}

func (v *krps) Rps30s() float64 {
	if !v.started {
		panic("should start krps first.")
	}

	return v.r30s.rps
}

func (v *krps) Rps300s() float64 {
	if !v.started {
		panic("should start krps first.")
	}

	return v.r300s.rps
}

func (v *krps) Average() float64 {
	if !v.started {
		panic("should start krps first.")
	}

	return v.sampleAverage(time.Now())
}

func (v *krps) sampleAverage(now time.Time) float64 {
	if v.source.NbRequests() == 0 {
		return 0
	}

	if v.average == 0 {
		v.average = v.source.NbRequests()
		v.create = now
		return 0
	}

	diff := int64(v.source.NbRequests() - v.average)
	if diff <= 0 {
		return 0
	}

	duration := int64(now.Sub(v.create) / time.Millisecond)
	if duration <= 0 {
		return 0
	}

	return float64(diff) * 1000 / float64(duration)
}

func (v *krps) doSample(now time.Time) (err error) {
	nbRequests := v.source.NbRequests()
	if nbRequests == 0 {
		return
	}

	if v.r10s.nbRequests == 0 {
		v.r10s.initialize(now, nbRequests)
		v.r30s.initialize(now, nbRequests)
		v.r300s.initialize(now, nbRequests)
		return
	}

	if !v.r10s.sample(now, nbRequests) {
		return
	}

	if !v.r30s.sample(now, nbRequests) {
		return
	}

	if !v.r300s.sample(now, nbRequests) {
		return
	}

	return
}

func (v *krps) Start() (err error) {
	ctx := v.ctx

	go func() {
		if err := v.sample(); err != nil {
			if err == krpsClosed {
				return
			}
			ol.W(ctx, "krps ignore sample failed, err is", err)
		}
		time.Sleep(time.Duration(10) * time.Second)
	}()

	v.started = true

	return
}

func (v *krps) sample() (err error) {
	ctx := v.ctx

	defer func() {
		if r := recover(); r != nil {
			ol.W(ctx, "recover kxps from", r)
		}
	}()

	v.lock.Lock()
	defer v.lock.Unlock()

	if v.closed {
		return krpsClosed
	}

	return v.doSample(time.Now())
}
