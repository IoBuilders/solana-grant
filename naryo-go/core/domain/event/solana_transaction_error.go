package event

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

// parsedTransactionError is the result of parsing a raw meta.err JSON value:
// either a data-less variant name (built-in TransactionError, or a data-less
// InstructionError) or a program-defined Custom(u32) code, plus the failing
// instruction's index when meta.err named one.
type parsedTransactionError struct {
	VariantName      string  // set when err is a bare string, at either level
	CustomCode       *uint32 // set only for {"InstructionError":[i,{"Custom":N}]}
	InstructionIndex *int    // set whenever err is {"InstructionError":[i,...]}
}

// parseTransactionError parses meta.err's mixed JSON shape: a bare string (a
// data-less top-level variant), or {"InstructionError":[i,V]} where V is
// itself either a bare string or {"Custom":N}.
func parseTransactionError(raw string) (parsedTransactionError, error) {
	var bare string
	if err := json.Unmarshal([]byte(raw), &bare); err == nil {
		return parsedTransactionError{VariantName: bare}, nil
	}

	var wrapper struct {
		InstructionError *[2]json.RawMessage `json:"InstructionError"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil || wrapper.InstructionError == nil {
		return parsedTransactionError{}, fmt.Errorf("unrecognized transaction error shape: %s", raw)
	}

	var index int
	if err := json.Unmarshal(wrapper.InstructionError[0], &index); err != nil {
		return parsedTransactionError{}, fmt.Errorf("invalid instruction index: %w", err)
	}

	inner := wrapper.InstructionError[1]
	var innerBare string
	if err := json.Unmarshal(inner, &innerBare); err == nil {
		return parsedTransactionError{VariantName: innerBare, InstructionIndex: &index}, nil
	}

	var custom struct {
		Custom *uint32 `json:"Custom"`
	}
	if err := json.Unmarshal(inner, &custom); err == nil && custom.Custom != nil {
		return parsedTransactionError{CustomCode: custom.Custom, InstructionIndex: &index}, nil
	}

	return parsedTransactionError{}, fmt.Errorf("unrecognized instruction error shape: %s", string(inner))
}

var anchorErrorLogPattern = regexp.MustCompile(
	`Error Code: (\w+)\. Error Number: (\d+)\. Error Message: (.+)\.`,
)

// anchorErrorFromLogs scans logs for Anchor's own error-report line matching
// code, returning the error's name and message as Anchor itself logged
// them. found is false only if no such line is present, which happens when
// an RPC node truncates or omits logs, or the error never routed through
// Anchor's logging wrapper.
func anchorErrorFromLogs(logs []string, code uint32) (name, message string, found bool) {
	codeStr := strconv.FormatUint(uint64(code), 10)
	for _, line := range logs {
		m := anchorErrorLogPattern.FindStringSubmatch(line)
		if m != nil && m[2] == codeStr {
			return m[1], m[3], true
		}
	}
	return "", "", false
}

// DecodedError returns a human-readable reason for a failed transaction. ok
// is false only when t did not fail. registry is consulted only as a
// fallback for a Custom(u32) error whose log line is unavailable; an empty
// registry still decodes everything the log line and the RPC's own error
// variants cover, just with a plainer result for the rest.
func (t SolanaTransactionEvent) DecodedError(registry ProgramErrorRegistry) (reason string, ok bool) {
	if t.Err == nil {
		return "", false
	}
	parsed, err := parseTransactionError(*t.Err)
	if err != nil {
		return *t.Err, true // unrecognized shape -- surface the raw value rather than nothing
	}
	if parsed.VariantName != "" {
		return parsed.VariantName, true // already human-readable, straight from the RPC
	}

	if name, message, found := anchorErrorFromLogs(t.Logs, *parsed.CustomCode); found {
		if message == "" {
			return name, true
		}
		return fmt.Sprintf("%s: %s", name, message), true
	}

	programID := t.instructionProgramID(parsed.InstructionIndex)
	if name, found := registry.Lookup(programID, *parsed.CustomCode); found {
		return name, true // log was truncated/missing -- fall back to the configured name
	}

	return fmt.Sprintf("custom program error: %d", *parsed.CustomCode), true // no log, no config entry
}

func (t SolanaTransactionEvent) instructionProgramID(index *int) string {
	if index == nil || *index < 0 || *index >= len(t.Instructions) {
		return ""
	}
	return t.Instructions[*index].ProgramID
}
