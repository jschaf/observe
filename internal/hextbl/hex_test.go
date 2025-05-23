package hextbl_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/internal/hextbl"
)

func TestReadByte(t *testing.T) {
	tests := []struct {
		name   string
		a      byte
		b      byte
		want   byte
		wantOk bool
	}{
		// Ok
		{name: "zero", a: '0', b: '0', want: 0, wantOk: true},
		{name: "ab", a: 'a', b: 'b', want: 0xab, wantOk: true},

		// Not ok
		{name: "invalid hex", a: 'g', b: '0', want: 0, wantOk: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := hextbl.ReadByte(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ReadByte() got = %v, want %v", got, tt.want)
			}
			if ok != tt.wantOk {
				t.Errorf("ReadByte() ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestParseUint64(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   uint64
		wantOk bool
	}{
		{
			name:   "filled",
			in:     "0123456789abcdef",
			want:   0x0123456789ABCDEF,
			wantOk: true,
		},
		{
			name:   "single max byte",
			in:     "ff00000000000000",
			want:   0xFF00000000000000,
			wantOk: true,
		},

		// Invalid
		{
			name:   "zero",
			in:     "0000000000000000",
			want:   0,
			wantOk: false,
		},
		{
			name:   "invalid length short",
			in:     "abc",
			want:   0,
			wantOk: false,
		},
		{
			name:   "filled uppercase",
			in:     "0123456789ABCDEF",
			want:   0,
			wantOk: false,
		},
		{
			name:   "single uppercase",
			in:     "0000A00000000000",
			want:   0,
			wantOk: false,
		},
		{
			name:   "two uppercase",
			in:     "FF00000000000000",
			want:   0,
			wantOk: false,
		},
		{
			name:   "invalid length long",
			in:     "0123456789abcdef0",
			want:   0,
			wantOk: false,
		},
		{
			name:   "invalid char",
			in:     "z123456789abcdef",
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hextbl.ParseUint64(tt.in)
			if got != tt.want {
				t.Errorf("ParseUint64Table(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func FuzzParseUint64(f *testing.F) {
	f.Add("0123456789abcdef")
	f.Add(strings.Repeat("0", 16))
	f.Add(strings.Repeat("f", 16))
	f.Fuzz(func(t *testing.T, s string) {
		slow := slowParseUint64(s)
		fast := hextbl.ParseUint64(s)
		if slow != fast {
			t.Errorf("ParseUint64(%q) = %x, want %x", s, fast, slow)
		}
	})
}

func slowParseUint64(s string) uint64 {
	if len(s) != 16 {
		return 0
	}
	// Return false for uppercase hex.
	for _, ch := range s {
		if ch >= 'A' && ch <= 'F' {
			return 0
		}
	}
	bs, err := hex.DecodeString(s)
	if err != nil {
		return 0
	}
	if len(bs) != 8 {
		return 0
	}
	// Convert to uint64.
	var n uint64
	for _, b := range bs {
		n <<= 8
		n |= uint64(b)
	}
	return n
}

func TestUint64Bytes(t *testing.T) {
	tests := []struct {
		name string
		u    uint64
		want string
	}{
		{
			name: "zero",
			u:    0,
			want: "0000000000000000",
		},
		{
			name: "non-zero hi",
			u:    0x0123456789ABCDEF,
			want: "0123456789abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hextbl.Uint64Bytes(tt.u)
			gotStr := string(got[:])
			difftest.AssertSame(t, "Uint64Bytes mismatch", tt.want, gotStr)
		})
	}
}

// Results as of 2025-05-23.
//
//	BenchmarkParseUint64/ParseUint64      318850330   3.68 ns/op  0 B/op  0 allocs/op
//	BenchmarkParseUint64/hex.DecodeString  69341446  16.55 ns/op  8 B/op  1 allocs/op
func BenchmarkParseUint64(b *testing.B) {
	b.Run("ParseUint64", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			n := hextbl.ParseUint64("0123456789abcdef")
			if n != 0x0123456789ABCDEF {
				b.Fatalf("ParseUint64Table() = %v, want %v", n, 0x0123456789ABCDEF)
			}
		}
	})
	b.Run("hex.DecodeString", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_, err := hex.DecodeString("0123456789abcdef")
			if err != nil {
				b.Fatalf("ParseUint64() invalid")
			}
		}
	})
}
