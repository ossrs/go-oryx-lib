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
	ol "github.com/ossrs/go-oryx-lib/logger"
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
	// Get the rps in last 10s.
	Rps10s() float64
	// Get the rps in last 30s.
	Rps30s() float64
	// Get the rps in last 300s.
	Rps300s() float64
	// Get the rps in average
	Average() float64

	// Sample the krps to calc the krps in different duration.
	Sample()
}

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

	diff := int(nbRequests - v.nbRequests)
	v.nbRequests = nbRequests
	if v.interval > 0 {
		v.lastSample = now
	}

	if diff < 0 {
		return true
	}

	interval := int(v.interval / time.Second)
	if interval == 0 {
		interval = int(now.Sub(v.create) / time.Second)
	}

	if interval > 0 {
		v.rps = float64(diff) / float64(interval)
	}

	return true
}

type krps struct {
	lock        *sync.Mutex
	initialized bool
	r10s        sample
	r30s        sample
	r300s       sample
	average     sample
	source      KrpsSource
	ctx         ol.Context
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

func (v *krps) Rps10s() float64 {
	return v.r10s.rps
}

func (v *krps) Rps30s() float64 {
	return v.r30s.rps
}

func (v *krps) Rps300s() float64 {
	return v.r300s.rps
}

func (v *krps) Average() float64 {
	return v.average.rps
}

func (v *krps) Sample() {
	ctx := v.ctx

	defer func() {
		if r := recover(); r != nil {
			ol.W(ctx, "recover kxps from", r)
		}
	}()

	v.lock.Lock()
	defer v.lock.Unlock()

	now := time.Now()
	nbRequests := v.source.NbRequests()

	if !v.initialized {
		v.initialized = true
		v.r10s.initialize(now, nbRequests)
		v.r30s.initialize(now, nbRequests)
		v.r300s.initialize(now, nbRequests)
		v.average.initialize(now, nbRequests)
		return
	}

	v.average.sample(now, nbRequests)

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
