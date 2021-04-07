package utils

func BytesToUint16(a, b byte) (v uint16) {
	return uint16(a)<<8 | uint16(b)
}

func Uint16ToBytes(v uint16) []byte {
	return []byte{
		uint8((v >> 8) & 0xFF),
		uint8(v & 0x00FF),
	}
}
