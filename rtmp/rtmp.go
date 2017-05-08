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

// The oryx rtmp package support bytes from/to rtmp packets.
package rtmp

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"github.com/ossrs/go-oryx-lib/amf0"
	"io"
	"math/rand"
)

// The handshake implements the RTMP handshake protocol.
type Handshake struct {
	r *rand.Rand
}

func NewHandshake(r *rand.Rand) *Handshake {
	return &Handshake{r: r}
}

func (v *Handshake) WriteC0S0(w io.Writer) (err error) {
	r := bytes.NewReader([]byte{0x03})
	if _, err = io.Copy(w, r); err != nil {
		return
	}

	return
}

func (v *Handshake) ReadC0S0(r io.Reader) (c0 []byte, err error) {
	b := &bytes.Buffer{}
	if _, err = io.CopyN(b, r, 1); err != nil {
		return
	}

	c0 = b.Bytes()

	return
}

func (v *Handshake) WriteC1S1(w io.Writer) (err error) {
	p := make([]byte, 1536)

	if _, err = v.r.Read(p[8:]); err != nil {
		return
	}

	r := bytes.NewReader(p)
	if _, err = io.Copy(w, r); err != nil {
		return
	}

	return
}

func (v *Handshake) ReadC1S1(r io.Reader) (c1 []byte, err error) {
	b := &bytes.Buffer{}
	if _, err = io.CopyN(b, r, 1536); err != nil {
		return
	}

	c1 = b.Bytes()

	return
}

func (v *Handshake) WriteC2S2(w io.Writer, s1c1 []byte) (err error) {
	r := bytes.NewReader(s1c1[:])
	if _, err = io.Copy(w, r); err != nil {
		return
	}

	return
}

func (v *Handshake) ReadC2S2(r io.Reader) (c2 []byte, err error) {
	b := &bytes.Buffer{}
	if _, err = io.CopyN(b, r, 1536); err != nil {
		return
	}

	c2 = b.Bytes()

	return
}

// Please read @doc rtmp_specification_1.0.pdf, @page 16, @section 6.1. Chunk Format
// Extended timestamp: 0 or 4 bytes
// This field MUST be sent when the normal timsestamp is set to
// 0xffffff, it MUST NOT be sent if the normal timestamp is set to
// anything else. So for values less than 0xffffff the normal
// timestamp field SHOULD be used in which case the extended timestamp
// MUST NOT be present. For values greater than or equal to 0xffffff
// the normal timestamp field MUST NOT be used and MUST be set to
// 0xffffff and the extended timestamp MUST be sent.
const extendedTimestamp = uint64(0xffffff)

// The intput or output settings for RTMP protocol.
type settings struct {
	chunkSize uint32
}

func newSettings() *settings {
	return &settings{
		chunkSize: 128,
	}
}

// The chunk stream which transport a message once.
type chunkStream struct {
	format            FormatType
	cid               ChunkID
	header            *Message
	message           *Message
	count             uint64
	extendedTimestamp bool
}

func newChunkStream() *chunkStream {
	return &chunkStream{
		header: NewMessage(),
	}
}

// The protocol implements the RTMP command and chunk stack.
type Protocol struct {
	input struct {
		opt    *settings
		chunks map[uint8]*chunkStream
	}
	output struct {
		opt *settings
	}
}

func NewProtocol() *Protocol {
	v := &Protocol{}
	v.input.opt = newSettings()
	v.input.chunks = map[uint8]*chunkStream{}
	v.output.opt = newSettings()
	return v
}

func (v *Protocol) ReadMessage(r io.Reader) (m *Message, err error) {
	for m == nil {
		var cid ChunkID
		var format FormatType
		if format, cid, err = v.readBasicHeader(r); err != nil {
			return
		}

		var ok bool
		var chunk *chunkStream
		if chunk, ok = v.input.chunks[cid]; !ok {
			chunk = newChunkStream()
			v.input.chunks[cid] = chunk
			chunk.header.betterCid = cid
		}

		if err = v.readMessageHeader(r, chunk, format); err != nil {
			return
		}

		if m, err = v.readMessagePayload(r, chunk); err != nil {
			return
		}
	}

	return
}

