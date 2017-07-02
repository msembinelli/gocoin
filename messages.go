package main

import (
    "fmt"
    "time"
    "net"
    "math/rand"
    "encoding/binary"
    "bytes"
    )

var PROTOCOL_VERSION int32 = 70015
var CADDR_TIME_VERSION uint32 = 31402
var MAINNET_MAGIC uint32 = 0xF9BEB4D9 // bitcoin main network
var MAINNET_TCP_PORT uint16 = 8333 // bitcoin main network port
var TESTNET_MAGIC uint32 = 0x0B110907 // bitcoin test network
var TESTNET_TCP_PORT uint16 = 18333 // bitcoin main network port

type net_addr struct {
    time uint32
    services uint64
    ip [16]byte
    port uint16
}

func (n net_addr) serialize() *bytes.Buffer {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.BigEndian, n)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer)
    }

    return buffer
}

type version_msg struct {
    version int32
    services uint64
    timestamp int64
    addr_recv net_addr
    addr_from net_addr
    nonce uint64
    user_agent [1]byte
    start_height int32
    relay bool
}

func (v version_msg) serialize() *bytes.Buffer {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.BigEndian, v)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer)
    }
    fmt.Println(buffer.Bytes())
    return buffer
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // Create version message
    var v version_msg
    v.version = PROTOCOL_VERSION
    v.services = 1
    v.timestamp = int64(time.Now().Unix())

    // Set structure addresses
    var r net_addr
    r.time = CADDR_TIME_VERSION
    r.services = 1
    ip := net.IP([]byte{127, 0, 0, 1})
    copy(r.ip[:], ip.To16())
    r.port = MAINNET_TCP_PORT
    v.addr_recv = r
    v.addr_from = r

    v.nonce = rand.Uint64()
    copy(v.user_agent[:], "")
    v.start_height = 0
    v.relay = false

    fmt.Println(v.serialize())

}
