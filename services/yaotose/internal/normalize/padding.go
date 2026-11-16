package normalize

func PadTo32(s string) string {
	var k [32]byte
	copy(k[:], s)
	return string(k[:])
}