func (v *Protocol) readMessagePayload(r io.Reader, chunk *chunkStream) (m *Message, err error) {
	// Empty payload message.
	if chunk.header.payloadLength == 0 {
		m = chunk.message
		chunk.message = nil
		return
	}

	// Calculate the chunk payload size.
	chunkedPayloadSize := int(chunk.header.payloadLength) - len(chunk.message.payload)
	if chunkedPayloadSize > int(v.input.opt.chunkSize) {
		chunkedPayloadSize = int(v.input.opt.chunkSize)
	}

	b := &bytes.Buffer{}
	if _, err = io.CopyN(b, r, int64(chunkedPayloadSize)); err != nil {
		return
	}

	chunk.message.payload = append(chunk.message.payload, b.Bytes()...)

	// Got entire RTMP message?
	if int(chunk.message.payloadLength) == len(chunk.message.payload) {
		m = chunk.message
		chunk.message = nil
	}

	return
}

// Please read @doc rtmp_specification_1.0.pdf, @page 18, @section 6.1.2. Chunk Message Header
// There are four different formats for the chunk message header,
// selected by the "fmt" field in the chunk basic header.
type FormatType uint8

const (
	// 6.1.2.1. Type 0
	// Chunks of Type 0 are 11 bytes long. This type MUST be used at the
	// start of a chunk stream, and whenever the stream timestamp goes
	// backward (e.g., because of a backward seek).
	FormatType0 FormatType = iota
	// 6.1.2.2. Type 1
	// Chunks of Type 1 are 7 bytes long. The message stream ID is not
	// included; this chunk takes the same stream ID as the preceding chunk.
	// Streams with variable-sized messages (for example, many video
	// formats) SHOULD use this format for the first chunk of each new
	// message after the first.
	FormatType1
	// 6.1.2.3. Type 2
	// Chunks of Type 2 are 3 bytes long. Neither the stream ID nor the
	// message length is included; this chunk has the same stream ID and
	// message length as the preceding chunk. Streams with constant-sized
	// messages (for example, some audio and data formats) SHOULD use this
	// format for the first chunk of each message after the first.
	FormatType2
	// 6.1.2.4. Type 3
	// Chunks of Type 3 have no header. Stream ID, message length and
	// timestamp delta are not present; chunks of this type take values from
	// the preceding chunk. When a single message is split into chunks, all
	// chunks of a message except the first one, SHOULD use this type. Refer
	// to example 2 in section 6.2.2. Stream consisting of messages of
	// exactly the same size, stream ID and spacing in time SHOULD use this
	// type for all chunks after chunk of Type 2. Refer to example 1 in
	// section 6.2.1. If the delta between the first message and the second
	// message is same as the time stamp of first message, then chunk of
	// type 3 would immediately follow the chunk of type 0 as there is no
	// need for a chunk of type 2 to register the delta. If Type 3 chunk
	// follows a Type 0 chunk, then timestamp delta for this Type 3 chunk is
	// the same as the timestamp of Type 0 chunk.
	FormatType3
)

// The message header size, index is format.
const messageHeaderSizes = []int{11, 7, 3, 0}

