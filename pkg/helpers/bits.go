package helpers

type Bits uint8

func (f Bits) Has(flag Bits) bool { return f&flag != 0 }
func (f *Bits) Set(flag Bits)     { *f |= flag }

//noinspection GoUnusedExportedFunction
func (f *Bits) ClearFlag(flag Bits) { *f &= ^flag }

//noinspection GoUnusedExportedFunction
func (f *Bits) ToggleFlag(flag Bits) { *f ^= flag }
