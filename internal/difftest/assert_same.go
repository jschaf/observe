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
		y, _ := b.([]bool)
		return diffSlices(x, y)
	case []int:
		y, _ := b.([]int)
		return diffSlices(x, y)
	case []int64:
		y, _ := b.([]int64)
		return diffSlices(x, y)
	case []float64:
		y, _ := b.([]float64)
		return diffSlices(x, y)
	case []string:
		y, _ := b.([]string)
		return diffSlices(x, y)
	case []uint64:
		y, _ := b.([]uint64)
		return diffSlices(x, y)
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

func diffSlices[T any](a []T, b []T) string {
	if len(a) == 0 && len(b) == 0 {
		return ""
	}
	aOut, err := json.Marshal(a)
	if err != nil {
		return fmt.Sprintf("<error> marshal %T: %v", a, err)
	}
	bOut, err := json.Marshal(b)
	if err != nil {
		return fmt.Sprintf("<error> marshal %T: %v", b, err)
	}
	return diffString(string(aOut), string(bOut))
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
