package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Pong struct {
	Nonce uint64
}

func (p Pong) Serialize() []byte {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, p)
	if err != nil {
		fmt.Println(err)
		// Return empty buffer
		return new(bytes.Buffer).Bytes()
	}

	return buffer.Bytes()
}
