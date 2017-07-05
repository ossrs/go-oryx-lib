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

package amf0

import "testing"

func TestAmf0Marker(t *testing.T) {
	pvs := []struct {
		m marker
		ms string
	}{
		{markerNumber, "Number"},
		{markerBoolean, "Boolean"},
		{markerString, "String"},
		{markerObject, "Object"},
		{markerNull, "Null"},
		{markerUndefined, "Undefined"},
		{markerReference, "Reference"},
		{markerEcmaArray, "EcmaArray"},
		{markerObjectEnd, "ObjectEnd"},
		{markerStrictArray, "StrictArray"},
		{markerDate, "Date"},
		{markerLongString, "LongString"},
		{markerUnsupported, "Unsupported"},
		{markerXmlDocument, "XmlDocument"},
		{markerTypedObject, "TypedObject"},
		{markerAvmPlusObject, "AvmPlusObject"},
		{markerMovieClip, "MovieClip"},
		{markerRecordSet, "RecordSet"},
	}
	for _,pv := range pvs {
		if v := pv.m.String(); v != pv.ms {
			t.Errorf("marker %v expect %v actual %v", pv.m, pv.ms, v)
		}
	}
}

func TestDiscovery(t *testing.T) {
	pvs := []struct{
		m marker
		mv byte
	}{
		{markerNumber, 0},
		{markerBoolean, 1},
		{markerString, 2},
		{markerObject, 3},
		{markerNull, 5},
		{markerUndefined, 6},
		{markerEcmaArray, 8},
		{markerObjectEnd, 9},
		{markerStrictArray, 10},
	}
	for _,pv := range pvs {
		if m,err := Discovery([]byte{pv.mv}); err != nil {
			t.Errorf("discovery err %+v", err)
		} else if v := m.amf0Marker(); v != pv.m {
			t.Errorf("invalid %v expect %v actual %v", pv.mv, pv.m, v)
		}
	}
}
