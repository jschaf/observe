package difftest

import "testing"

func AssertSame[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("diff mismatch:\nwant %v\ngot  %v", want, got)
	}
}
