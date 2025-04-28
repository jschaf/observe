package trace

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strings"
)

// State provides vendor-specific trace information as key-value pairs called
// list-members. Conforms to the W3C Trace Context specification.
// https://www.w3.org/TR/trace-context-1, meaning:
//
//   - No duplicate list-member keys.
//   - At most 32 list-members.
//   - Each key-value pair is valid.
type State struct {
	s       string // s is a valid tracestate string
	isClean bool   // isClean means s doesn't have extra whitespace
}

// ParseState parses a tracestate string into State.
func ParseState(ts string) (State, error) {
	const maxMembers = 32
	if len(ts) == 0 {
		return State{}, nil
	}
	if len(ts) > maxMembers*(256+256+2) {
		return State{}, fmt.Errorf("tracestate too long")
	}

	// Track duplicate lengths as we validate members. If no keys have the same
	// length, then all keys are unique, and we can skip checking for duplicates.
	seenLenBits := new(bitset)
	dupeLenBits := new(bitset)
	nextStart := 0
	count := 0
	isClean := true // track if we can return the original string in String
	for pos := range splitMembers(ts) {
		count++
		if count > maxMembers {
			return State{}, fmt.Errorf("too many members")
		}
		if pos.isEmpty() {
			isClean = false
			continue // spec allows skipping empty members, like "foo=bar ,"
		}
		if !pos.isValid() {
			return State{}, fmt.Errorf("invalid member: %q", pos.memberString(ts))
		}
		if !checkKey(ts, pos) {
			return State{}, fmt.Errorf("invalid key: %q", pos.keyString(ts))
		}
		if !checkVal(ts, pos) {
			return State{}, fmt.Errorf("invalid value: %q", pos.valString(ts))
		}

		isClean = isClean && pos.startsAt(nextStart)
		nextStart = pos.last() + 1
		l := pos.keyLen()
		if seenLenBits.hasBit(l) {
			dupeLenBits.setBit(l)
		}
		seenLenBits.setBit(l)
	}
	isClean = isClean && nextStart > len(ts)

	// Check for dupe keys. We only check duplicated lengths.
	if dupeLenBits.hasAny() {
		// Note: we're checking for duplicates in the hash of the key, not the key
		// itself. We may have false positives, but it should be vanishingly rare.
		var hashes = [32]keyHash{}
		idx := 0
		for pos := range splitMembers(ts) {
			if pos.isEmpty() {
				continue
			}
			if !dupeLenBits.hasBit(pos.keyLen()) {
				continue
			}
			hashes[idx] = pos.hashKey(ts)
			idx++
		}
		slices.Sort(hashes[:idx]) // sorting 32 entries is 2x faster than a map
		for i := 1; i < idx; i++ {
			if hashes[i] == hashes[i-1] {
				return State{}, fmt.Errorf("duplicate key")
			}
		}
	}

	return State{s: ts, isClean: isClean}, nil
}

// Members iterate over each key-value list-member in the tracestate string.
func (st State) Members() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for pos := range splitMembers(st.s) {
			if pos.isEmpty() || !pos.isValid() {
				continue
			}
			k := pos.keyString(st.s)
			v := pos.valString(st.s)
			if !yield(k, v) {
				return
			}
		}
	}
}

func (st State) String() string {
	if st.isClean {
		return st.s
	}
	sb := strings.Builder{}
	sb.Grow(len(st.s))
	for k, v := range st.Members() {
		if sb.Len() > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
	}
	return sb.String()
}

// MarshalJSON marshals the TraceState into JSON.
func (st State) MarshalJSON() ([]byte, error) {
	return json.Marshal(st.String())
}

// bitset tracks up to 64 values. Clamps all values to 0-63.
type bitset uint64

// setBit sets the bit at i to 1.
func (b *bitset) setBit(i uint64) { *b |= bitset(1 << min(i, 63)) }

// hasBit returns true if the bit at index i is set.
func (b *bitset) hasBit(i uint64) bool { return *b&bitset(1<<min(i, 63)) > 0 }

// hasAny returns true if any bit is set.
func (b *bitset) hasAny() bool { return *b != 0 }

// memberPos is the offsets of a list-member key and value in the tracestate
// string.
type memberPos struct {
	keyLo, keyHi int
	valLo, valHi int
}

