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

// The oryx FLV package support bytes from/to FLV tags.
package flv

import (
	"bytes"
	"errors"
	"io"
)

// FLV Tag Type is the type of tag,
// refer to @doc video_file_format_spec_v10.pdf, @page 9, @section FLV tags
type TagType uint8

const (
	TagTypeForbidden  TagType = 0
	TagTypeAudio      TagType = 8
	TagTypeVideo      TagType = 9
	TagTypeScriptData TagType = 18
)

func (v TagType) String() string {
	switch v {
	case TagTypeVideo:
		return "Video"
	case TagTypeAudio:
		return "Audio"
	case TagTypeScriptData:
		return "Data"
	default:
		return "Forbidden"
	}
}

// FLV Demuxer is used to demux FLV file.
// Refer to @doc video_file_format_spec_v10.pdf, @page 7, @section The FLV File Format
// A FLV file must consist the bellow parts:
//	1. A FLV header, refer to @doc video_file_format_spec_v10.pdf, @page 8, @section The FLV header
//	2. One or more tags, refer to @doc video_file_format_spec_v10.pdf, @page 9, @section FLV tags
// @remark We always ignore the previous tag size.
type Demuxer interface {
	// Read the FLV header, return the version of FLV, whether hasVideo or hasAudio in header.
	ReadHeader() (version uint8, hasVideo, hasAudio bool, err error)
	// Read the FLV tag header, return the tag information, especially the tag size,
	// then user can read the tag payload.
	ReadTagHeader() (tagType TagType, tagSize, timestamp uint32, err error)
	// Read the FLV tag body, drop the next 4 bytes previous tag size.
	ReadTag(tagSize uint32) (tag []byte, err error)
}

// When FLV signature is not "FLV"
var ErrSignature = errors.New("FLV signatures are illegal")

// Create a demuxer object.
func NewDemuxer(r io.Reader) Demuxer {
	return &demuxer{
		r: r,
	}
}

type demuxer struct {
	r io.Reader
}

func (v *demuxer) ReadHeader() (version uint8, hasVideo, hasAudio bool, err error) {
	h := &bytes.Buffer{}
	if _, err = io.CopyN(h, v.r, 13); err != nil {
		return
	}

	p := h.Bytes()

	if !bytes.Equal([]byte{byte('F'), byte('L'), byte('V')}, p[:3]) {
		err = ErrSignature
		return
	}

	version = uint8(p[3])
	hasVideo = (p[4] & 0x01) == 0x01
	hasAudio = ((p[4] >> 2) & 0x01) == 0x01

	return
}

func (v *demuxer) ReadTagHeader() (tagType TagType, tagSize uint32, timestamp uint32, err error) {
	h := &bytes.Buffer{}
	if _, err = io.CopyN(h, v.r, 11); err != nil {
		return
	}

	p := h.Bytes()

	tagType = TagType(p[0])
	tagSize = uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3])
	timestamp = uint32(p[7])<<24 | uint32(p[4])<<16 | uint32(p[5])<<8 | uint32(p[6])

	return
}

func (v *demuxer) ReadTag(tagSize uint32) (tag []byte, err error) {
	h := &bytes.Buffer{}
	if _, err = io.CopyN(h, v.r, int64(tagSize+4)); err != nil {
		return
	}

	p := h.Bytes()
	tag = p[0 : len(p)-4]

	return
}
