package trace

const FlagsSampled = Flags(0x01)

// Flags represent flags on a Context.
type Flags uint8

func (f Flags) IsSampled() bool { return f&FlagsSampled == FlagsSampled }
