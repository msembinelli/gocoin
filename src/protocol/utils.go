package protocol

import (
	"crypto/sha256"
	"encoding/binary"
)

// Calculate message checksum. First 4 bytes of sha256(sha256(payload))
func MessageChecksum(payload []byte) uint32 {
	hash := sha256.New()
	hash.Write(payload)
	sum := hash.Sum(nil)
	hash.Reset()
	hash.Write(sum)
	sum = hash.Sum(nil)
	return binary.LittleEndian.Uint32(sum[:4])
}
