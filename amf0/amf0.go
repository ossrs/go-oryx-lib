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

// The oryx amf0 package support AMF0 codec.
package amf0

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"math"
)

// Please read @doc amf0_spec_121207.pdf, @page 4, @section 2.1 Types Overview
type marker uint8

const (
	markerNumber marker = iota
	markerBoolean
	markerString
	markerObject
	markerMovieClip
	markerNull
	markerUndefined
	markerReference
	markerEcmaArray
	markerObjectEnd
	markerStrictArray
	markerDate
	markerLongString
	markerUnsupported
	markerRecordSet
	markerXmlDocument
	markerTypedObject
	markerAvmPlusObject

	markerForbidden marker = 0xff
)

func (v marker) String() string {
	switch v {
	case markerNumber:
		return "Number"
	case markerBoolean:
		return "Boolean"
	case markerString:
		return "String"
	case markerObject:
		return "Object"
	case markerNull:
		return "Null"
	case markerUndefined:
		return "Undefined"
	case markerReference:
		return "Reference"
	case markerEcmaArray:
		return "EcmaArray"
	case markerObjectEnd:
		return "ObjectEnd"
	case markerStrictArray:
		return "StrictArray"
	case markerDate:
		return "Date"
	case markerLongString:
		return "LongString"
	case markerUnsupported:
		return "Unsupported"
	case markerXmlDocument:
		return "XmlDocument"
	case markerTypedObject:
		return "TypedObject"
	case markerAvmPlusObject:
		return "AvmPlusObject"
	case markerForbidden, markerMovieClip, markerRecordSet:
		fallthrough
	default:
		return "Forbidden"
	}
}

var errDataNotEnough = errors.New("data is not enough")
var errMarkerIllegal = errors.New("marker is illegal")
var errNotSupported = errors.New("object is not supported")

// All AMF0 things.
type Amf0 interface {
	// Binary marshaler and unmarshaler.
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler
	// Get the size of bytes to marshal this object.
	Size() int

	// Get the Marker of any AMF0 stuff.
	amf0Marker() marker
}

// Discovery the amf0 object from the bytes b.
func Discovery(p []byte) (a Amf0, err error) {
	if len(p) < 1 {
		return nil, errDataNotEnough
	}
	m := marker(p[0])

	switch m {
	case markerNumber:
		return NewNumber(0), errNotSupported
	case markerBoolean:
		return nil, errNotSupported
	case markerString:
		return NewString(""), nil
	case markerObject:
		return NewObject(), nil
	case markerNull:
		return nil, errNotSupported
	case markerUndefined:
		return nil, errNotSupported
	case markerReference:
		return nil, errNotSupported
	case markerEcmaArray:
		return nil, errNotSupported
	case markerObjectEnd:
		return nil, errNotSupported
	case markerStrictArray:
		return nil, errNotSupported
	case markerDate:
		return nil, errNotSupported
	case markerLongString:
		return nil, errNotSupported
	case markerUnsupported:
		return nil, errNotSupported
	case markerXmlDocument:
		return nil, errNotSupported
	case markerTypedObject:
		return nil, errNotSupported
	case markerAvmPlusObject:
		return nil, errNotSupported
	case markerForbidden, markerMovieClip, markerRecordSet:
		fallthrough
	default:
		return nil, errMarkerIllegal
	}

	return
}

// The UTF8 string, please read @doc amf0_spec_121207.pdf, @page 3, @section 1.3.1 Strings and UTF-8
type amf0UTF8 string

func (v *amf0UTF8) Size() int {
	return 2 + len(string(*v))
}

func (v *amf0UTF8) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 2 {
		return errDataNotEnough
	}
	size := uint16(p[0])<<8 | uint16(p[1])

	if p = data[2:]; len(p) < int(size) {
		return errDataNotEnough
	}
	*v = amf0UTF8(string(p[:size]))

	return
}

func (v *amf0UTF8) MarshalBinary() (data []byte, err error) {
	data = make([]byte, v.Size())

	size := uint16(len(string(*v)))
	data[0] = byte(size >> 8)
	data[1] = byte(size)

	if size > 0 {
		copy(data[2:], []byte(*v))
	}

	return
}

