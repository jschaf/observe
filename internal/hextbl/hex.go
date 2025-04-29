package hextbl

//goland:noinspection SpellCheckingInspection
const (
	Lookup  = "0123456789abcdef"
	Reverse = "" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\x0a\x0b\x0c\x0d\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
)

// ReadByte reads two hex characters and returns the byte value and if the hex
// characters were valid.
func ReadByte(a, b byte) (byte, bool) {
	x := Reverse[a]
	y := Reverse[b]
	isValid := x|y != 0xff
	if !isValid {
		return 0, false
	}
	return (x << 4) | y, true
}

// ParseUint64 parses a 16-byte hex string into a uint64.
func ParseUint64(s string) (uint64, bool) {
	if len(s) != 16 {
		return 0, false
	}
	var n uint64
	invalidMark := byte(0)
	for i := 0; i < len(s); i += 4 {
		shift := uint((15 - i) * 4) //nolint:gosec
		n |= uint64(Reverse[s[i]]) << shift
		n |= uint64(Reverse[s[i+1]]) << (shift - 4)
		n |= uint64(Reverse[s[i+2]]) << (shift - 8)
		n |= uint64(Reverse[s[i+3]]) << (shift - 12)
		invalidMark |= Reverse[s[i]] | Reverse[s[i+1]] | Reverse[s[i+2]] | Reverse[s[i+3]]
	}
	if invalidMark&0xf0 != 0 {
		return 0, false
	}
	return n, true
}

// Uint64Bytes converts a uint64 to a 16-byte hex array.
func Uint64Bytes(n uint64) [16]byte {
	return [16]byte{
		Lookup[(n>>0o74)&0xf], Lookup[(n>>0o70)&0xf],
		Lookup[(n>>0o64)&0xf], Lookup[(n>>0o60)&0xf],
		Lookup[(n>>0o54)&0xf], Lookup[(n>>0o50)&0xf],
		Lookup[(n>>0o44)&0xf], Lookup[(n>>0o40)&0xf],
		Lookup[(n>>0o34)&0xf], Lookup[(n>>0o30)&0xf],
		Lookup[(n>>0o24)&0xf], Lookup[(n>>0o20)&0xf],
		Lookup[(n>>0o14)&0xf], Lookup[(n>>0o10)&0xf],
		Lookup[(n>>0o04)&0xf], Lookup[(n>>0o00)&0xf],
	}
}
