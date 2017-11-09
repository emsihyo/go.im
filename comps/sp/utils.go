package sp

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"
	"io"
)

//SlotBalance SlotBalance
func SlotBalance(str string, slotCount uint32) uint32 {
	return CRC32(MD5(str)) % slotCount
}

//CRC32 CRC32
func CRC32(str string) uint32 {
	ieee := crc32.NewIEEE()
	io.WriteString(ieee, str)
	return ieee.Sum32()
}

//MD5 MD5
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
