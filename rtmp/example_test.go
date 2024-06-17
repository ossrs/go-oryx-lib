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

package rtmp_test

import (
	"math/rand"
	"net"
	"time"

	"github.com/ossrs/go-oryx-lib/rtmp"
	"github.com/ossrs/go-oryx-lib/amf0"
)

func ExampleRtmpClientHandshake() {
	// Connect to RTMP server via TCP client.
	c, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1935})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	hs := rtmp.NewHandshake(rd)

	// Client send C0,C1
	if err := hs.WriteC0S0(c); err != nil {
		panic(err)
	}
	if err := hs.WriteC1S1(c); err != nil {
		panic(err)
	}

	// Receive S0,S1,S2 from RTMP server.
	if _, err = hs.ReadC0S0(c); err != nil {
		panic(err)
	}
	s1, err := hs.ReadC1S1(c)
	if err != nil {
		panic(err)
	}
	if _, err = hs.ReadC2S2(c); err != nil {
		panic(err)
	}

	// Client send C2
	if err := hs.WriteC2S2(c, s1); err != nil {
		panic(err)
	}
}

func ExampleRtmpClientConnect() {
	// Connect to RTMP server via TCP, then finish the RTMP handshake see ExampleRtmpClientHandshake.
	var c *net.TCPConn

	// Create a RTMP client.
	client := rtmp.NewProtocol(c)

	// Send RTMP connect tcURL packet.
	connectApp := rtmp.NewConnectAppPacket()
	connectApp.CommandObject.Set("tcUrl", amf0.NewString("rtmp://localhost/live"))
	if err := client.WritePacket(connectApp, 1); err != nil {
		panic(err)
	}

	var connectAppRes *rtmp.ConnectAppResPacket
	if _, err := client.ExpectPacket(&connectAppRes); err != nil {
		panic(err)
	}
}
