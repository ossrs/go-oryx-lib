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

package kxps

import (
	"testing"
	"time"
)

type mockSource struct {
	s uint64
}

func (v *mockSource) NbRequests() uint64 {
	return v.s
}

func TestKrps_Average(t *testing.T) {
	s := &mockSource{}
	krps := NewKrps(nil, s).(*krps)

	if v := krps.sampleAverage(time.Unix(0, 0)); v != 0 {
		t.Errorf("invalid average %v", v)
	}

	s.s = 10
	if v := krps.sampleAverage(time.Unix(10, 0)); v != 0 {
		t.Errorf("invalid average %v", v)
	}

	s.s = 20
	if v := krps.sampleAverage(time.Unix(10, 0)); v != 0 {
		t.Errorf("invalid average %v", v)
	} else if v := krps.sampleAverage(time.Unix(20, 0)); v != 10.0/10.0 {
		t.Errorf("invalid average %v", v)
	}
}

func TestKrps_Rps10s(t *testing.T) {
	s := &mockSource{}
	krps := NewKrps(nil, s).(*krps)

	if err := krps.doSample(time.Unix(0, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	}

	s.s = 10
	if err := krps.doSample(time.Unix(10, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	}

	s.s = 20
	if err := krps.doSample(time.Unix(20, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 10.0/10.0 || krps.r30s.rps != 0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	} else if err := krps.doSample(time.Unix(30, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	}

	s.s = 30
	if err := krps.doSample(time.Unix(40, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 10.0/10.0 || krps.r30s.rps != 20.0/30.0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	} else if err := krps.doSample(time.Unix(50, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 20.0/30.0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	}

	s.s = 40
	if err := krps.doSample(time.Unix(310, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 10.0/10.0 || krps.r30s.rps != 10.0/30.0 || krps.r300s.rps != 30.0/300.0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	} else if err := krps.doSample(time.Unix(320, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 10.0/30.0 || krps.r300s.rps != 30.0/300.0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	} else if err := krps.doSample(time.Unix(340, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 0 || krps.r300s.rps != 30.0/300.0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	} else if err := krps.doSample(time.Unix(610, 0)); err != nil {
		t.Errorf("sample failed, err is", err)
	} else if krps.r10s.rps != 0 || krps.r30s.rps != 0 || krps.r300s.rps != 0 {
		t.Errorf("sample invalid, 10s=%v, 30s=%v, 300s=%v", krps.r10s.rps, krps.r30s.rps, krps.r300s.rps)
	}
}
