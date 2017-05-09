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
	"fmt"
	"io"
	"math"
	"sync"
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
	case markerMovieClip:
		return "MovieClip"
	case markerRecordSet:
		return "RecordSet"
	default:
		return "Forbidden"
	}
}

var errDataNotEnough = errors.New("data is not enough")

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
		return NewNumber(0), nil
	case markerBoolean:
		return NewBoolean(false), nil
	case markerString:
		return NewString(""), nil
	case markerObject:
		return NewObject(), nil
	case markerNull:
		return NewNull(), nil
	case markerUndefined:
		return NewUndefined(), nil
	case markerReference:
	case markerEcmaArray:
		return NewEcmaArray(), nil
	case markerObjectEnd:
		return &objectEOF{}, nil
	case markerStrictArray:
		return NewStrictArray(), nil
	case markerDate:
	case markerLongString:
	case markerUnsupported:
	case markerXmlDocument:
	case markerTypedObject:
	case markerAvmPlusObject:
	case markerForbidden, markerMovieClip, markerRecordSet:
		fallthrough
	default:
		return nil, fmt.Errorf("Marker %v is illegal", m)
	}

	return nil, fmt.Errorf("%v is not supported", m)
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
		return fmt.Errorf("Number marker %v is illegal", m)
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
		return fmt.Errorf("String marker %v is illegal", m)
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
		return fmt.Errorf("EOF marker %v is illegal", p[0:3])
	}
	return
}

func (v *objectEOF) MarshalBinary() (data []byte, err error) {
	return []byte{0, 0, 9}, nil
}

// Use array for object and ecma array, to keep the original order.
type property struct {
	key   amf0UTF8
	value Amf0
}

// The object-like AMF0 structure, like object and ecma array and strict array.
type objectBase struct {
	properties []*property
	lock       sync.Mutex
}

func (v *objectBase) Size() int {
	v.lock.Lock()
	defer v.lock.Unlock()

	var size int

	for _, p := range v.properties {
		key, value := p.key, p.value
		size += key.Size() + value.Size()
	}

	return size
}

func (v *objectBase) Get(key string) Amf0 {
	v.lock.Lock()
	defer v.lock.Unlock()

	for _, p := range v.properties {
		if string(p.key) == key {
			return p.value
		}
	}

	return nil
}

func (v *objectBase) Set(key string, value Amf0) {
	v.lock.Lock()
	defer v.lock.Unlock()

	prop := &property{key: amf0UTF8(key), value: value}

	var ok bool
	for i, p := range v.properties {
		if string(p.key) == key {
			v.properties[i] = prop
			ok = true
		}
	}

	if !ok {
		v.properties = append(v.properties, prop)
	}
}

func (v *objectBase) unmarshal(p []byte, eof bool, maxElems int) (err error) {
	for len(p) > 0 {
		var u amf0UTF8
		if err = u.UnmarshalBinary(p); err != nil {
			return fmt.Errorf("Unmarhsal prop name, %v", err)
		}
		p = p[u.Size():]

		var a Amf0
		if a, err = Discovery(p); err != nil {
			return fmt.Errorf("Discover prop %v, %v", u, err)
		}

		// For object EOF, we should only consume total 3bytes.
		if eof && u.Size() == 2 && a.amf0Marker() == markerObjectEnd {
			p = p[1:]
			break
		}

		// For object property, consume the whole bytes.
		if err = a.UnmarshalBinary(p); err != nil {
			return fmt.Errorf("Unmarshal prop %v, %v", u, err)
		}

		v.Set(string(u), a)
		p = p[a.Size():]

		if maxElems > 0 && len(v.properties) >= maxElems {
			break
		}
	}

	return
}

func (v *objectBase) marshal(b io.Writer) (err error) {
	v.lock.Lock()
	defer v.lock.Unlock()

	var pb []byte
	for _, p := range v.properties {
		key, value := p.key, p.value

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

	return
}

// The AMF0 object, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.5 Object Type
type Object struct {
	objectBase
	eof objectEOF
}

func NewObject() *Object {
	v := &Object{}
	v.properties = []*property{}
	return v
}

func (v *Object) amf0Marker() marker {
	return markerObject
}

func (v *Object) Size() int {
	return int(1) + v.eof.Size() + v.objectBase.Size()
}

func (v *Object) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 1 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerObject {
		return fmt.Errorf("Object marker %v is illegal", m)
	}
	p = p[1:]

	if err = v.unmarshal(p, true, -1); err != nil {
		return fmt.Errorf("Object %v", err)
	}

	return
}

func (v *Object) MarshalBinary() (data []byte, err error) {
	b := bytes.Buffer{}

	if err = b.WriteByte(byte(markerObject)); err != nil {
		return
	}

	if err = v.marshal(&b); err != nil {
		return nil, fmt.Errorf("Object %v", err)
	}

	var pb []byte
	if pb, err = v.eof.MarshalBinary(); err != nil {
		return
	}
	if _, err = b.Write(pb); err != nil {
		return
	}

	return b.Bytes(), nil
}

