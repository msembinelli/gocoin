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
    "bufio"
    "strings"
    "./message_fifo"
    )

// Bitcoin protocol constants for this node
var PROTOCOL_VERSION int32 = 70015
var CADDR_TIME_VERSION uint32 = 31402
var MAINNET_MAGIC uint32 = 0xD9B4BEF9 // bitcoin main network
var MAINNET_TCP_PORT uint16 = 8333 // bitcoin main network port
var TESTNET_MAGIC uint32 = 0xDAB5BFFA // bitcoin test network
var TESTNET_TCP_PORT uint16 = 18333 // bitcoin test network port
var NODE_SERVICES uint64 = 1
var MESSAGE_HEADER_LENGTH uint32 = 24

var TX_FIFO_SIZE uint32 = 32
var RX_FIFO_SIZE uint32 = 32
var TX_FIFO *message_fifo.FIFO
var RX_FIFO *message_fifo.FIFO

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
        // For types NetAddrNoTime, a custom serialize function exists, so we must call that instead
        // Reflect can do this, but it looks a little verbose
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
func message_checksum(slice []byte) uint32 {
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
    h.Checksum = message_checksum(payload)
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

func build_verack_message(magic uint32) []byte {
    return build_message(magic, "verack", []byte{})
}

func print_message_header(h MessageHeader) {
    fmt.Println("**MESSAGE HEADER**\n")
    fmt.Printf("  magic 0x%X\n", h.Magic)
    fmt.Printf("  type %s\n", string(h.Command[:]))
    fmt.Printf("  length %d\n", h.Length)
    fmt.Printf("  checksum 0x%X\n\n", h.Checksum)
}

func print_message_header_hex(header []byte) {
    fmt.Println("**MESSAGE HEADER HEX**")
    var i uint32
    i = 0
    for _, v := range header {
        if (i % 16) == 0 {
            fmt.Printf("\n")
        } else if (i % 8) == 0 {
            fmt.Printf(" ")
        }
        fmt.Printf("%02X ", v)
        i++
    }
    fmt.Printf("\n\n")
}

func deserialize_message_header(received []byte) MessageHeader {
    var h MessageHeader
    buf := bytes.NewReader(received)
    err := binary.Read(buf, binary.LittleEndian, &h)
    if err != nil {
        fmt.Println("binary.Read failed:", err)
    }
    print_message_header(h)
    print_message_header_hex(received)
    return h
}

func print_message_payload_hex(payload []byte, length uint32) {
    if length == 0 {
        return
    }

    fmt.Println("**MESSAGE PAYLOAD HEX**")
    var i uint32
    i = 0
    for _, v := range payload {
        if (i % 16) == 0 {
            fmt.Printf("\n")
        } else if (i % 8) == 0 {
            fmt.Printf(" ")
        }
        fmt.Printf("%02X ", v)
        i++
    }
    fmt.Printf("\n\n")
}

func is_checksum_valid(checksum uint32, payload []byte) bool {
    if message_checksum(payload) == checksum {
        return true
    }
    return false
}

func process_rx_queue() {
    for {
        if(RX_FIFO.Len() > 0) {
            var node *message_fifo.NODE
            node = RX_FIFO.Pop()
            header := deserialize_message_header(node.Message[:MESSAGE_HEADER_LENGTH])
            print_message_payload_hex(node.Message[MESSAGE_HEADER_LENGTH:], header.Length)
            switch(strings.TrimRight(string(header.Command[:]), "\x00")) {
            case "version":
                var node message_fifo.NODE
                node.Message = build_verack_message(MAINNET_MAGIC)
                TX_FIFO.Push(&node)
                break;
            case "verack":
                // Do nothing
                break;
            default:
                break;
            }
        }
    }
}

func tx_message_handler(conn net.Conn) {
    //fmt.Println("Handling new connection...")

    if(TX_FIFO.Len() > 0) {
        var node *message_fifo.NODE
        node = TX_FIFO.Pop()

        conn.Write(node.Message)
    }
}

func rx_message_handler(conn net.Conn) {
    //fmt.Println("Handling new connection...")

    // Close connection when this function ends
    defer func() {
        //fmt.Println("Closing connection...")
        conn.Close()
    }()

    timeout_duration := 5 * time.Second
    buf_reader := bufio.NewReader(conn)
    for {
		// Set a deadline for reading. Read operation will fail if no data
		// is received after deadline.
		conn.SetReadDeadline(time.Now().Add(timeout_duration))

        // Read message header first, so we know how much to read out of the payload
        header_raw := make([]byte, MESSAGE_HEADER_LENGTH)
		n, err := buf_reader.Read(header_raw)
		if err != nil || uint32(n) != MESSAGE_HEADER_LENGTH {
			//fmt.Println(err)
			return
		}
        header_struct := deserialize_message_header(header_raw)
        payload_raw := make([]byte, header_struct.Length)
        n, err = buf_reader.Read(payload_raw)
		if err != nil || uint32(n) != header_struct.Length {
			//fmt.Println(err)
			return
		}

        // Ensure checksum is valid
        if(is_checksum_valid(header_struct.Checksum, payload_raw)) {
            // Add to message queue
            var node message_fifo.NODE
            node.Message = append(header_raw, payload_raw...)
            RX_FIFO.Push(&node)
        }
	}
}

func main() {
    //fmt.Println("Send version message:", version_message)
    RX_FIFO = message_fifo.GENERIC_New(RX_FIFO_SIZE)
    TX_FIFO = message_fifo.GENERIC_New(TX_FIFO_SIZE)

    go process_rx_queue()

    // Send version message once
    var node message_fifo.NODE
    node.Message = build_version_message(MAINNET_MAGIC, "", 0)
    TX_FIFO.Push(&node)

    // connect to this socket
    for {
        conn, err := net.Dial("tcp", "52.41.9.64:8333")
        if err != nil {
            fmt.Println(err)
        }

        go tx_message_handler(conn)
        go rx_message_handler(conn)
    }
}