func (p memberPos) isEmpty() bool                 { return p == memberPos{} }
func (p memberPos) isValid() bool                 { return p.keyLo < p.keyHi && p.valLo < p.valHi }
func (p memberPos) startsAt(offs int) bool        { return p.keyLo == offs }
func (p memberPos) keyLen() uint64                { return uint64(p.keyHi - p.keyLo) }
func (p memberPos) last() int                     { return p.valHi }
func (p memberPos) memberString(ts string) string { return ts[p.keyLo:p.valHi] }
func (p memberPos) keyString(ts string) string    { return ts[p.keyLo:p.keyHi] }
func (p memberPos) valString(ts string) string    { return ts[p.valLo:p.valHi] }

// findMemberPos returns the offsets of the key and value in the tracestate
// string between lo and hi. Returns a zero-value memberPos if the string is
// invalid.
func findMemberPos(ts string, lo, hi int) memberPos {
	// Trim spaces and tabs from the left and right. Required by the spec.
	for lo < hi && isSpace(ts[lo]) {
		lo++
	}
	for hi > lo && isSpace(ts[hi-1]) {
		hi--
	}
	if hi == lo {
		return memberPos{}
	}

	d := strings.IndexByte(ts[lo:hi], '=')
	return memberPos{
		keyLo: lo,
		keyHi: lo + d,
		valLo: lo + d + 1,
		valHi: hi,
	}
}

type keyHash uint64

// Inline FNV hash function.
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
const (
	offset64 = keyHash(14695981039346656037)
	prime64  = 1099511628211
)

// hashKey hashes the key of a list-member to detect duplicates.
func (p memberPos) hashKey(ts string) keyHash {
	h := offset64
	lo, hi := p.keyLo, p.keyHi
	_, _ = ts[lo], ts[hi-1] // check bounds
	for ; lo < hi; lo++ {
		h ^= keyHash(ts[lo])
		h *= prime64
	}
	return h
}

// splitMembers iterates over the tracestate string by list-members. Does not
// validate the members.
func splitMembers(ts string) iter.Seq[memberPos] {
	const delimiter = ','
	return func(yield func(pos memberPos) bool) {
		if len(ts) == 0 {
			return
		}
		lo := 0
		hi := strings.IndexByte(ts, delimiter)
		for ; hi >= 0; hi = strings.IndexByte(ts[lo:], delimiter) {
			hi += lo
			pos := findMemberPos(ts, lo, hi)
			lo = hi + 1
			if !yield(pos) {
				return
			}
		}
		hi = len(ts)
		finalPos := findMemberPos(ts, lo, hi)
		if !yield(finalPos) {
			return
		}
	}
}

func checkKey(ts string, pos memberPos) bool {
	lo, hi := pos.keyLo, pos.keyHi
	if hi <= lo {
		return false
	}
	tenantOffs := strings.IndexByte(ts[lo:hi], '@')
	if tenantOffs < 0 {
		if hi-lo > 256 {
			return false
		}
		return isAlpha(ts[lo]) && checkKeyRest(ts, lo+1, hi)
	}

	if tenantOffs == 0 || tenantOffs == hi-lo-1 {
		return false
	}
	// Check the tenant-id part of a multi-tenant-key.
	if tenantOffs > 241 || !isAlnum(ts[lo]) || !checkKeyRest(ts, lo+1, lo+tenantOffs) {
		return false
	}
	// Check the system-id part of a multi-tenant-key.
	sysStart := lo + tenantOffs + 1
	if hi-lo-tenantOffs-1 > 14 || !isAlpha(ts[sysStart]) || !checkKeyRest(ts, sysStart+1, hi) {
		return false
	}
	return true
}

func checkKeyRest(ts string, lo, hi int) bool {
	if hi == lo {
		return true
	}
	_, _ = ts[lo], ts[hi-1] // check bounds out of loop
	for i := lo; i < hi; i++ {
		if !isAlnum(ts[i]) && ts[i] != '_' && ts[i] != '-' && ts[i] != '*' && ts[i] != '/' {
			return false
		}
	}
	return true

}

func checkVal(ts string, pos memberPos) bool {
	lo, hi := pos.valLo, pos.valHi
	if hi <= lo || hi-lo > 256 {
		return false
	}
	_, _ = ts[lo], ts[hi-1] // check bounds out of loop
	for i := lo; i < hi; i++ {
		if !isPrintable(ts[i]) {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool     { return c >= 'a' && c <= 'z' }
func isAlnum(c byte) bool     { return isAlpha(c) || (c >= '0' && c <= '9') }
func isSpace(c byte) bool     { return c == ' ' || c == '\t' }
func isPrintable(c byte) bool { return c >= ' ' && c <= '~' && c != '=' }