// Parse the chunk message header.
//   3bytes: timestamp delta,    fmt=0,1,2
//   3bytes: payload length,     fmt=0,1
//   1bytes: message type,       fmt=0,1
//   4bytes: stream id,          fmt=0
// where:
//   fmt=0, 0x0X
//   fmt=1, 0x4X
//   fmt=2, 0x8X
//   fmt=3, 0xCX
func (v *Protocol) readMessageHeader(r io.Reader, chunk *chunkStream, format FormatType) (err error) {
	// We should not assert anything about fmt, for the first packet.
	// (when first packet, the chunk.message is nil).
	// the fmt maybe 0/1/2/3, the FMLE will send a 0xC4 for some audio packet.
	// the previous packet is:
	//     04                // fmt=0, cid=4
	//     00 00 1a          // timestamp=26
	//     00 00 9d          // payload_length=157
	//     08                // message_type=8(audio)
	//     01 00 00 00       // stream_id=1
	// the current packet maybe:
	//     c4             // fmt=3, cid=4
	// it's ok, for the packet is audio, and timestamp delta is 26.
	// the current packet must be parsed as:
	//     fmt=0, cid=4
	//     timestamp=26+26=52
	//     payload_length=157
	//     message_type=8(audio)
	//     stream_id=1
	// so we must update the timestamp even fmt=3 for first packet.
	//
	// The fresh packet used to update the timestamp even fmt=3 for first packet.
	// fresh packet always means the chunk is the first one of message.
	var isFirstChunkOfMsg bool
	if chunk.message == nil {
		isFirstChunkOfMsg = true
	}

	// But, we can ensure that when a chunk stream is fresh,
	// the fmt must be 0, a new stream.
	if chunk.count == 0 && format != FormatType0 {
		// For librtmp, if ping, it will send a fresh stream with fmt=1,
		// 0x42             where: fmt=1, cid=2, protocol contorl user-control message
		// 0x00 0x00 0x00   where: timestamp=0
		// 0x00 0x00 0x06   where: payload_length=6
		// 0x04             where: message_type=4(protocol control user-control message)
		// 0x00 0x06            where: event Ping(0x06)
		// 0x00 0x00 0x0d 0x0f  where: event data 4bytes ping timestamp.
		// @see: https://github.com/ossrs/srs/issues/98
		if chunk.cid == ChunkIDProtocolControl && format == FormatType1 {
			// We accept cid=2, fmt=1 to make librtmp happy.
		} else {
			return fmt.Errorf("For fresh chunk, fmt %v != %v(required), cid is %v", format, FormatType0, chunk.cid)
		}
	}

	// When exists cache msg, means got an partial message,
	// the fmt must not be type0 which means new message.
	if chunk.message != nil && format == FormatType0 {
		return fmt.Errorf("For exists chunk, fmt is %v, cid is %v", format, chunk.cid)
	}

	// Create msg when new chunk stream start
	if chunk.message == nil {
		chunk.message = NewMessage()
	}

	// Read the message header.
	b := &bytes.Buffer{}
	if _, err = io.CopyN(b, r, int64(messageHeaderSizes[format])); err != nil {
		return
	}

	// Prse the message header.
	//   3bytes: timestamp delta,    fmt=0,1,2
	//   3bytes: payload length,     fmt=0,1
	//   1bytes: message type,       fmt=0,1
	//   4bytes: stream id,          fmt=0
	// where:
	//   fmt=0, 0x0X
	//   fmt=1, 0x4X
	//   fmt=2, 0x8X
	//   fmt=3, 0xCX
	if format <= FormatType2 {
		p := b.Bytes()
		chunk.header.timestampDelta = uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16
		p = p[3:]

		// fmt: 0
		// timestamp: 3 bytes
		// If the timestamp is greater than or equal to 16777215
		// (hexadecimal 0x00ffffff), this value MUST be 16777215, and the
		// 'extended timestamp header' MUST be present. Otherwise, this value
		// SHOULD be the entire timestamp.
		//
		// fmt: 1 or 2
		// timestamp delta: 3 bytes
		// If the delta is greater than or equal to 16777215 (hexadecimal
		// 0x00ffffff), this value MUST be 16777215, and the 'extended
		// timestamp header' MUST be present. Otherwise, this value SHOULD be
		// the entire delta.
		chunk.extendedTimestamp = false
		if uint64(chunk.header.timestampDelta) >= extendedTimestamp {
			chunk.extendedTimestamp = true

			// Extended timestamp: 0 or 4 bytes
			// This field MUST be sent when the normal timsestamp is set to
			// 0xffffff, it MUST NOT be sent if the normal timestamp is set to
			// anything else. So for values less than 0xffffff the normal
			// timestamp field SHOULD be used in which case the extended timestamp
			// MUST NOT be present. For values greater than or equal to 0xffffff
			// the normal timestamp field MUST NOT be used and MUST be set to
			// 0xffffff and the extended timestamp MUST be sent.
			if format == FormatType0 {
				// 6.1.2.1. Type 0
				// For a type-0 chunk, the absolute timestamp of the message is sent
				// here.
				chunk.header.timestamp = uint64(chunk.header.timestampDelta)
			} else {
				// 6.1.2.2. Type 1
				// 6.1.2.3. Type 2
				// For a type-1 or type-2 chunk, the difference between the previous
				// chunk's timestamp and the current chunk's timestamp is sent here.
				chunk.header.timestamp += uint64(chunk.header.timestampDelta)
			}
		}

		if format <= FormatType1 {
			payloadLength := uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16
			p = p[3:]

			// For a message, if msg exists in cache, the size must not changed.
			// always use the actual msg size to compare, for the cache payload length can changed,
			// for the fmt type1(stream_id not changed), user can change the payload
			// length(it's not allowed in the continue chunks).
			if !isFirstChunkOfMsg && chunk.header.payloadLength != payloadLength {
				return fmt.Errorf("Chunk message size %v != %v(required)", payloadLength, chunk.header.payloadLength)
			}
			chunk.header.payloadLength = payloadLength

			chunk.header.messageType = MessageType(p[0])
			p = p[1:]

			if format == FormatType0 {
				chunk.header.streamID = uint32(p[0])<<24 | uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3])
				p = p[4:]
			}
		}
	} else {
		// Update the timestamp even fmt=3 for first chunk packet
		if isFirstChunkOfMsg && !chunk.extendedTimestamp {
			chunk.header.timestamp += uint64(chunk.header.timestampDelta)
		}
	}

	// Read extended-timestamp
	if chunk.extendedTimestamp {
		var timestamp uint32
		if err = binary.Read(r, binary.BigEndian, &timestamp); err != nil {
			return
		}

		// We always use 31bits timestamp, for some server may use 32bits extended timestamp.
		// @see https://github.com/ossrs/srs/issues/111
		timestamp &= 0x7fffffff

		// TODO: FIXME: Support detect the extended timestamp.
		// @see http://blog.csdn.net/win_lin/article/details/13363699
		chunk.header.timestamp = uint64(timestamp)
	}

	// The extended-timestamp must be unsigned-int,
	//         24bits timestamp: 0xffffff = 16777215ms = 16777.215s = 4.66h
	//         32bits timestamp: 0xffffffff = 4294967295ms = 4294967.295s = 1193.046h = 49.71d
	// because the rtmp protocol says the 32bits timestamp is about "50 days":
	//         3. Byte Order, Alignment, and Time Format
	//                Because timestamps are generally only 32 bits long, they will roll
	//                over after fewer than 50 days.
	//
	// but, its sample says the timestamp is 31bits:
	//         An application could assume, for example, that all
	//        adjacent timestamps are within 2^31 milliseconds of each other, so
	//        10000 comes after 4000000000, while 3000000000 comes before
	//        4000000000.
	// and flv specification says timestamp is 31bits:
	//        Extension of the Timestamp field to form a SI32 value. This
	//        field represents the upper 8 bits, while the previous
	//        Timestamp field represents the lower 24 bits of the time in
	//        milliseconds.
	// in a word, 31bits timestamp is ok.
	// convert extended timestamp to 31bits.
	chunk.header.timestamp &= 0x7fffffff

	// Copy header to msg
	chunk.message.timestamp = chunk.header.timestamp
	chunk.message.timestampDelta = chunk.header.timestampDelta
	chunk.message.betterCid = chunk.header.betterCid
	chunk.message.messageType = chunk.header.messageType
	chunk.message.payloadLength = chunk.header.payloadLength
	chunk.message.streamID = chunk.header.streamID

	// Increase the msg count, the chunk stream can accept fmt=1/2/3 message now.
	chunk.count++

	return
}

