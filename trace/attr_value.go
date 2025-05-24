package trace

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"
)

// A Value represents an int64, float64, bool, string, or a slice of those
// types.
type Value struct {
	_ [0]func() // disallow ==
	// num is a taggedNum that stores a value for bool, int64, or float64,
	// or a tag and length for a string or slice.
	num taggedNum
	// data is a pointer to the data for a string or slice, or a marker
	// pointer for a bool, int64, or float64.
	data unsafe.Pointer
}

// taggedNum is a uint64 that stores a value and an optional tag.
//   - bool: no tag, value is 0 or 1
//   - float64: no tag, value is math.Float64bits(v)
//   - int64: no tag, value is int64(v)
//   - string: tag is kindString, value is len(v)
//   - bool slice: tag is the kindBoolSlice, value is len(v)
//   - float64 slice: tag is the kindFloat64Slice, value is len(v)
//   - int64 slice: tag is the kindInt64Slice, value is len(v)
//   - string slice: tag is the kindStringSlice, value is len(v)
type taggedNum struct {
	n uint64
}

// tagSize is the number of bits most significant bits used for the tag.
const tagSize = 8

func newTaggedNum(kind valueKind, num uint64) taggedNum {
	switch kind {
	case kindUnset:
		return taggedNum{n: 0}
	case kindBool, kindFloat64, kindInt64:
		return taggedNum{n: num}
	case kindBoolSlice:
		return taggedNum{n: num} // length encoded in the most significant bits
	case kindString, kindFloat64Slice, kindInt64Slice, kindStringSlice:
		return taggedNum{n: uint64(kind)<<(64-tagSize) | num}
	default:
		panic(fmt.Sprintf("unknown value kind: %s", kind))
	}
}

// value returns the value of a bool, int64, float64 value, or the bool slice
// bitset.
func (t taggedNum) value() uint64 { return t.n }

// kind returns the kind of a string or slice.
func (t taggedNum) kind() valueKind {
	return valueKind(t.n >> (64 - tagSize)) //nolint:gosec // safe bit shift
}

// len returns the length of a string or slice.
func (t taggedNum) len() uint64 { return t.n << tagSize >> tagSize }

type valueKind uint8

const (
	kindUnset        valueKind = iota
	kindBool                   // data == dataKindBool      num == 0 or 1
	kindFloat64                // data == dataKindFloat64   num == math.Float64bits(v)
	kindInt64                  // data == dataKindInt64     num == int64(v)
	kindString                 // data == string data ptr   num == tag & len(v)
	kindBoolSlice              // data == slice data ptr    num == tag & len(v)
	kindFloat64Slice           // data == slice data ptr    num == tag & len(v)
	kindInt64Slice             // data == slice data ptr    num == tag & len(v)
	kindStringSlice            // data == slice data ptr    num == tag & len(v)
)

// Marker pointers for the data field of Value. Indicates the kind of data
// store in Value.num.
var (
	dataKindBool      = (unsafe.Pointer)(new(byte)) //nolint:gochecknoglobals // marker pointer
	dataKindFloat64   = (unsafe.Pointer)(new(byte)) //nolint:gochecknoglobals // marker pointer
	dataKindInt64     = (unsafe.Pointer)(new(byte)) //nolint:gochecknoglobals // marker pointer
	dataKindBoolSlice = (unsafe.Pointer)(new(byte)) //nolint:gochecknoglobals // marker pointer
)

//nolint:gochecknoglobals // string literals for string func
var valueKindStrings = []string{
	"Unset",
	"Bool",
	"Float64",
	"Int64",
	"String",
	"BoolSlice",
	"Float64Slice",
	"Int64Slice",
	"StringSlice",
}

func (k valueKind) String() string {
	return valueKindStrings[k]
}

// boolValue returns a [Value] for a bool.
func boolValue(v bool) Value {
	u := uint64(0)
	if v {
		u = 1
	}
	return Value{
		num:  newTaggedNum(kindBool, u),
		data: dataKindBool,
	}
}

// int64Value returns a [Value] for an int64.
func int64Value(v int64) Value {
	return Value{
		num:  newTaggedNum(kindInt64, uint64(v)), //nolint:gosec // safe conversion to uint64
		data: dataKindInt64,
	}
}

