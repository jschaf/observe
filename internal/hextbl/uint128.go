package hextbl

import (
	"fmt"
)

// Uint128 represents a Uint128 using two uint64s.
//
// When the methods below mention a bit, bit 0 is the most significant bit
// (in Hi) and bit 127 is the lowest (Lo&1).
type Uint128 struct {
	Hi uint64
	Lo uint64
}

// ParseUint128 parses a 32-byte hex string to Uint128.
func ParseUint128(s string) (Uint128, bool) {
	if len(s) != 32 {
		return Uint128{}, false
	}
	var hi, lo uint64
	invalidMark := byte(0)
	for i := 0; i < 16; i += 4 {
		shift := uint((15 - i) * 4) //nolint:gosec

		hi |= uint64(Reverse[s[i]]) << shift
		hi |= uint64(Reverse[s[i+1]]) << (shift - 4)
		hi |= uint64(Reverse[s[i+2]]) << (shift - 8)
		hi |= uint64(Reverse[s[i+3]]) << (shift - 12)

		lo |= uint64(Reverse[s[16+i]]) << shift
		lo |= uint64(Reverse[s[16+i+1]]) << (shift - 4)
		lo |= uint64(Reverse[s[16+i+2]]) << (shift - 8)
		lo |= uint64(Reverse[s[16+i+3]]) << (shift - 12)

		invalidMark |= Reverse[s[i]] | Reverse[s[i+1]] | Reverse[s[i+2]] | Reverse[s[i+3]]
		invalidMark |= Reverse[s[16+i]] | Reverse[s[16+i+1]] | Reverse[s[16+i+2]] | Reverse[s[16+i+3]]
	}
	if invalidMark&0xf0 != 0 {
		return Uint128{}, false
	}
	return Uint128{Hi: hi, Lo: lo}, true
}

// IsZero reports whether u == 0.
//
// It's faster than u == (Uint128{}) because the compiler (as of Go
// 1.15/1.16b1) doesn't do this trick and instead inserts a branch in
// its eq alg's generated code.
func (u Uint128) IsZero() bool { return u.Hi|u.Lo == 0 }

// Bytes returns the hex bytes of a Uint128.
func (u Uint128) Bytes() [32]byte {
	return [32]byte{
		Lookup[(u.Hi>>0o74)&0xf], Lookup[(u.Hi>>0o70)&0xf],
		Lookup[(u.Hi>>0o64)&0xf], Lookup[(u.Hi>>0o60)&0xf],
		Lookup[(u.Hi>>0o54)&0xf], Lookup[(u.Hi>>0o50)&0xf],
		Lookup[(u.Hi>>0o44)&0xf], Lookup[(u.Hi>>0o40)&0xf],
		Lookup[(u.Hi>>0o34)&0xf], Lookup[(u.Hi>>0o30)&0xf],
		Lookup[(u.Hi>>0o24)&0xf], Lookup[(u.Hi>>0o20)&0xf],
		Lookup[(u.Hi>>0o14)&0xf], Lookup[(u.Hi>>0o10)&0xf],
		Lookup[(u.Hi>>0o04)&0xf], Lookup[(u.Hi>>0o00)&0xf],

		Lookup[(u.Lo>>0o74)&0xf], Lookup[(u.Lo>>0o70)&0xf],
		Lookup[(u.Lo>>0o64)&0xf], Lookup[(u.Lo>>0o60)&0xf],
		Lookup[(u.Lo>>0o54)&0xf], Lookup[(u.Lo>>0o50)&0xf],
		Lookup[(u.Lo>>0o44)&0xf], Lookup[(u.Lo>>0o40)&0xf],
		Lookup[(u.Lo>>0o34)&0xf], Lookup[(u.Lo>>0o30)&0xf],
		Lookup[(u.Lo>>0o24)&0xf], Lookup[(u.Lo>>0o20)&0xf],
		Lookup[(u.Lo>>0o14)&0xf], Lookup[(u.Lo>>0o10)&0xf],
		Lookup[(u.Lo>>0o04)&0xf], Lookup[(u.Lo>>0o00)&0xf],
	}
}

func (u Uint128) GoString() string {
	return fmt.Sprintf("Uint128{Hi: %x, Lo: %x}", u.Hi, u.Lo)
}