// Please read @doc rtmp_specification_1.0.pdf, @page 17, @section 6.1.1. Chunk Basic Header
// The Chunk Basic Header encodes the chunk stream ID and the chunk
// type(represented by fmt field in the figure below). Chunk type
// determines the format of the encoded message header. Chunk Basic
// Header field may be 1, 2, or 3 bytes, depending on the chunk stream
// ID.
//
// The bits 0-5 (least significant) in the chunk basic header represent
// the chunk stream ID.
//
// Chunk stream IDs 2-63 can be encoded in the 1-byte version of this
// field.
//    0 1 2 3 4 5 6 7
//   +-+-+-+-+-+-+-+-+
//   |fmt|   cs id   |
//   +-+-+-+-+-+-+-+-+
//   Figure 6 Chunk basic header 1
//
// Chunk stream IDs 64-319 can be encoded in the 2-byte version of this
// field. ID is computed as (the second byte + 64).
//   0                   1
//   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5
//   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//   |fmt|    0      | cs id - 64    |
//   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//   Figure 7 Chunk basic header 2
//
// Chunk stream IDs 64-65599 can be encoded in the 3-byte version of
// this field. ID is computed as ((the third byte)*256 + the second byte
// + 64).
//    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3
//   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//   |fmt|     1     |         cs id - 64            |
//   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//   Figure 8 Chunk basic header 3
//
// cs id: 6 bits
// fmt: 2 bits
// cs id - 64: 8 or 16 bits
//
// Chunk stream IDs with values 64-319 could be represented by both 2-
// byte version and 3-byte version of this field.
func (v *Protocol) readBasicHeader(r io.Reader) (format FormatType, cid ChunkID, err error) {
	// 2-63, 1B chunk header
	var t uint8
	if err = binary.Read(r, binary.BigEndian, &t); err != nil {
		return
	}
	cid = ChunkID(t & 0x3f)
	format = FormatType((t >> 6) & 0x03)

	if cid > 1 {
		return
	}

	// 64-319, 2B chunk header
	if err = binary.Read(r, binary.BigEndian, &t); err != nil {
		return
	}
	cid = ChunkID(64 + uint32(t))

	// 64-65599, 3B chunk header
	if cid == 1 {
		if err = binary.Read(r, binary.BigEndian, &t); err != nil {
			return
		}
		cid += ChunkID(uint32(t) * 256)
	}

	return
}

