package trace

// An Attr is a key-value pair.
type Attr struct {
	Key   string
	Value Value
}

// Bool returns an Attr for a bool.
func Bool(key string, value bool) Attr {
	return Attr{key, boolValue(value)}
}

// Bools returns an Attr for a bool slice.
// Only records the first 56 elements, dropping subsequent elements.
func Bools(key string, value []bool) Attr {
	return Attr{key, boolsValue(value)}
}

// Float64 returns an Attr for a floating-point number.
func Float64(key string, value float64) Attr {
	return Attr{key, float64Value(value)}
}

// Int64 returns an Attr for an int64.
func Int64(key string, value int64) Attr {
	return Attr{key, int64Value(value)}
}

// Int converts an int to an int64 and returns an Attr with that value.
func Int(key string, value int) Attr {
	return Attr{key, int64Value(int64(value))}
}

// String returns an Attr for a string value.
func String(key, value string) Attr {
	return Attr{key, stringValue(value)}
}

func (a Attr) String() string {
	return a.Key + "=" + a.Value.String()
}
