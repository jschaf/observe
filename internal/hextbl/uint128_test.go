package hextbl_test

import (
	"testing"

	"github.com/jschaf/observe/internal/hextbl"
)

func TestParseUint128(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   hextbl.Uint128
		wantOk bool
	}{
		{
			name:   "zero",
			in:     "00000000000000000000000000000000",
			want:   hextbl.Uint128{},
			wantOk: true,
		},
		{
			name:   "non-zero hi",
			in:     "0123456789abcdef0000000000000000",
			want:   hextbl.Uint128{Hi: 0x0123456789ABCDEF},
			wantOk: true,
		},
		{
			name:   "non-zero lo",
			in:     "00000000000000000123456789abcdef",
			want:   hextbl.Uint128{Lo: 0x0123456789ABCDEF},
			wantOk: true,
		},
		{
			name:   "filled",
			in:     "0123456789abcdef0123456789abcdef",
			want:   hextbl.Uint128{Hi: 0x0123456789abcdef, Lo: 0x0123456789abcdef},
			wantOk: true,
		},

		// Invalid
		{
			name:   "invalid length short",
			in:     "abc",
			want:   hextbl.Uint128{},
			wantOk: false,
		},
		{
			name:   "invalid length long",
			in:     "0123456789abcdef0123456789abcdef1",
			want:   hextbl.Uint128{},
			wantOk: false,
		},
		{
			name:   "invalid char",
			in:     "z123456789abcdef0123456789abcdef",
			want:   hextbl.Uint128{},
			wantOk: false,
		},
		{
			name:   "filled uppercase",
			in:     "0123456789ABCDEF0123456789ABCDEF",
			want:   hextbl.Uint128{Hi: 0, Lo: 0},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := hextbl.ParseUint128(tt.in)
			if got != tt.want {
				t.Errorf("ParseUint128(%q) = %v, want %v", tt.in, got, tt.want)
			}
			if ok != tt.wantOk {
				t.Errorf("ParseUint128(%q) = %v, want %v", tt.in, ok, tt.wantOk)
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
