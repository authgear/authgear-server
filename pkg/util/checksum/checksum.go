package checksum

import (
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
)

func CRC32IEEEInHex(data []byte) string {
	// Calculate the checksum with crc32 IEEE
	checksum := crc32.ChecksumIEEE(data)
	// Turn the 32-bit unsigned checksum into 4 bytes in big endian order.
	byteSlice := make([]byte, 4)
	byteSlice = binary.BigEndian.AppendUint32(byteSlice, checksum)

	// Encode the 4 bytes in hex format.
	checksumString := hex.EncodeToString(byteSlice)

	return checksumString
}
