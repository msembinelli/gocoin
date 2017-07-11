package messages

import (
	"../protocol"
	"math/rand"
	"net"
	"time"
)

// Combine the network message header with the payload
func BuildMessage(magic uint32, command string, payload []byte) []byte {
	var h Header
	h.Magic = magic
	h.Checksum = protocol.MessageChecksum(payload)
	h.Length = uint32(len(payload))
	copy(h.Command[:], command)
	return append(h.Serialize(), payload...)
}

func BuildVersionMessage(magic uint32, user_agent string, last_block int32) []byte {
	// Seed RNG for version message nonce
	rand.Seed(time.Now().UnixNano())

	// Create version message
	var v Version

	// Populate version message
	v.Version = protocol.PROTOCOL_VERSION
	v.Services = protocol.NODE_SERVICES
	v.Timestamp = int64(time.Now().Unix())

	// Set structure addresses
	var r NetAddrNoTime
	r.Services = protocol.NODE_SERVICES
	ip := net.IP([]byte{127, 0, 0, 1})
	copy(r.Ip[:], ip.To16())
	r.Port = protocol.MAINNET_TCP_PORT
	v.AddrRecv = r
	v.AddrFrom = r

	v.Nonce = rand.Uint64()
	copy(v.UserAgent[:], user_agent)
	v.StartHeight = last_block
	v.Relay = true

	return BuildMessage(magic, "version", v.Serialize())
}

func BuildVerackMessage(magic uint32) []byte {
	return BuildMessage(magic, "verack", []byte{})
}

func BuildPongMessage(magic uint32) []byte {
	// Seed RNG for version message nonce
	rand.Seed(time.Now().UnixNano())
	var p Pong
	p.Nonce = rand.Uint64()

	return BuildMessage(magic, "pong", p.Serialize())
}

func BuildGetaddrMessage(magic uint32) []byte {
	return BuildMessage(magic, "getaddr", []byte{})
}