func (v *Protocol) WritePacket(w io.Writer, pkt Packet, streamID int) (err error) {
	m := NewMessage()

	if m.payload, err = pkt.MarshalBinary(); err != nil {
		return
	}

	m.payloadLength = uint32(len(m.payload))
	m.messageType = pkt.Type()
	m.streamID = uint32(streamID)
	m.betterCid = pkt.BetterCid()

	if err = v.writeMessage(w, m); err != nil {
		return
	}

	if err = v.onPacketWriten(m, pkt); err != nil {
		return
	}

	return
}

func (v *Protocol) onPacketWriten(m *Message, pkt Packet) (err error) {
	// TODO: FIXME: Implements it.
	return
}

func (v *Protocol) writeMessage(w io.Writer, m *Message) (err error) {
	// TODO: FIXME: Use writev to write for high performance.
	var c0h, c3h []byte
	if c0h, err = m.generateC0Header(); err != nil {
		return
	}
	if c3h, err = m.generateC3Header(); err != nil {
		return
	}

	var h []byte
	p := m.payload
	for len(p) > 0 {
		if h == nil {
			h = c0h
		} else {
			h = c3h
		}

		if _, err = io.Copy(w, bytes.NewReader(h)); err != nil {
			return
		}

		size := len(p)
		if size > int(v.output.opt.chunkSize) {
			size = int(v.output.opt.chunkSize)
		}

		if _, err = io.Copy(w, bytes.NewReader(p[:size])); err != nil {
			return
		}
		p = p[size:]
	}

	return
}

