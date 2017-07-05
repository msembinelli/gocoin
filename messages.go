package main

import (
    "fmt"
    "time"
    "net"
    "math/rand"
    "encoding/binary"
    "bytes"
    "crypto/sha256"
    "reflect"
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
type MessageHeader struct {
    Magic uint32
    Command [12]byte
    Length uint32
    Checksum uint32
}

func (h MessageHeader) serialize() []byte {
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
type NetAddrNoTime struct {
    Services uint64
    Ip [16]byte
    Port uint16
}

func (n NetAddrNoTime) Serialize() []byte {
    buffer := new(bytes.Buffer)

    v := reflect.ValueOf(n)

    for i := 0; i < v.NumField(); i++ {
        var err error
        // Port is the only field to be written in BigEndian format
        if v.Type().Field(i).Name != "Port" {
            err = binary.Write(buffer, binary.LittleEndian, v.Field(i).Interface())
        } else {
            err = binary.Write(buffer, binary.BigEndian, v.Field(i).Interface())
        }
        if err != nil {
            fmt.Println(err)
            // Return empty buffer
            return new(bytes.Buffer).Bytes()
        }
    }

    return buffer.Bytes()
}

type NetAddr struct {
    Time uint32
    NetAddrNoTime
}

func (n NetAddr) Serialize() []byte {
    buffer := new(bytes.Buffer)

    err := binary.Write(buffer, binary.LittleEndian, n.Time)
    if err != nil {
        fmt.Println(err)
        // Return empty buffer
        return new(bytes.Buffer).Bytes()
    }

    return append(buffer.Bytes(), n.NetAddrNoTime.Serialize()...)
}

type Version struct {
    Version int32
    Services uint64
    Timestamp int64
    AddrRecv NetAddrNoTime
    AddrFrom NetAddrNoTime
    Nonce uint64
    UserAgent [1]byte
    StartHeight int32
    Relay bool
}

func (v Version) Serialize() []byte {
    buffer := new(bytes.Buffer)

    s := reflect.ValueOf(v)

    for i := 0; i < s.NumField(); i++ {
        var err error
        if reflect.Value.Type(s.Field(i)).Name() == "NetAddrNoTime" {
            r := reflect.ValueOf(s.Field(i).Interface()).MethodByName("Serialize")
            err = binary.Write(buffer, binary.LittleEndian, r.Call([]reflect.Value{})[0].Bytes())
        } else {
            err = binary.Write(buffer, binary.LittleEndian, s.Field(i).Interface())
        }
        if err != nil {
            fmt.Println(err)
            // Return empty buffer
            return new(bytes.Buffer).Bytes()
        }
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
    var h MessageHeader
    h.Magic = magic
    h.Checksum = checksum(payload)
    h.Length = uint32(len(payload))
    copy(h.Command[:], command)
    return append(h.serialize(), payload...)
}

func build_version_message(magic uint32, user_agent string, last_block int32) []byte {
    // Seed RNG for version message nonce
    rand.Seed(time.Now().UnixNano())

    // Create version message
    var v Version

    // Populate version message
    v.Version = PROTOCOL_VERSION
    v.Services = NODE_SERVICES
    v.Timestamp = int64(time.Now().Unix())

    // Set structure addresses
    var r NetAddrNoTime
    r.Services = NODE_SERVICES
    ip := net.IP([]byte{127, 0, 0, 1})
    copy(r.Ip[:], ip.To16())
    r.Port = MAINNET_TCP_PORT
    v.AddrRecv = r
    v.AddrFrom = r

    v.Nonce = rand.Uint64()
    copy(v.UserAgent[:], user_agent)
    v.StartHeight = last_block
    v.Relay = true

    return build_message(magic, "version", v.Serialize())
}

func main() {
    version_message := build_version_message(MAINNET_MAGIC, "", 0)
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
