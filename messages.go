package main

import (
    "fmt"
    "time"
    "net"
    "math/rand"
    )
// import(
//     "encoding/binary"
//     "bytes"
// )

var PROTOCOL_VERSION int32 = 70015
var CADDR_TIME_VERSION uint32 = 31402
var MAINNET_MAGIC uint32 = 0xF9BEB4D9 // bitcoin main network
var MAINNET_TCP_PORT uint16 = 8333 // bitcoin main network port
var TESTNET_MAGIC uint32 = 0x0B110907 // bitcoin test network
var TESTNET_TCP_PORT uint16 = 18333 // bitcoin main network port

// TODO: Add serialization function
type net_addr struct {
    time uint32
    services uint64
    ip []byte
    port uint16
}

// TODO: Add serialization function
type version_msg struct {
    version int32
    services uint64
    timestamp int64
    addr_recv net_addr
    addr_from net_addr
    nonce uint64
    user_agent string
    start_height int32
    relay bool
}

func main() {
    var v version_msg
    v.version = PROTOCOL_VERSION
    v.services = 1
    v.timestamp = int64(time.Now().Unix())

    // Set structure addresses
    var r net_addr
    r.time = CADDR_TIME_VERSION
    r.services = 1
    ip := net.IP([]byte{127, 0, 0, 1})
    fmt.Println(ip.To16())
    r.ip = ip.To16()
    r.port = MAINNET_TCP_PORT
    v.addr_recv = r
    v.addr_from = r

    v.nonce = rand.Uint64()
    v.user_agent = ""
    v.start_height = 0
    v.relay = false

    // TODO: fix up and move to serialization functions for structs above
    // ip := binary.BigEndian.Uint32(net.IPv4(127, 0, 0, 1))
    // var port uint32 = 8333
    // buf := new(bytes.Buffer)
    // for i := 0; i < 2; i++ {
    //     // pack addr_me and addr_you
    //     err := binary.Write(buf, binary.BigEndian, ip)
    //     if err != nil {
    //      fmt.Println(err)
    //      return
    //     }
    //     buf_port := new(bytes.Buffer)
    //     err = binary.Write(buf, binary.BigEndian, port)
    //     if err != nil {
    //      fmt.Println(err)
    //      return
    //     }
    // }
}
