package difftest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func AssertSame[T any](t *testing.T, msg string, want, got T) {
	t.Helper()
	d := diff(want, got)
	if d != "" {
		t.Error(msg + " (-want +got)\n" + d)
	}
}

//golint:ignore:asciicheck
func diff(a, b any) string {
	switch x := a.(type) {
	case bool, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return diffString(fmt.Sprint(a), fmt.Sprint(b))
	case []bool:
		aOut, err := json.Marshal(a)
		if err != nil {
			return fmt.Sprintf("<error> marshal []bool: %v", err)
		}
		bOut, err := json.Marshal(b)
		if err != nil {
			return fmt.Sprintf("<error> marshal []bool: %v", err)
		}
		return diffString(string(aOut), string(bOut))
	case string:
		y, _ := b.(string)
		return diffString(x, y)
	case time.Time:
		y, _ := b.(time.Time)
		return diffTime(x, y)
	case http.Header:
		y, _ := b.(http.Header)
		if equalMaps(x, y) {
			return ""
		}
		return fmt.Sprintf("- %v\n+ %v", a, b)
	default:
		return fmt.Sprintf("diff not implemented for type: %T", a)
	}
}

func diffString(a, b string) string {
	if a != b {
		return fmt.Sprintf("- %s\n+ %s", a, b)
	}
	return ""
}

func diffTime(a, b time.Time) string {
	if a.Equal(b) {
		return ""
	}
	return fmt.Sprintf("- %s\n+ %s", a.Format(time.RFC3339Nano), b.Format(time.RFC3339Nano))
}

func equalMaps(m1, m2 map[string][]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		v2, ok := m2[k]
		if !ok {
			return false
		}
		if len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if v1[i] != v2[i] {
				return false
			}
		}
	}
	for k, v2 := range m2 {
		v1, ok := m1[k]
		if !ok {
			return false
		}
		if len(v2) != len(v1) {
			return false
		}
		for i := range v2 {
			if v2[i] != v1[i] {
				return false
			}
		}
	}
	return true
}