// Please read @doc rtmp_specification_1.0.pdf, @page 30, @section 4.1. Message Header
// 1byte. One byte field to represent the message type. A range of type IDs
// (1-7) are reserved for protocol control messages.
type MessageType uint8

const (
	// Please read @doc rtmp_specification_1.0.pdf, @page 30, @section 5. Protocol Control Messages
	// RTMP reserves message type IDs 1-7 for protocol control messages.
	// These messages contain information needed by the RTM Chunk Stream
	// protocol or RTMP itself. Protocol messages with IDs 1 & 2 are
	// reserved for usage with RTM Chunk Stream protocol. Protocol messages
	// with IDs 3-6 are reserved for usage of RTMP. Protocol message with ID
	// 7 is used between edge server and origin server.
	MessageTypeSetChunkSize MessageType = 0x01 + iota
	MessageTypeAbortMessage
	MessageTypeAcknowledgement
	MessageTypeUserControlMessage
	MessageTypeWindowAcknowledgementSize
	MessageTypeSetPeerBandwidth
	MessageTypeEdgeAndOriginServerCommand
	// Please read @doc rtmp_specification_1.0.pdf, @page 38, @section 3. Types of messages
	// The server and the client send messages over the network to
	// communicate with each other. The messages can be of any type which
	// includes audio messages, video messages, command messages, shared
	// object messages, data messages, and user control messages.
	//
	// Please read @doc rtmp_specification_1.0.pdf, @page 41, @section 3.4. Audio message
	// The client or the server sends this message to send audio data to the
	// peer. The message type value of 8 is reserved for audio messages.
	MessageTypeAudio MessageType = 0x08 + iota
	// Please read @doc rtmp_specification_1.0.pdf, @page 41, @section 3.5. Video message
	// The client or the server sends this message to send video data to the
	// peer. The message type value of 9 is reserved for video messages.
	// These messages are large and can delay the sending of other type of
	// messages. To avoid such a situation, the video message is assigned
	// the lowest priority.
	MessageTypeVideo
	// Please read @doc rtmp_specification_1.0.pdf, @page 38, @section 3.1. Command message
	// Command messages carry the AMF-encoded commands between the client
	// and the server. These messages have been assigned message type value
	// of 20 for AMF0 encoding and message type value of 17 for AMF3
	// encoding. These messages are sent to perform some operations like
	// connect, createStream, publish, play, pause on the peer. Command
	// messages like onstatus, result etc. are used to inform the sender
	// about the status of the requested commands. A command message
	// consists of command name, transaction ID, and command object that
	// contains related parameters. A client or a server can request Remote
	// Procedure Calls (RPC) over streams that are communicated using the
	// command messages to the peer.
	MessageTypeAMF3CommandMessage MessageType = 17 // 0x11
	MessageTypeAMF0CommandMessage MessageType = 20 // 0x14
	// Please read @doc rtmp_specification_1.0.pdf, @page 38, @section 3.2. Data message
	// The client or the server sends this message to send Metadata or any
	// user data to the peer. Metadata includes details about the
	// data(audio, video etc.) like creation time, duration, theme and so
	// on. These messages have been assigned message type value of 18 for
	// AMF0 and message type value of 15 for AMF3.
	MessageTypeAMF0DataMessage MessageType = 18 // 0x12
	MessageTypeAMF3DataMessage MessageType = 15 // 0x0f
)

