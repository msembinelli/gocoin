package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

var MESSAGE_HEADER_LENGTH uint32 = 24

// Generic protocol message header
type Header struct {
	Magic    uint32
	Command  [12]byte
	Length   uint32
	Checksum uint32
}

func (h Header) Serialize() []byte {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, h)
	if err != nil {
		fmt.Println(err)
		// Return empty buffer
		return new(bytes.Buffer).Bytes()
	}

	return buffer.Bytes()
}
