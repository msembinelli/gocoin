package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// version message requires a network address struct with no time field
type NetAddrNoTime struct {
	Services uint64
	Ip       [16]byte
	Port     uint16
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
