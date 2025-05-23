package hextbl

import "math"

//goland:noinspection SpellCheckingInspection
const (
	Lookup  = "0123456789abcdef"
	Reverse = "" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\xff\xff\xff\xff\xff\xff" +
		"\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff" +
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

const invalidReverse = math.MaxUint64

func shiftReverseTable(shift int) [256]uint64 {
	var table [256]uint64
	for i := range Reverse {
		if Reverse[i] == 0xff {
			table[i] = invalidReverse
			continue
		}
		table[i] = uint64(Reverse[i]) << shift
	}
	return table
}

// reverseDigit0 is the first digit in a 4-byte hex string.
var reverseDigit0 = shiftReverseTable(12) //nolint:gochecknoglobals

// reverseDigit1 is the second digit in a 4-byte hex string.
var reverseDigit1 = shiftReverseTable(8) //nolint:gochecknoglobals

// reverseDigit2 is the third digit in a 4-byte hex string.
var reverseDigit2 = shiftReverseTable(4) //nolint:gochecknoglobals

// reverseDigit3 is the last digit in a 4-byte hex string.
var reverseDigit3 = shiftReverseTable(0) //nolint:gochecknoglobals

// ReadByte reads two hex characters and returns the byte value, and if the hex
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

// ParseUint64 parses a 16-byte hex string into a uint64. Returns 0 if the
// string is not valid. All zeros are not a valid Span ID.
func ParseUint64(s string) uint64 {
	if len(s) != 16 {
		return 0
	}

	n0 := reverseDigit0[s[0x0]] | reverseDigit1[s[0x1]] | reverseDigit2[s[0x2]] | reverseDigit3[s[0x3]]
	n1 := reverseDigit0[s[0x4]] | reverseDigit1[s[0x5]] | reverseDigit2[s[0x6]] | reverseDigit3[s[0x7]]
	n2 := reverseDigit0[s[0x8]] | reverseDigit1[s[0x9]] | reverseDigit2[s[0xa]] | reverseDigit3[s[0xb]]
	n3 := reverseDigit0[s[0xc]] | reverseDigit1[s[0xd]] | reverseDigit2[s[0xe]] | reverseDigit3[s[0xf]]

	invalid := (n0&invalidReverse | n1&invalidReverse | n2&invalidReverse | n3&invalidReverse) == invalidReverse
	if invalid {
		return 0
	}
	return (n0 << 48) | (n1 << 32) | (n2 << 16) | n3
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