// The RTMP message, transport over chunk stream in RTMP.
// Please read the cs id of @doc rtmp_specification_1.0.pdf, @page 30, @section 4.1. Message Header
type Message struct {
	// 3bytes.
	// Three-byte field that contains a timestamp delta of the message.
	// @remark, only used for decoding message from chunk stream.
	timestampDelta uint32
	// 3bytes.
	// Three-byte field that represents the size of the payload in bytes.
	// It is set in big-endian format.
	payloadLength uint32
	// 1byte.
	// One byte field to represent the message type. A range of type IDs
	// (1-7) are reserved for protocol control messages.
	messageType MessageType
	// 4bytes.
	// Four-byte field that identifies the stream of the message. These
	// bytes are set in little-endian format.
	streamID uint32

	// The chunk stream id over which transport.
	betterCid ChunkID

	// Four-byte field that contains a timestamp of the message.
	// The 4 bytes are packed in the big-endian order.
	// @remark, we use 64bits for large time for jitter detect and for large tbn like HLS.
	timestamp uint64

	// The payload which carries the RTMP packet.
	payload []byte
}

func NewMessage() *Message {
	return &Message{}
}

func (v *Message) generateC3Header() ([]byte, error) {
	var c3h []byte
	if v.timestamp < extendedTimestamp {
		c3h = make([]byte, 1)
	} else {
		c3h = make([]byte, 1+4)
	}

	p := c3h
	p[0] = 0xc0 | byte(v.betterCid&0x3f)
	p = p[1:]

	// In RTMP protocol, there must not any timestamp in C3 header,
	// but actually all products from adobe, such as FMS/AMS and Flash player and FMLE,
	// always carry a extended timestamp in C3 header.
	// @see: http://blog.csdn.net/win_lin/article/details/13363699
	if v.timestamp >= extendedTimestamp {
		p[0] = byte(v.timestamp >> 24)
		p[1] = byte(v.timestamp >> 16)
		p[2] = byte(v.timestamp >> 8)
		p[3] = byte(v.timestamp)
	}

	return c3h, nil
}

func (v *Message) generateC0Header() ([]byte, error) {
	var c0h []byte
	if v.timestamp < extendedTimestamp {
		c0h = make([]byte, 1+3+3+1+4)
	} else {
		c0h = make([]byte, 1+3+3+1+4+4)
	}

	p := c0h
	p[0] = byte(v.betterCid) & 0x3f
	p = p[1:]

	if v.timestamp < extendedTimestamp {
		p[0] = byte(v.timestamp >> 16)
		p[1] = byte(v.timestamp >> 8)
		p[2] = byte(v.timestamp)
	} else {
		p[0] = 0xff
		p[1] = 0xff
		p[2] = 0xff
	}
	p = p[3:]

	p[0] = byte(v.payloadLength >> 16)
	p[1] = byte(v.payloadLength >> 8)
	p[2] = byte(v.payloadLength)
	p = p[3:]

	p[0] = byte(v.messageType)
	p = p[1:]

	p[0] = byte(v.streamID)
	p[1] = byte(v.streamID >> 8)
	p[2] = byte(v.streamID >> 16)
	p[3] = byte(v.streamID >> 24)
	p = p[4:]

	if v.timestamp >= extendedTimestamp {
		p[0] = byte(v.timestamp >> 24)
		p[1] = byte(v.timestamp >> 16)
		p[2] = byte(v.timestamp >> 8)
		p[3] = byte(v.timestamp)
	}

	return c0h, nil
}

// Please read the cs id of @doc rtmp_specification_1.0.pdf, @page 17, @section 6.1.1. Chunk Basic Header
type ChunkID uint32

const (
	ChunkIDProtocolControl ChunkID = 0x02 + iota
	ChunkIDOverConnection
	ChunkIDOverConnection2
	ChunkIDOverStream
	ChunkIDOverStream2
	ChunkIDVideo
	ChunkIDAudio
)

