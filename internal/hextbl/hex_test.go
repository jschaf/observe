package hextbl_test

import (
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
			name:   "zero",
			in:     "0000000000000000",
			want:   0,
			wantOk: true,
		},
		{
			name:   "filled",
			in:     "0123456789abcdef",
			want:   0x0123456789ABCDEF,
			wantOk: true,
		},
		{
			name:   "filled uppercase",
			in:     "0123456789ABCDEF",
			want:   0x0123456789ABCDEF,
			wantOk: true,
		},

		// Invalid
		{
			name:   "invalid length short",
			in:     "abc",
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
			got, ok := hextbl.ParseUint64(tt.in)
			if got != tt.want {
				t.Errorf("ParseUint64(%q) = %v, want %v", tt.in, got, tt.want)
			}
			if ok != tt.wantOk {
				t.Errorf("ParseUint64(%q).ok = %v, want %v", tt.in, ok, tt.wantOk)
			}
		})
	}
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
