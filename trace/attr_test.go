package trace

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/jschaf/observe/internal/difftest"
)

func TestAttr_RoundTrip(t *testing.T) {
	assertRoundTrip(t, "bool", true)
	assertRoundTrip(t, "bool", false)

	assertRoundTrip(t, "float64", -1999999999.0)
	assertRoundTrip(t, "float64", -1.0)
	assertRoundTrip(t, "float64", 0.0)
	assertRoundTrip(t, "float64", math.MaxFloat64)

	assertRoundTrip(t, "int", math.MaxInt)
	assertRoundTrip(t, "int", 0)
	assertRoundTrip(t, "int", -1)

	assertRoundTrip(t, "int64", int64(math.MaxInt))

	assertRoundTrip(t, "string", "")
	assertRoundTrip(t, "string", "string-value")

	assertRoundTrip(t, "bools", []bool{})
	assertRoundTrip(t, "bools", []bool{true})
	assertRoundTrip(t, "bools", []bool{false})
	assertRoundTrip(t, "bools", []bool{true, false, true, false, false})

	assertRoundTrip(t, "float64s", []float64{})
	assertRoundTrip(t, "float64s", []float64{1.0})
	assertRoundTrip(t, "float64s", []float64{1.0, 2.0, 3.0, 4.0, 5.0})
	assertRoundTrip(t, "float64s", slices.Repeat([]float64{50.0}, 1024))

	assertRoundTrip(t, "ints", []int{})
	assertRoundTrip(t, "ints", []int{1})
	assertRoundTrip(t, "ints", []int{1, 2, 3, 4, 5})
	assertRoundTrip(t, "ints", slices.Repeat([]int{50}, 1024))

	assertRoundTrip(t, "int64s", []int64{})
	assertRoundTrip(t, "int64s", []int64{1})
	assertRoundTrip(t, "int64s", []int64{1, 2, 3, 4, 5})
	assertRoundTrip(t, "int64s", slices.Repeat([]int64{50}, 1024))

	assertRoundTrip(t, "strings", []string{})
	assertRoundTrip(t, "strings", []string{"foo"})
	assertRoundTrip(t, "strings", []string{"foo", "bar", "baz"})
	assertRoundTrip(t, "strings", slices.Repeat([]string{"foo"}, 1024))
}

func assertRoundTrip(t *testing.T, key string, val any) {
	t.Helper()
	var attr Attr
	var wantValStr string
	switch v := val.(type) {
	case bool:
		attr = Bool(key, v)
		difftest.AssertSame(t, "round trip value bool", v, attr.Value.Bool())
	case int:
		attr = Int(key, v)
		difftest.AssertSame(t, "round trip value int", int64(v), attr.Value.Int64())
	case int64:
		attr = Int64(key, v)
		difftest.AssertSame(t, "round trip value int64", v, attr.Value.Int64())
	case float64:
		attr = Float64(key, v)
		difftest.AssertSame(t, "round trip value float64", v, attr.Value.Float64())
	case string:
		attr = String(key, v)
		difftest.AssertSame(t, "round trip value string", v, attr.Value.String())
	case []bool:
		attr = Bools(key, v)
		out, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal bool slice %v: %v", v, err)
		}
		wantValStr = string(out)
		difftest.AssertSame(t, "round trip value bools", v, slices.Collect(attr.Value.Bools()))
	case []float64:
		attr = Float64s(key, v)
		out, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal float64 slice %v: %v", v, err)
		}
		wantValStr = string(out)
		difftest.AssertSame(t, "round trip value float64s", v, slices.Collect(attr.Value.Float64s()))
	case []int:
		attr = Ints(key, v)
		out, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal int64 slice %v: %v", v, err)
		}
		wantValStr = string(out)
		difftest.AssertSame(t, "round trip value ints", v, slices.Collect(attr.Value.Ints()))
	case []int64:
		attr = Int64s(key, v)
		out, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal int64 slice %v: %v", v, err)
		}
		wantValStr = string(out)
		difftest.AssertSame(t, "round trip value int64s", v, slices.Collect(attr.Value.Int64s()))
	case []string:
		attr = Strings(key, v)
		out, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal string slice %v: %v", v, err)
		}
		wantValStr = string(out)
		difftest.AssertSame(t, "round trip value strings", v, slices.Collect(attr.Value.Strings()))
	default:
		t.Fatalf("unsupported type %T for key %s", v, key)
	}
	msg := fmt.Sprintf("round trip %s=%v", key, val)
	if wantValStr == "" {
		wantValStr = fmt.Sprint(val)
	}
	wantAttrStr := key + "=" + wantValStr
	gotAttrStr := attr.String()
	difftest.AssertSame(t, msg, wantAttrStr, gotAttrStr)
}

func TestAttr_NoAlloc(t *testing.T) {
	// Assign values just to make sure the compiler doesn't optimize away the
	// statements.
	var (
		b bool
		f float64
		i int64
		s string
	)
	a := int(testing.AllocsPerRun(5, func() {
		b = Bool("key", true).Value.Bool()
		f = Float64("key", 1).Value.Float64()
		i = Int64("key", 1).Value.Int64()
		i = Int("key", 1).Value.Int64()
		s = String("key", "foo").Value.String()
	}))
	if a != 0 {
		t.Errorf("got %d allocs, want zero", a)
	}
	_ = b
	_ = f
	_ = i
	_ = s
}

func BenchmarkAttrString(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = Int64("key", 1).String()
		_ = Float64("key", 1).String()
		_ = Bool("key", true).String()
		_ = String("key", "foo").String()
	}
}

func BenchmarkBool(b *testing.B) {
	k, v := "bool", true
	kv := Bool(k, true)

	b.Run("Value", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			boolValue(v)
		}
	})
	b.Run("Attr", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			Bool(k, v)
		}
	})
	b.Run("ToBool", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			kv.Value.Bool()
		}
	})
	b.Run("String", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = kv.Value.String()
		}
	})
}

func BenchmarkString(b *testing.B) {
	k, v := "string", "foo-bar"
	kv := String(k, "foo-bar")

	b.Run("Value", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			stringValue(v)
		}
	})
	b.Run("Attr", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			String(k, v)
		}
	})
	b.Run("String", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = kv.Value.String()
		}
	})
}
