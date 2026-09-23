package event

// ProgramErrorRegistry maps a program's Anchor Custom(u32) codes to their
// declared names, supplied directly by whoever configures the node: naryo-go
// has no IDL to derive this from automatically. Consulted only as a
// decode-time fallback for a program whose log line is missing or
// truncated -- never as filter-matching criteria, and never required for
// decoding to proceed. Keyed by program ID because Anchor error codes are
// only unique within a single program.
type ProgramErrorRegistry map[string]map[uint32]string

// Lookup returns the declared name for code within programID's table, if
// any. ok is false when programID is not registered or code is not one of
// its declared entries.
func (r ProgramErrorRegistry) Lookup(programID string, code uint32) (string, bool) {
	codes, ok := r[programID]
	if !ok {
		return "", false
	}
	name, ok := codes[code]
	return name, ok
}
