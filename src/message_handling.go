package main

import (
	"./message_fifo"
	"./messages"
	"./protocol"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

//"time"

var TX_FIFO_SIZE uint32 = 32
var RX_FIFO_SIZE uint32 = 32
var TX_FIFO *message_fifo.FIFO
var RX_FIFO *message_fifo.FIFO

func print_message_header(h messages.Header) {
	fmt.Printf("**MESSAGE HEADER**\n")
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

func deserialize_message_header(received []byte) messages.Header {
	var h messages.Header
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
	if protocol.MessageChecksum(payload) == checksum {
		return true
	}
	return false
}

func process_rx_queue() {
	for {
		if RX_FIFO.Len() > 0 {
			var node *message_fifo.NODE
			node = RX_FIFO.Pop()
			header := deserialize_message_header(node.Message[:messages.MESSAGE_HEADER_LENGTH])
			print_message_payload_hex(node.Message[messages.MESSAGE_HEADER_LENGTH:], header.Length)
			switch strings.TrimRight(string(header.Command[:]), "\x00") {
			case "version":
				var node message_fifo.NODE
				node.Message = messages.BuildVerackMessage(protocol.MAINNET_MAGIC)
				TX_FIFO.Push(&node)
				break
			case "verack":
				var node message_fifo.NODE
				node.Message = messages.BuildGetaddrMessage(protocol.MAINNET_MAGIC)
				TX_FIFO.Push(&node)
				break
			case "ping":
				var node message_fifo.NODE
				node.Message = messages.BuildPongMessage(protocol.MAINNET_MAGIC)
				TX_FIFO.Push(&node)
			default:
				break
			}
		}
	}
}

func tx_message_handler(conn net.Conn) {
	for {
		if TX_FIFO.Len() > 0 {
			var node *message_fifo.NODE
			node = TX_FIFO.Pop()

			conn.Write(node.Message)
		} else {
			break
		}
	}
}

func rx_message_handler(conn net.Conn) {
	fmt.Println("RX Handling new connection...")

	// Close connection when this function ends
	defer func() {
		fmt.Println("RX Closing connection...")
		conn.Close()
	}()

	//timeout_duration := 30 * time.Second
	buf_reader := bufio.NewReader(conn)
	for {
		// Set a deadline for reading. Read operation will fail if no data
		// is received after deadline.
		//conn.SetReadDeadline(time.Now().Add(timeout_duration))

		// Read message header first, so we know how much to read out of the payload
		header_raw := make([]byte, messages.MESSAGE_HEADER_LENGTH)
		n, err := buf_reader.Read(header_raw)
		fmt.Println("header read size", n)
		if err != nil {
			fmt.Println(err)
			return
		}
		header_struct := deserialize_message_header(header_raw)
		payload_raw := make([]byte, header_struct.Length)
		n, err = buf_reader.Read(payload_raw)
		fmt.Println("payload read size", n)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Ensure checksum is valid
		if is_checksum_valid(header_struct.Checksum, payload_raw) {
			// Add to message queue
			fmt.Println("Adding message to queue!")
			var node message_fifo.NODE
			node.Message = append(header_raw, payload_raw...)
			RX_FIFO.Push(&node)
		}
		buf_reader.Reset(conn)
	}
}

func main() {
	//fmt.Println("Send version message:", version_message)
	RX_FIFO = message_fifo.GENERIC_New(RX_FIFO_SIZE)
	TX_FIFO = message_fifo.GENERIC_New(TX_FIFO_SIZE)

	go process_rx_queue()

	// Send version message once
	var node message_fifo.NODE
	node.Message = messages.BuildVersionMessage(protocol.MAINNET_MAGIC, "", 0)
	TX_FIFO.Push(&node)

	conn, err := net.Dial("tcp", "198.37.201.59:8333")
	if err != nil {
		fmt.Println(err)
	}
	go rx_message_handler(conn)

	// connect to this socket
	for {
		tx_message_handler(conn)
	}
}
