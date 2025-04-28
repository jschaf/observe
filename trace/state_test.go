package trace_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/jschaf/observe/internal/difftest"
	"github.com/jschaf/observe/trace"
)

func TestParseState(t *testing.T) {
	// Taken from the W3C tests:
	// https://github.com/w3c/trace-context/blob/dcd3ad9b7d6ac36f70ff3739874b73c11b0302a1/test/test_data.json
	tests := []struct {
		name string
		in   string
		want string
		err  string
	}{
		// Error cases
		{
			name: "duplicate with the same value",
			in:   "foo=1,foo=1",
			err:  "duplicate key",
		},
		{
			name: "duplicate with different values",
			in:   "foo=1,foo=2",
			err:  "duplicate key",
		},
		{
			name: "improperly formatted key/value pair",
			in:   "foo =1",
			err:  "invalid key",
		},
		{
			name: "upper case key",
			in:   "FOO=1",
			err:  "invalid key",
		},
		{
			name: "no equal",
			in:   "no-eq",
			err:  "invalid member",
		},
		{
			name: "no val",
			in:   "a=",
			err:  "invalid member",
		},
		{
			name: "only eq",
			in:   " = ",
			err:  "invalid member",
		},
		{
			name: "mixed case key",
			in:   "aA=0",
			err:  "invalid key",
		},
		{
			name: "key with invalid character",
			in:   "foo.bar=1",
			err:  "invalid key",
		},
		{
			name: "multiple keys, one with empty tenant key",
			in:   "foo@=1,bar=2",
			err:  "invalid key",
		},
		{
			name: "multiple keys, one with only tenant",
			in:   "@foo=1,bar=2",
			err:  "invalid key",
		},
		{
			name: "multiple keys, one with double tenant separator",
			in:   "foo@@bar=1,bar=2",
			err:  "invalid key",
		},
		{
			name: "multiple keys, one with multiple tenants",
			in:   "foo@bar@baz=1,bar=2",
			err:  "invalid key",
		},
		{
			name: "key too long",
			in:   "foo=1," + strings.Repeat("z", 257) + "=1",
			err:  "invalid key",
		},
		{
			name: "key too long, with tenant",
			in:   "foo=1," + strings.Repeat("t", 242) + "@v=1",
			err:  "invalid key",
		},
		{
			name: "tenant too long",
			in:   "foo=1,t@vvvvvvvvvvvvvvv=1",
			err:  "invalid key",
		},
		{
			name: "multiple values for a single key",
			in:   "foo=bar=baz",
			err:  "invalid value",
		},
		{
			name: "second member no value",
			in:   "foo=,bar=3",
			err:  "invalid member",
		},
		{
			name: "too many members",
			in:   genTracestateSequence("bar%d=%1", 33),
			err:  "too many members",
		},
		// Valid cases
		{
			name: "valid key/value list",
			in:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
			want: "abcdefghijklmnopqrstuvwxyz0123456789_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		},
		{
			name: "valid key/value list with tenant",
			in:   "abcdefghijklmnopqrstuvwxyz0123456789_-*/@a-z0-9_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
			want: "abcdefghijklmnopqrstuvwxyz0123456789_-*/@a-z0-9_-*/= !\"#$%&'()*+-./0123456789:;<>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
		{
			name: "single key and value",
			in:   "foo=1",
			want: "foo=1",
		},
		{
			name: "empty key first",
			in:   "  ,  foo=1",
			want: "foo=1",
		},
		{
			name: "empty key last",
			in:   "foo=1,bar=2\t,  ",
			want: "foo=1,bar=2",
		},
		{
			name: "only comma",
			in:   ",",
			want: "",
		},
		{
			name: "tenant with digit",
			in:   "0@a=0",
			want: "0@a=0",
		},
		{
			name: "space after equal",
			in:   "a= 0",
			want: "a= 0",
		},
		{
			name: "single key and value with empty separator",
			in:   "foo=1,",
			want: "foo=1",
		},
		{
			name: "space between two keys",
			in:   "foo=1\t,  bar=3",
			want: "foo=1,bar=3",
		},
		{
			name: "multiple keys and values",
			in:   "foo=1,bar=2",
			want: "foo=1,bar=2",
		},
		{
			name: "with a key at maximum length",
			in:   "foo=1," + strings.Repeat("z", 256) + "=1",
			want: "foo=1," + strings.Repeat("z", 256) + "=1",
		},
		{
			name: "with a key and tenant at maximum length",
			in:   "foo=1," + strings.Repeat("t", 241) + "@" + strings.Repeat("v", 14) + "=1",
			want: "foo=1," + strings.Repeat("t", 241) + "@" + strings.Repeat("v", 14) + "=1",
		},
		{
			name: "with maximum members",
			in:   genTracestateSequence("bar%d=%d", 32),
			want: genTracestateSequence("bar%d=%d", 32),
		},
		{
			name: "with several members",
			in:   "foo=1,bar=2,rojo=1,congo=2,baz=3",
			want: "foo=1,bar=2,rojo=1,congo=2,baz=3",
		},
		{
			name: "medium keys",
			in:   "redis/inc-count@tenant=123,x-trace-id=deadbeefCafe1234,x-cache-hit=1,pg-req-id=1234567890,x-dd-trace-id=deadbeefCafe1234",
			want: "redis/inc-count@tenant=123,x-trace-id=deadbeefCafe1234,x-cache-hit=1,pg-req-id=1234567890,x-dd-trace-id=deadbeefCafe1234",
		},
		{
			name: "with tabs between members",
			in:   "foo=1 \t , \t bar=2, \t baz=3",
			want: "foo=1,bar=2,baz=3",
		},
		{
			name: "with multiple tabs between members",
			in:   "foo=1\t \t,\t \tbar=2,\t \tbaz=3",
			want: "foo=1,bar=2,baz=3",
		},
		{
			name: "with space at the end of the member",
			in:   "foo=1 ",
			want: "foo=1",
		},
		{
			name: "with tab at the end of the member",
			in:   "foo=1\t",
			want: "foo=1",
		},
		{
			name: "with tab and space at the end of the member",
			in:   "foo=1 \t",
			want: "foo=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := trace.ParseState(tt.in)
			if err != nil {
				if tt.err == "" {
					t.Errorf("got parse state error: %v", err)
				} else if !strings.Contains(err.Error(), tt.err) {
					t.Errorf("want error %q substring; not found in: %s", tt.err, err.Error())
				}
				return
			}
			if tt.err != "" {
				t.Errorf("want error %q, got nil", tt.err)
				return
			}

			// String
			gotStr := got.String()
			difftest.AssertSame(t, "State.String() mismatch", tt.want, gotStr)

			// Members
			gotMembers := ""
			for k, v := range got.Members() {
				if gotMembers != "" {
					gotMembers += ","
				}
				gotMembers += k + "=" + v
			}
			difftest.AssertSame(t, "State.Members() mismatch", tt.want, gotMembers)

			// JSON
			wantJSON, err := json.Marshal(tt.want)
			if err != nil {
				t.Errorf("marshal wanted trace state json: %v", err)
			}
			gotJSON, err := json.Marshal(got)
			if err != nil {
				t.Errorf("marshal trace state json: %v", err)
			}
			difftest.AssertSame(t, "State JSON mismatch", string(wantJSON), string(gotJSON))
		})
	}
}

func genTracestateSequence(tmpl string, n int) string {
	seq := make([]string, n)
	for i := 0; i < n; i++ {
		seq[i] = strings.ReplaceAll(tmpl, "%d", strconv.Itoa(i))
	}
	return strings.Join(seq, ",")
}

func FuzzParseState(f *testing.F) {
	f.Add("   , ")
	f.Add("no-key")
	f.Add("   = \t")
	f.Add("foo=1,bar=2\t,  ")
	f.Add("0a@123/456=abCDef/foo-bar")
	f.Fuzz(func(t *testing.T, in string) {
		fastState, fastErr := trace.ParseState(in)
		slowState, slowErr := slowParseState(in)
		if (fastErr == nil) != (slowErr == nil) {
			t.Errorf("fast parse error: %v, slow parse error: %v\ninput: %q", fastErr, slowErr, in)
			return
		}
		if fastErr != nil {
			return
		}
		if fastState.String() != slowState {
			t.Errorf("fast state: %q, slow state: %q", fastState.String(), slowState)
			return
		}
	})
}

var (
	stateKeyRE = regexp.MustCompile(`^([a-z][a-z0-9_*/-]{0,255}|[a-z0-9][a-z0-9_*/-]{0,240}@[a-z][a-z0-9_*/-]{0,13})$`)
	statValRE  = regexp.MustCompile(`^[\x20-\x2B\x2D-\x3C\x3E-\x7E]{0,255}[\x21-\x2B\x2D-\x3C\x3E-\x7E]$`)
)

// slowParseState is an alternate implementation of ParseState to verify
// correctness with fuzzing.
func slowParseState(s string) (string, error) {
	sb := strings.Builder{}
	seen := make(map[string]struct{})
	i := 0
	for m := range strings.SplitSeq(s, ",") {
		i++
		if i > 32 {
			return "", fmt.Errorf("too many members")
		}
		m = strings.Trim(m, " \t")
		if m == "" {
			continue
		}
		k, v, hasEq := strings.Cut(m, "=")
		if !hasEq {
			return "", fmt.Errorf("invalid member: %q", m)
		}
		if !stateKeyRE.MatchString(k) {
			return "", fmt.Errorf("invalid key: %q", k)
		}
		if !statValRE.MatchString(v) {
			return "", fmt.Errorf("invalid value: %q", v)
		}
		if _, ok := seen[k]; ok {
			return "", fmt.Errorf("duplicate key: %q", k)
		}
		seen[k] = struct{}{}
		if sb.Len() > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
	}
	return sb.String(), nil
}

func Test_slowParseState(t *testing.T) {
	tests := []struct {
		name string
		in   string
		err  string
	}{
		{
			name: "empty",
			in:   "",
		},
		{
			name: "valid state",
			in:   "a= 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := slowParseState(tt.in)
			if (err != nil) != (tt.err != "") {
				t.Errorf("got parse state error: %v, want %q", err, tt.err)
				return
			}
			if err != nil {
				t.Errorf("slow parse state error: %v", err)
				return
			}
		})
	}
}

func BenchmarkParseState(b *testing.B) {
	benches := []struct {
		name string
		in   string
	}{
		{
			name: "empty state",
			in:   "",
		},
		{
			name: "single key",
			in:   "longish-key=alpha-bravo-charlie",
		},
		{
			name: "single tenant key",
			in:   "longish-key@a-tenant=foo-bar-baz-qux",
		},
		{
			name: "five realistic keys",
			in:   "redis/inc-count@tenant=foobar123,x-trace-id=deadbeefCafe1234,x-cache-hit=1,pg-req-id=1234567890,x-dd-trace-id=deadbeefCafe1234",
		},
		{
			name: "max keys",
			in:   genTracestateSequence("bar%d=%d", 32),
		},
	}
	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, err := trace.ParseState(bench.in)
				if err != nil {
					b.Fatalf("parse trace state: %v", err)
				}
			}
		})
	}
}