// The AMF0 ecma array, please read @doc amf0_spec_121207.pdf, @page 6, @section 2.10 ECMA Array Type
type EcmaArray struct {
	objectBase
	count uint32
	eof   objectEOF
}

func NewEcmaArray() *EcmaArray {
	v := &EcmaArray{}
	v.properties = []*property{}
	return v
}

func (v *EcmaArray) amf0Marker() marker {
	return markerEcmaArray
}

func (v *EcmaArray) Size() int {
	return int(1) + 4 + v.eof.Size() + v.objectBase.Size()
}

func (v *EcmaArray) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 5 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerEcmaArray {
		return fmt.Errorf("EcmaArray marker %v is illegal", m)
	}
	v.count = binary.BigEndian.Uint32(p[1:])
	p = p[5:]

	if err = v.unmarshal(p, true, -1); err != nil {
		return fmt.Errorf("EcmaArray %v", err)
	}
	return
}

func (v *EcmaArray) MarshalBinary() (data []byte, err error) {
	b := bytes.Buffer{}

	if err = b.WriteByte(byte(markerEcmaArray)); err != nil {
		return
	}

	if err = binary.Write(&b, binary.BigEndian, v.count); err != nil {
		return
	}

	if err = v.marshal(&b); err != nil {
		return nil, fmt.Errorf("EcmaArray %v", err)
	}

	var pb []byte
	if pb, err = v.eof.MarshalBinary(); err != nil {
		return
	}
	if _, err = b.Write(pb); err != nil {
		return
	}

	return b.Bytes(), nil
}

// The AMF0 strict array, please read @doc amf0_spec_121207.pdf, @page 7, @section 2.12 Strict Array Type
type StrictArray struct {
	objectBase
	count uint32
}

func NewStrictArray() *StrictArray {
	v := &StrictArray{}
	v.properties = []*property{}
	return v
}

func (v *StrictArray) amf0Marker() marker {
	return markerStrictArray
}

func (v *StrictArray) Size() int {
	return int(1) + 4 + v.objectBase.Size()
}

func (v *StrictArray) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 5 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerStrictArray {
		return fmt.Errorf("StrictArray marker %v is illegal", m)
	}
	v.count = binary.BigEndian.Uint32(p[1:])
	p = p[5:]

	if err = v.unmarshal(p, false, int(v.count)); err != nil {
		return fmt.Errorf("StrictArray %v", err)
	}
	return
}

func (v *StrictArray) MarshalBinary() (data []byte, err error) {
	b := bytes.Buffer{}

	if err = b.WriteByte(byte(markerStrictArray)); err != nil {
		return
	}

	if err = binary.Write(&b, binary.BigEndian, v.count); err != nil {
		return
	}

	if err = v.marshal(&b); err != nil {
		return nil, fmt.Errorf("StrictArray %v", err)
	}

	return b.Bytes(), nil
}

// The single marker object, for all AMF0 which only has the marker, like null and undefined.
type singleMarkerObject struct {
	target marker
}

func (v *singleMarkerObject) amf0Marker() marker {
	return v.target
}

func (v *singleMarkerObject) Size() int {
	return int(1)
}

func (v *singleMarkerObject) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 1 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != v.target {
		return fmt.Errorf("%v marker %v is illegal", v.target, m)
	}
	return
}

func (v *singleMarkerObject) MarshalBinary() (data []byte, err error) {
	return []byte{byte(v.target)}, nil
}

// The AMF0 null, please read @doc amf0_spec_121207.pdf, @page 6, @section 2.7 null Type
type null struct {
	singleMarkerObject
}

func NewNull() Amf0 {
	v := null{}
	v.singleMarkerObject.target = markerNull
	return &v
}

// The AMF0 undefined, please read @doc amf0_spec_121207.pdf, @page 6, @section 2.8 undefined Type
type undefined struct {
	singleMarkerObject
}

func NewUndefined() Amf0 {
	v := undefined{}
	v.singleMarkerObject.target = markerUndefined
	return &v
}

// The AMF0 boolean, please read @doc amf0_spec_121207.pdf, @page 5, @section 2.3 Boolean Type
type Boolean bool

func NewBoolean(b bool) Amf0 {
	v := Boolean(b)
	return &v
}

func (v *Boolean) amf0Marker() marker {
	return markerBoolean
}

func (v *Boolean) Size() int {
	return int(2)
}

func (v *Boolean) UnmarshalBinary(data []byte) (err error) {
	var p []byte
	if p = data; len(p) < 2 {
		return errDataNotEnough
	}
	if m := marker(p[0]); m != markerBoolean {
		return fmt.Errorf("BOOL marker %v is illegal", m)
	}
	if p[1] == 0 {
		*v = false
	} else {
		*v = true
	}
	return
}

func (v *Boolean) MarshalBinary() (data []byte, err error) {
	var b byte
	if *v {
		b = 1
	}
	return []byte{byte(markerBoolean), b}, nil
}
