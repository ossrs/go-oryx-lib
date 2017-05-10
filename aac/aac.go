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

// The oryx AAC package includes some utilites.
package aac

import (
	"errors"
)

// The ADTS is a format of AAC.
// We can encode the RAW AAC frame in ADTS muxer.
// We can also decode the ADTS data to RAW AAC frame.
type ADTS interface {
	// Set the ASC, the codec information.
	// Before encoding raw frame, user must set the asc.
	SetASC(asc []byte) (err error)
	// Encode the raw aac frame to adts data.
	// @remark User must set the asc first.
	Encode(raw []byte) (adts []byte, err error)

	// Decode the adts data to raw frame.
	// @remark User can get the asc after decode ok.
	Decode(adts []byte) (raw []byte, err error)
	// Get the ASC, the codec information.
	// When decode a adts data or set the asc, user can use this API to get it.
	ASC() (asc []byte)
}

// The AAC object type in RAW AAC frame.
// Refer to @doc ISO_IEC_14496-3-AAC-2001.pdf, @page 23, @section 1.5.1.1 Audio object type definition
type ObjectType uint8

const (
	ObjectTypeForbidden ObjectType = iota

	ObjectTypeMain
	ObjectTypeLC
	ObjectTypeSSR

	ObjectTypeHE   ObjectType = 5  // HE=LC+SBR
	ObjectTypeHEv2 ObjectType = 29 // HEv2=LC+SBR+PS
)

func (v ObjectType) String() string {
	switch v {
	case ObjectTypeMain:
		return "Main"
	case ObjectTypeLC:
		return "LC"
	case ObjectTypeSSR:
		return "SSR"
	case ObjectTypeHE:
		return "HE"
	case ObjectTypeHEv2:
		return "HEv2"
	default:
		return "Forbidden"
	}
}

func (v ObjectType) ToProfile() Profile {
	switch v {
	case ObjectTypeMain:
		return ProfileMain
	case ObjectTypeHE, ObjectTypeHEv2, ObjectTypeLC:
		return ProfileLC
	case ObjectTypeSSR:
		return ProfileSSR
	default:
		return ProfileForbidden
	}
}

// The profile of AAC in ADTS.
// Refer to @doc ISO_IEC_13818-7-AAC-2004.pdf, @page 40, @section 7.1 Profiles
type Profile uint8

const (
	ProfileMain Profile = iota
	ProfileLC
	ProfileSSR
	ProfileForbidden
)

func (v Profile) String() string {
	switch v {
	case ProfileMain:
		return "Main"
	case ProfileLC:
		return "LC"
	case ProfileSSR:
		return "SSR"
	default:
		return "Forbidden"
	}
}

func (v Profile) ToObjectType() ObjectType {
	switch v {
	case ProfileMain:
		return ObjectTypeMain
	case ProfileLC:
		return ObjectTypeLC
	case ProfileSSR:
		return ObjectTypeSSR
	default:
		return ObjectTypeForbidden
	}
}

var errDataNotEnough = errors.New("Data not enough")

type adts struct {
	object     ObjectType // AAC object type.
	sampleRate uint8      // AAC sample rate, not the FLV sampling rate.
	channels   uint8      // AAC channel configuration.
}

func NewADTS() (ADTS, error) {
	return &adts{}, nil
}

func (v *adts) SetASC(asc []byte) (err error) {
	// AudioSpecificConfig
	// Refer to @doc ISO_IEC_14496-3-AAC-2001.pdf, @page 33, @section 1.6.2.1 AudioSpecificConfig
	//
	// only need to decode the first 2bytes:
	// audioObjectType, 5bits.
	// samplingFrequencyIndex, aac_sample_rate, 4bits.
	// channelConfiguration, aac_channels, 4bits
	//
	// @see SrsAacTransmuxer::write_audio
	if len(asc) < 2 {
		return errDataNotEnough
	}

	t0, t1 := uint8(asc[0]), uint8(asc[1])

	v.object = ObjectType((t0 >> 3) & 0x1f)
	v.sampleRate = ((t0 << 1) & 0x0e) | ((t1 >> 7) & 0x01)
	v.channels = (t1 >> 3) & 0x0f

	return
}

func (v *adts) Encode(raw []byte) (adts []byte, err error) {
	// write the ADTS header.
	// Refer to @doc ISO_IEC_14496-3-AAC-2001.pdf, @page 75, @section 1.A.2.2 Audio_Data_Transport_Stream frame, ADTS
	// @see https://github.com/ossrs/srs/issues/212#issuecomment-64145885
	// byte_alignment()

	// adts_fixed_header:
	//      12bits syncword,
	//      16bits left.
	// adts_variable_header:
	//      28bits
	//      12+16+28=56bits
	// adts_error_check:
	//      16bits if protection_absent
	//      56+16=72bits
	// if protection_absent:
	//      require(7bytes)=56bits
	// else
	//      require(9bytes)=72bits
	aacFixedHeader := make([]byte, 7)
	p := aacFixedHeader

	// Syncword 12 bslbf
	p[0] = byte(0xff)
	// 4bits left.
	// adts_fixed_header(), 1.A.2.2.1 Fixed Header of ADTS
	// ID 1 bslbf
	// Layer 2 uimsbf
	// protection_absent 1 bslbf
	p[1] = byte(0xf1)

	// profile 2 uimsbf
	// sampling_frequency_index 4 uimsbf
	// private_bit 1 bslbf
	// channel_configuration 3 uimsbf
	// original/copy 1 bslbf
	// home 1 bslbf
	profile := v.object.ToProfile()
	p[2] = byte((profile<<6)&0xc0) | byte((v.sampleRate<<2)&0x3c) | byte((v.channels>>2)&0x01)

	// 4bits left.
	// adts_variable_header(), 1.A.2.2.2 Variable Header of ADTS
	// copyright_identification_bit 1 bslbf
	// copyright_identification_start 1 bslbf
	aacFrameLength := uint16(len(raw) + len(aacFixedHeader))
	p[3] = byte((v.channels<<6)&0xc0) | byte((aacFrameLength>>11)&0x03)

	// aac_frame_length 13 bslbf: Length of the frame including headers and error_check in bytes.
	// use the left 2bits as the 13 and 12 bit,
	// the aac_frame_length is 13bits, so we move 13-2=11.
	p[4] = byte(aacFrameLength >> 3)
	// adts_buffer_fullness 11 bslbf
	p[5] = byte(aacFrameLength<<5) & byte(0xe0)

	// no_raw_data_blocks_in_frame 2 uimsbf
	p[6] = byte(0xfc)

	return append(p, raw...), nil
}

func (v *adts) Decode(adts []byte) (raw []byte, err error) {
	return
}

func (v *adts) ASC() (asc []byte) {
	return
}
