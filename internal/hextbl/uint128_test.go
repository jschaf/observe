package hextbl_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jschaf/observe/internal/hextbl"
)

func TestParseUint128(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want hextbl.Uint128
	}{
		{
			name: "zero",
			in:   "00000000000000000000000000000000",
			want: hextbl.Uint128{},
		},
		{
			name: "non-zero hi",
			in:   "0123456789abcdef0000000000000000",
			want: hextbl.Uint128{Hi: 0x0123456789ABCDEF},
		},
		{
			name: "non-zero lo",
			in:   "00000000000000000123456789abcdef",
			want: hextbl.Uint128{Lo: 0x0123456789ABCDEF},
		},
		{
			name: "filled",
			in:   "0123456789abcdef0123456789abcdef",
			want: hextbl.Uint128{Hi: 0x0123456789abcdef, Lo: 0x0123456789abcdef},
		},

		// Invalid
		{
			name: "invalid length short",
			in:   "abc",
			want: hextbl.Uint128{},
		},
		{
			name: "invalid length long",
			in:   "0123456789abcdef0123456789abcdef1",
			want: hextbl.Uint128{},
		},
		{
			name: "invalid char",
			in:   "z123456789abcdef0123456789abcdef",
			want: hextbl.Uint128{},
		},
		{
			name: "filled uppercase",
			in:   "0123456789ABCDEF0123456789ABCDEF",
			want: hextbl.Uint128{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hextbl.ParseUint128(tt.in)
			if got != tt.want {
				t.Errorf("ParseUint128(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestUint128_IsZero(t *testing.T) {
	tests := []struct {
		name string
		u    hextbl.Uint128
		want bool
	}{
		{name: "zero", u: hextbl.Uint128{}, want: true},
		{name: "non-zero hi", u: hextbl.Uint128{Hi: 1}, want: false},
		{name: "non-zero lo", u: hextbl.Uint128{Lo: 1}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.IsZero(); got != tt.want {
				t.Errorf("%#v.IsZero() = %v, want %v", tt.u, got, tt.want)
			}
		})
	}
}

func TestUint128_Bytes(t *testing.T) {
	tests := []struct {
		name string
		u    hextbl.Uint128
		want string
	}{
		{
			name: "zero",
			u:    hextbl.Uint128{},
			want: "00000000000000000000000000000000",
		},
		{
			name: "non-zero hi",
			u:    hextbl.Uint128{Hi: 0x0123456789ABCDEF},
			want: "0123456789abcdef0000000000000000",
		},
		{
			name: "non-zero lo",
			u:    hextbl.Uint128{Lo: 0x0123456789ABCDEF},
			want: "00000000000000000123456789abcdef",
		},
		{
			name: "filled",
			u:    hextbl.Uint128{Hi: 0x0123456789abcdef, Lo: 0x0123456789abcdef},
			want: "0123456789abcdef0123456789abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.u.Bytes()
			gotStr := string(got[:])
			if gotStr != tt.want {
				t.Errorf("%#v.Bytes() = %v, want %v", tt.u, got, tt.want)
			}
		})
	}
}

func FuzzParseUint128(f *testing.F) {
	f.Add("0123456789abcdef")
	f.Add(strings.Repeat("0", 16))
	f.Add(strings.Repeat("f", 16))
	f.Fuzz(func(t *testing.T, s string) {
		slow := slowParseUint128(s)
		fast := hextbl.ParseUint128(s)
		if slow != fast {
			t.Errorf("ParseUint128(%q) = %v, want %v", s, fast, slow)
		}
	})
}

func slowParseUint128(s string) hextbl.Uint128 {
	if len(s) != 32 {
		return hextbl.Uint128{}
	}
	// Return 0 for uppercase hex.
	for _, ch := range s {
		if ch >= 'A' && ch <= 'F' {
			return hextbl.Uint128{}
		}
	}
	bs, err := hex.DecodeString(s)
	if err != nil {
		return hextbl.Uint128{}
	}
	// Convert to uint64.
	var hi uint64
	for _, b := range bs[:8] {
		hi <<= 8
		hi |= uint64(b)
	}
	var lo uint64
	for _, b := range bs[8:] {
		lo <<= 8
		lo |= uint64(b)
	}
	return hextbl.Uint128{Hi: hi, Lo: lo}
}

// Results as of 2025-05-23.
//
//	BenchmarkParseUint12/ParseUint128      168090454	  7.10 ns/op   0 B/op  0 allocs/op
//	BenchmarkParseUint12/hex.DecodeString   44913612  25.40 ns/op  16 B/op  1 allocs/op
func BenchmarkParseUint12(b *testing.B) {
	b.Run("ParseUint128", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			n := hextbl.ParseUint128("0123456789abcdef0123456789abcdef")
			if n != (hextbl.Uint128{Hi: 0x0123456789abcdef, Lo: 0x0123456789abcdef}) {
				b.Fatalf("ParseUint128() = %v, want %v", n, "0x0123456789abcdef0123456789abcdef")
			}
		}
	})
	b.Run("hex.DecodeString", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, err := hex.DecodeString("0123456789abcdef0123456789abcdef")
			if err != nil {
				b.Fatalf("decode hex: %v", err)
			}
		}
	})
}
