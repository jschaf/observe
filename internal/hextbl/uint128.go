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
func ParseUint128(s string) Uint128 {
	if len(s) != 32 {
		return Uint128{}
	}

	hi0 := reverseDigit0[s[0x0]] | reverseDigit1[s[0x1]] | reverseDigit2[s[0x2]] | reverseDigit3[s[0x3]]
	hi1 := reverseDigit0[s[0x4]] | reverseDigit1[s[0x5]] | reverseDigit2[s[0x6]] | reverseDigit3[s[0x7]]
	hi2 := reverseDigit0[s[0x8]] | reverseDigit1[s[0x9]] | reverseDigit2[s[0xa]] | reverseDigit3[s[0xb]]
	hi3 := reverseDigit0[s[0xc]] | reverseDigit1[s[0xd]] | reverseDigit2[s[0xe]] | reverseDigit3[s[0xf]]

	lo0 := reverseDigit0[s[0x10]] | reverseDigit1[s[0x11]] | reverseDigit2[s[0x12]] | reverseDigit3[s[0x13]]
	lo1 := reverseDigit0[s[0x14]] | reverseDigit1[s[0x15]] | reverseDigit2[s[0x16]] | reverseDigit3[s[0x17]]
	lo2 := reverseDigit0[s[0x18]] | reverseDigit1[s[0x19]] | reverseDigit2[s[0x1a]] | reverseDigit3[s[0x1b]]
	lo3 := reverseDigit0[s[0x1c]] | reverseDigit1[s[0x1d]] | reverseDigit2[s[0x1e]] | reverseDigit3[s[0x1f]]

	hiInvalid := (hi0&invalidReverse | hi1&invalidReverse | hi2&invalidReverse | hi3&invalidReverse) == invalidReverse
	loInvalid := (lo0&invalidReverse | lo1&invalidReverse | lo2&invalidReverse | lo3&invalidReverse) == invalidReverse
	if hiInvalid || loInvalid {
		return Uint128{}
	}

	hi := (hi0 << 48) | (hi1 << 32) | (hi2 << 16) | hi3
	lo := (lo0 << 48) | (lo1 << 32) | (lo2 << 16) | lo3
	return Uint128{Hi: hi, Lo: lo}
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