// The number object, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.2 Number Type
type Number float64

func NewNumber(f float64) *Number {
	v := Number(f)
	return &v
}

func (v *Number) amf0Marker() marker {
	return markerNumber
}

func (v *Number) Size() int {
	return 1 + 8
}

func (v *Number) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 9 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerNumber {
		return errMarkerIllegal
	}

	f := binary.BigEndian.Uint64(p[1:])
	*v = Number(math.Float64frombits(f))
	return
}

func (v *Number) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 9)
	data[0] = byte(markerNumber)
	f := math.Float64bits(float64(*v))
	binary.BigEndian.PutUint64(data[1:], f)
	return
}

// The string objet, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.4 String Type
type String string

func NewString(s string) *String {
	v := String(s)
	return &v
}

func (v *String) amf0Marker() marker {
	return markerString
}

func (v *String) Size() int {
	u := amf0UTF8(*v)
	return 1 + u.Size()
}

func (v *String) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 1 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerString {
		return errMarkerIllegal
	}

	var sv amf0UTF8
	if err = sv.UnmarshalBinary(p[1:]); err != nil {
		return
	}
	*v = String(string(sv))
	return
}

func (v *String) MarshalBinary() (data []byte, err error) {
	u := amf0UTF8(*v)

	var pb []byte
	if pb, err = u.MarshalBinary(); err != nil {
		return
	}

	data = append([]byte{byte(markerString)}, pb...)
	return
}

// The AMF0 object end type, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.11 Object End Type
type objectEOF struct {
}

func (v *objectEOF) amf0Marker() marker {
	return markerObjectEnd
}

func (v *objectEOF) Size() int {
	return 3
}

func (v *objectEOF) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p[0] != 0 || p[1] != 0 || p[2] != 9 {
		return errMarkerIllegal
	}
	return
}

func (v *objectEOF) MarshalBinary() (data []byte, err error) {
	return []byte{0, 0, 9}, nil
}

// The AMF0 object, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.5 Object Type
type Object struct {
	m          marker
	properties map[amf0UTF8]Amf0
	eof        objectEOF
}

func NewObject() *Object {
	return &Object{
		m:          markerObject,
		properties: map[amf0UTF8]Amf0{},
	}
}

func (v *Object) Set(key string, value Amf0) {
	v.properties[amf0UTF8(key)] = value
}

func (v *Object) amf0Marker() marker {
	return v.m
}

func (v *Object) Size() int {
	size := int(1) + v.eof.Size()
	for key, value := range v.properties {
		size += key.Size() + value.Size()
	}
	return size
}

func (v *Object) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 1 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerObject {
		return errMarkerIllegal
	}

	for len(p) > 0 {
		var u amf0UTF8
		if err = u.UnmarshalBinary(p); err != nil {
			return
		}
		p = p[u.Size():]

		var a Amf0
		if a, err = Discovery(p); err != nil {
			return
		}

		// For object EOF, we should only consume total 3bytes.
		if u.Size() == 0 && a.amf0Marker() == markerObjectEnd {
			p = p[1:]
			break
		}

		// For object property, consume the whole bytes.
		if err = a.UnmarshalBinary(p); err != nil {
			return
		}
		v.properties[u] = a
		p = p[a.Size():]
	}
	return
}

func (v *Object) MarshalBinary() (data []byte, err error) {
	b := bytes.Buffer{}

	if err = b.WriteByte(byte(markerObject)); err != nil {
		return
	}

	var pb []byte
	for key, value := range v.properties {
		if pb, err = key.MarshalBinary(); err != nil {
			return
		}
		if _, err = b.Write(pb); err != nil {
			return
		}

		if pb, err = value.MarshalBinary(); err != nil {
			return
		}
		if _, err = b.Write(pb); err != nil {
			return
		}
	}

	if pb, err = v.eof.MarshalBinary(); err != nil {
		return
	}
	if _, err = b.Write(pb); err != nil {
		return
	}

	return b.Bytes(), nil
}
