package main

import (
    "fmt"
    "time"
    "net"
    "math/rand"
    "encoding/binary"
    "bytes"
    "crypto/sha256"
    )

// Bitcoin protocol constants for this node
var PROTOCOL_VERSION int32 = 70015
var CADDR_TIME_VERSION uint32 = 31402
var MAINNET_MAGIC uint32 = 0xD9B4BEF9 // bitcoin main network
var MAINNET_TCP_PORT uint16 = 8333 // bitcoin main network port
var TESTNET_MAGIC uint32 = 0xDAB5BFFA // bitcoin test network
var TESTNET_TCP_PORT uint16 = 18333 // bitcoin test network port
var NODE_SERVICES uint64 = 1

// Generic protocol message header
type message_header struct {
    magic uint32
    command [12]byte
    length uint32
    checksum uint32
}

func (h message_header) serialize() []byte {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.LittleEndian, h)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    return buffer.Bytes()
}

// version message requires a network address struct with no time field
type net_addr_no_time struct {
    services uint64
    ip [16]byte
    port uint16
}

func (n net_addr_no_time) serialize() []byte {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.LittleEndian, n.services)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, n.ip)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.BigEndian, n.port)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    return buffer.Bytes()
}

type net_addr struct {
    time uint32
    net_addr_no_time
}

func (n net_addr) serialize() []byte {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.LittleEndian, n.time)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    return append(buffer.Bytes(), n.net_addr_no_time.serialize()...)
}

type version struct {
    version int32
    services uint64
    timestamp int64
    addr_recv net_addr_no_time
    addr_from net_addr_no_time
    nonce uint64
    user_agent [1]byte
    start_height int32
    relay bool
}

// TODO: shrink this code. How can we iterate over each struct member?
func (v version) serialize() []byte {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.LittleEndian, v.version)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.services)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.timestamp)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.addr_recv.serialize())
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.addr_from.serialize())
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.nonce)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.user_agent)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.start_height)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    err = binary.Write(buffer, binary.LittleEndian, v.relay)
    if err != nil {
        fmt.Println(err)
        return new(bytes.Buffer).Bytes()
    }

    return buffer.Bytes()
}

// Calculate message checksum. First 4 bytes of sha256(sha256(payload))
func checksum(slice []byte) uint32 {
    hash := sha256.New()
    hash.Write(slice)
    sum := hash.Sum(nil)
    hash.Reset()
    hash.Write(sum)
    sum = hash.Sum(nil)
    return binary.LittleEndian.Uint32(sum[:4])
}

// Combine the network message header with the payload
func build_message(magic uint32, command string, payload []byte) []byte {
    var h message_header
    h.magic = magic
    h.checksum = checksum(payload)
    h.length = uint32(len(payload))
    copy(h.command[:], command)
    return append(h.serialize(), payload...)
}

func build_version_message(user_agent string, last_block int32) []byte {
    // Seed RNG for version message nonce
    rand.Seed(time.Now().UnixNano())

    // Create version message
    var v version

    // Populate version message
    v.version = PROTOCOL_VERSION
    v.services = NODE_SERVICES
    v.timestamp = int64(time.Now().Unix())

    // Set structure addresses
    var r net_addr_no_time
    r.services = NODE_SERVICES
    ip := net.IP([]byte{127, 0, 0, 1})
    copy(r.ip[:], ip.To16())
    r.port = MAINNET_TCP_PORT
    v.addr_recv = r
    v.addr_from = r

    v.nonce = rand.Uint64()
    copy(v.user_agent[:], user_agent)
    v.start_height = last_block
    v.relay = true

    return build_message(MAINNET_MAGIC, "version", v.serialize())
}

func main() {
    version_message := build_version_message("", 0)
    fmt.Println("Send version message:", version_message)

    // connect to this socket
    conn, err := net.Dial("tcp", "209.73.142.226:8333")
    if err != nil {
        fmt.Println(err)
    }

    conn.Write(version_message)
    buff := make([]byte, 1024)

    n, _ := conn.Read(buff)
    fmt.Println("Message received from server:", buff[:n])
}