// The Command Name of message.
const (
	CommandConnect          amf0.String = amf0.String("connect")
	CommandCreateStream     amf0.String = amf0.String("createStream")
	CommandCloseStream      amf0.String = amf0.String("closeStream")
	CommandPlay             amf0.String = amf0.String("play")
	CommandPause            amf0.String = amf0.String("pause")
	CommandOnBWDone         amf0.String = amf0.String("onBWDone")
	CommandOnStatus         amf0.String = amf0.String("onStatus")
	CommandResult           amf0.String = amf0.String("_result")
	CommandError            amf0.String = amf0.String("_error")
	CommandReleaseStream    amf0.String = amf0.String("releaseStream")
	CommandFCPublish        amf0.String = amf0.String("FCPublish")
	CommandFCUnpublish      amf0.String = amf0.String("FCUnpublish")
	CommandPublish          amf0.String = amf0.String("publish")
	CommandRtmpSampleAccess amf0.String = amf0.String("|RtmpSampleAccess")
)

// The RTMP packet, transport as payload of RTMP message.
type Packet interface {
	// Marshaler and unmarshaler
	Size() int
	encoding.BinaryUnmarshaler
	encoding.BinaryMarshaler

	// RTMP protocol fields for each packet.
	BetterCid() ChunkID
	Type() MessageType
}

// Please read @doc rtmp_specification_1.0.pdf, @page 45, @section 4.1.1. connect
// The client sends the connect command to the server to request
// connection to a server application instance.
type ConnectAppPacket struct {
	// Name of the command. Set to "connect".
	CommandName amf0.String
	// Always set to 1.
	TransactionID amf0.Number
	// Command information object which has the name-value pairs.
	CommandObject *amf0.Object
	// Any optional information
	Args *amf0.Object
}

func NewConnectAppPacket() *ConnectAppPacket {
	return &ConnectAppPacket{
		CommandName:   CommandConnect,
		CommandObject: amf0.NewObject(),
		TransactionID: amf0.Number(1.0),
	}
}

func (v *ConnectAppPacket) BetterCid() ChunkID {
	return ChunkIDOverConnection
}

func (v *ConnectAppPacket) Type() MessageType {
	return MessageTypeAMF0CommandMessage
}

func (v *ConnectAppPacket) Size() int {
	size := v.CommandName.Size() + v.TransactionID.Size() + v.CommandObject.Size()
	if v.Args != nil {
		size += v.Args.Size()
	}
	return size
}

func (v *ConnectAppPacket) UnmarshalBinary(data []byte) (err error) {
	p := data
	if err = v.CommandName.UnmarshalBinary(p); err != nil {
		return
	}
	if v.CommandName != CommandConnect {
		return fmt.Errorf("Invalid command name %v", string(v.CommandName))
	}
	p = p[v.CommandName.Size():]

	if err = v.TransactionID.UnmarshalBinary(p); err != nil {
		return
	}
	if v.TransactionID != 1.0 {
		return fmt.Errorf("Invalid transaction ID %v", float64(v.TransactionID))
	}
	p = p[v.TransactionID.Size():]

	if err = v.CommandObject.UnmarshalBinary(p); err != nil {
		return
	}
	p = p[v.CommandObject.Size():]

	if len(p) == 0 {
		return
	}

	v.Args = amf0.NewObject()
	if err = v.Args.UnmarshalBinary(p); err != nil {
		return
	}

	return
}

func (v *ConnectAppPacket) MarshalBinary() (data []byte, err error) {
	var pb []byte
	if pb, err = v.CommandName.MarshalBinary(); err != nil {
		return
	}
	data = append(data, pb...)

	if pb, err = v.TransactionID.MarshalBinary(); err != nil {
		return
	}
	data = append(data, pb...)

	if pb, err = v.CommandObject.MarshalBinary(); err != nil {
		return
	}
	data = append(data, pb...)

	if v.Args != nil {
		if pb, err = v.Args.MarshalBinary(); err != nil {
			return
		}
		data = append(data, pb...)
	}

	return
}
