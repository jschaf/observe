package trace

import (
	"strings"
	"testing"

	"github.com/jschaf/observe/internal/difftest"
)

func TestKindString(t *testing.T) {
	if got, want := kindString.String(), "String"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := kindBoolSlice.String(), "BoolSlice"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := kindInt64Slice.String(), "Int64Slice"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := kindFloat64Slice.String(), "Float64Slice"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := kindStringSlice.String(), "StringSlice"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValueString(t *testing.T) {
	for _, test := range []struct {
		v    Value
		want string
	}{
		{boolValue(true), "true"},
		{float64Value(.15), "0.15"},
		{int64Value(-3), "-3"},
		{stringValue("foo"), "foo"},
		{boolsValue([]bool{true, false, true, false, false}), "[true,false,true,false,false]"},
	} {
		if got := test.v.String(); got != test.want {
			t.Errorf("%#v:\ngot  %q\nwant %q", test.v, got, test.want)
		}
	}
}

func TestValue_BoolsTruncated(t *testing.T) {
	bs := make([]bool, 100)
	val := boolsValue(bs)
	want := "[" + strings.Repeat("false,", 55) + "false]"
	got := val.String()
	difftest.AssertSame(t, "BoolsTruncated", got, want)
}

func TestValueNoAlloc(t *testing.T) {
	// Assign values just to make sure the compiler doesn't optimize away the
	// statements.
	var (
		i int64
		f float64
		b bool
		s string
	)
	a := int(testing.AllocsPerRun(5, func() {
		i = int64Value(1).Int64()
		f = float64Value(1).Float64()
		b = boolValue(true).Bool()
		s = stringValue("foo").String()
	}))
	if a != 0 {
		t.Errorf("got %d allocs, want zero", a)
	}
	_ = i
	_ = f
	_ = b
	_ = s
}
