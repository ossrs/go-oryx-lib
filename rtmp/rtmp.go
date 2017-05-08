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

// The protocol implements the RTMP command and chunk stack.
type Protocol struct {
	output *settings
}

func NewProtocol() *Protocol {
	return &Protocol{
		output: newSettings(),
	}
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
		if size > int(v.output.chunkSize) {
			size = int(v.output.chunkSize)
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
type ChunkID uint8

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