// float64Value returns a [Value] for a floating-point number.
func float64Value(v float64) Value {
	return Value{
		num:  newTaggedNum(kindFloat64, math.Float64bits(v)),
		data: dataKindFloat64,
	}
}

// stringValue returns a new [Value] for a string.
func stringValue(value string) Value {
	return Value{
		num:  newTaggedNum(kindString, uint64(len(value))),
		data: (unsafe.Pointer)(unsafe.StringData(value)),
	}
}

func boolsValue(v []bool) Value {
	hi := uint64(min(len(v), 56)) //nolint:gosec // len is positive, no underflow
	var n uint64
	for i := range hi {
		n <<= 1
		if v[i] {
			n |= 1
		}
	}
	n |= hi << (64 - tagSize)

	return Value{
		num:  newTaggedNum(kindBoolSlice, n),
		data: dataKindBoolSlice,
	}
}

func (v Value) kind() valueKind {
	switch v.data {
	case dataKindBool:
		return kindBool
	case dataKindInt64:
		return kindInt64
	case dataKindFloat64:
		return kindFloat64
	case dataKindBoolSlice:
		return kindBoolSlice
	default:
		return v.num.kind()
	}
}

// Unchecked accessors
// ===================

func (v Value) uncheckedBool() bool       { return v.num.value() == 1 }
func (v Value) uncheckedFloat64() float64 { return math.Float64frombits(v.num.value()) }
func (v Value) uncheckedInt64() int64     { return int64(v.num.value()) } //nolint:gosec // safe conversion to int64
func (v Value) uncheckedString() string   { return unsafe.String((*byte)(v.data), v.num.len()) }

func (v Value) uncheckedBools() []bool {
	count := int(v.num.n >> (64 - tagSize)) //nolint:gosec // only 8 bits
	bs := make([]bool, count)
	n := v.num.value() << tagSize >> tagSize
	for i := count - 1; i >= 0; i-- {
		bs[i] = (n & 1) != 0
		n >>= 1
	}
	return bs
}

// Checked accessors
// =================

// Any returns the value of v as any.
func (v Value) Any() any {
	switch k := v.kind(); k {
	case kindUnset:
		return nil
	case kindBool:
		return v.uncheckedBool()
	case kindFloat64:
		return v.uncheckedFloat64()
	case kindInt64:
		return v.uncheckedInt64()
	case kindString:
		return v.uncheckedString()
	case kindBoolSlice, kindFloat64Slice, kindInt64Slice, kindStringSlice:
		panic("unimplemented: Value.Any() for slices")
	default:
		panic(fmt.Sprintf("bad kind: %s", k))
	}
}

// Bool returns v's value as a bool. It panics if v is not a bool.
func (v Value) Bool() bool {
	if g, w := v.kind(), kindBool; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}
	return v.uncheckedBool()
}

// Int64 returns v's value as an int64. It panics if v is not an int64.
func (v Value) Int64() int64 {
	if g, w := v.kind(), kindInt64; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}
	return v.uncheckedInt64()
}

// Float64 returns v's value as a float64. It panics if v is not a float64.
func (v Value) Float64() float64 {
	if g, w := v.kind(), kindFloat64; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}
	return v.uncheckedFloat64()
}

// String returns Value's value as a string, formatted like [fmt.Sprint].
// String never panics, even if v is not a string.
func (v Value) String() string {
	switch k := v.kind(); k {
	case kindUnset:
		return ""
	case kindBool:
		return strconv.FormatBool(v.uncheckedBool())
	case kindInt64:
		return strconv.FormatInt(v.uncheckedInt64(), 10)
	case kindFloat64:
		return strconv.FormatFloat(v.uncheckedFloat64(), 'g', -1, 64)
	case kindString:
		return v.uncheckedString()
	case kindBoolSlice:
		sb := strings.Builder{}
		sb.WriteByte('[')
		for i, b := range v.Bools() {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(strconv.FormatBool(b))
		}
		sb.WriteByte(']')
		return sb.String()
	case kindFloat64Slice, kindInt64Slice, kindStringSlice:
		panic("unimplemented: Value.String() for slices")
	default:
		panic(fmt.Sprintf("bad kind: %s", k))
	}
}

func (v Value) Bools() []bool {
	if g, w := v.kind(), kindBoolSlice; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}
	return v.uncheckedBools()
}
