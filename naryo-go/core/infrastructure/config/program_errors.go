package config

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// ProgramErrorCodeProperties is a single Anchor Custom(u32) code -> name
// entry, as it appears in YAML.
type ProgramErrorCodeProperties struct {
	Name string `mapstructure:"name"`
	Code uint32 `mapstructure:"code"`
}

// ProgramErrorsProperties is one program's Anchor error-code table, as it
// appears in YAML.
type ProgramErrorsProperties struct {
	ProgramID string                       `mapstructure:"programId"`
	Codes     []ProgramErrorCodeProperties `mapstructure:"codes"`
}

// ProgramErrors is the raw naryo.programErrors YAML section: a list of
// per-program Anchor error-code tables, one entry per program ID. Unlike
// FilterProperties/NodeProperties, which each map to one domain value,
// this whole list maps to a single event.ProgramErrorRegistry -- Map
// aggregates every entry into that one map. If a program ID appears more
// than once across entries, the later entry's codes win.
type ProgramErrors []*ProgramErrorsProperties

func (p ProgramErrors) Map() (event.ProgramErrorRegistry, error) {
	registry := make(event.ProgramErrorRegistry, len(p))
	for _, entry := range p {
		codes := make(map[uint32]string, len(entry.Codes))
		for _, c := range entry.Codes {
			codes[c.Code] = c.Name
		}
		registry[entry.ProgramID] = codes
	}
	return registry, nil
}

var _ descriptor.ProgramErrorRegistry = ProgramErrors(nil)
