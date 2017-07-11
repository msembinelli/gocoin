package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

type Version struct {
	Version     int32
	Services    uint64
	Timestamp   int64
	AddrRecv    NetAddrNoTime
	AddrFrom    NetAddrNoTime
	Nonce       uint64
	UserAgent   [1]byte
	StartHeight int32
	Relay       bool
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
