package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// anchorSignaturePattern matches an Ethereum-style event signature, e.g.
// "Transfer(uint16,string,bytes)" — a name followed by a parenthesized,
// comma-separated (and possibly empty) list of parameter type tokens.
var anchorSignaturePattern = regexp.MustCompile(`^([A-Za-z_]\w*)\((.*)\)$`)

// optionParameterPattern matches Anchor's option<T> IDL type, e.g.
// "option<string>" or "option<uint64>" — capturing T, which may itself be
// any parameter type token, including another compound one (an array, a
// tuple, or a nested option).
var optionParameterPattern = regexp.MustCompile(`^option<(.+)>$`)

// parseAnchorSignature parses an Ethereum-ABI-style signature into the event
// name and the ordered solana.ParameterDefinition of each declared
// parameter, including compound types: "type[]" (dynamic Vec<T>),
// "type[N]" (fixed [T; N]), and "(type1,type2,...)" (a positional
// tuple/struct) — each recursively parsed, so nesting to arbitrary depth is
// supported (an array of tuples, a tuple containing an array, and so on).
func parseAnchorSignature(signature string) (eventName string, params []solana.ParameterDefinition, err error) {
	matches := anchorSignaturePattern.FindStringSubmatch(strings.TrimSpace(signature))
	if matches == nil {
		return "", nil, fmt.Errorf("invalid signature %q: expected format Name(type1,type2,...)", signature)
	}

	eventName = matches[1]
	body := strings.TrimSpace(matches[2])
	if body == "" {
		return eventName, nil, nil
	}

	tokens, err := splitTopLevel(body)
	if err != nil {
		return "", nil, fmt.Errorf("invalid signature %q: %w", signature, err)
	}

	params = make([]solana.ParameterDefinition, 0, len(tokens))
	for position, token := range tokens {
		def, defErr := parseAnchorParameterType(token, position)
		if defErr != nil {
			return "", nil, fmt.Errorf("invalid signature %q: %w", signature, defErr)
		}
		params = append(params, def)
	}
	return eventName, params, nil
}

// splitTopLevel splits s on commas that sit at bracket/paren depth 0, so a
// tuple field or array element that itself contains commas or brackets is
// not split apart (e.g. "(uint64,bool),string" splits into two tokens, not
// three). Returns an error on unbalanced delimiters.
func splitTopLevel(s string) ([]string, error) {
	var tokens []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced delimiters in %q", s)
			}
		case ',':
			if depth == 0 {
				tokens = append(tokens, s[start:i])
				start = i + 1
			}
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced delimiters in %q", s)
	}
	return append(tokens, s[start:]), nil
}

// parseAnchorParameterType maps a single Ethereum-style type token — scalar
// or compound — to a solana.ParameterDefinition at the given position.
func parseAnchorParameterType(token string, position int) (solana.ParameterDefinition, error) {
	token = strings.TrimSpace(token)

	if strings.HasSuffix(token, "]") {
		return parseArrayParameterType(token, position)
	}
	if strings.HasPrefix(token, "(") && strings.HasSuffix(token, ")") {
		return parseStructParameterType(token, position)
	}
	if matches := optionParameterPattern.FindStringSubmatch(token); matches != nil {
		return parseOptionParameterType(matches[1], position)
	}

	switch token {
	case "bool":
		return solana.NewBoolParameterDefinition(position)
	case "string":
		return solana.NewStringParameterDefinition(position)
	case "address", "pubkey":
		return solana.NewPublicKeyParameterDefinition(position)
	case "bytes":
		return solana.NewBytesParameterDefinition(position)
	}

	if rest, ok := strings.CutPrefix(token, "bytes"); ok {
		if !isPositiveInteger(rest) {
			return nil, fmt.Errorf("invalid fixed-bytes size in %q", token)
		}
		n, _ := strconv.Atoi(rest)
		return solana.NewBytesFixedParameterDefinition(position, n)
	}
	if rest, ok := strings.CutPrefix(token, "uint"); ok {
		if !isPositiveInteger(rest) {
			return nil, fmt.Errorf("uint width is required in %q", token)
		}
		n, _ := strconv.Atoi(rest)
		return solana.NewUintParameterDefinition(position, n)
	}
	if rest, ok := strings.CutPrefix(token, "int"); ok {
		if !isPositiveInteger(rest) {
			return nil, fmt.Errorf("int width is required in %q", token)
		}
		n, _ := strconv.Atoi(rest)
		return solana.NewIntParameterDefinition(position, n)
	}
	if rest, ok := strings.CutPrefix(token, "float"); ok {
		if !isPositiveInteger(rest) {
			return nil, fmt.Errorf("float width is required in %q", token)
		}
		n, _ := strconv.Atoi(rest)
		return solana.NewFloatParameterDefinition(position, n)
	}

	return nil, fmt.Errorf("unsupported parameter type %q", token)
}

// parseOptionParameterType parses inner — the part of an "option<...>" token
// captured by optionParameterPattern — recursively as a parameter type, and
// wraps the result in an OptionParameterDefinition.
func parseOptionParameterType(inner string, position int) (solana.ParameterDefinition, error) {
	innerDef, err := parseAnchorParameterType(inner, 0)
	if err != nil {
		return nil, err
	}
	return solana.NewOptionParameterDefinition(position, innerDef)
}

// parseArrayParameterType parses a token known to end in "]" — either
// "type[]" (dynamic Vec<T>) or "type[N]" (fixed [T; N]) — by locating the
// bracket matching the trailing "]" (scanning right-to-left and tracking
// depth, so "uint64[][]" and "(uint64,bool)[]" both resolve to their
// outermost trailing bracket pair rather than an inner one).
func parseArrayParameterType(token string, position int) (solana.ParameterDefinition, error) {
	depth := 0
	openIdx := -1
	for i := len(token) - 1; i >= 0; i-- {
		switch token[i] {
		case ']':
			depth++
		case '[':
			depth--
		}
		if depth == 0 {
			openIdx = i
			break
		}
	}
	if openIdx == -1 {
		return nil, fmt.Errorf("unbalanced brackets in %q", token)
	}

	elementDef, err := parseAnchorParameterType(token[:openIdx], 0)
	if err != nil {
		return nil, err
	}

	lengthToken := token[openIdx+1 : len(token)-1]
	var length *int
	if lengthToken != "" {
		if !isPositiveInteger(lengthToken) {
			return nil, fmt.Errorf("invalid array length in %q", token)
		}
		n, _ := strconv.Atoi(lengthToken)
		length = &n
	}

	return solana.NewArrayParameterDefinition(position, elementDef, length)
}

// parseStructParameterType parses a token known to be wrapped in "(...)"
// into a positional tuple/struct — each field recursively parsed, with its
// own 0-based position restarting within the struct.
func parseStructParameterType(token string, position int) (solana.ParameterDefinition, error) {
	fieldTokens, err := splitTopLevel(token[1 : len(token)-1])
	if err != nil {
		return nil, err
	}

	fields := make([]solana.ParameterDefinition, 0, len(fieldTokens))
	for i, ft := range fieldTokens {
		fieldDef, fieldErr := parseAnchorParameterType(ft, i)
		if fieldErr != nil {
			return nil, fieldErr
		}
		fields = append(fields, fieldDef)
	}

	return solana.NewStructParameterDefinition(position, fields)
}

// isPositiveInteger reports whether s is a non-empty sequence of digits
// representing a strictly positive integer (rejects "", "0", "-1", "abc").
func isPositiveInteger(s string) bool {
	if s == "" {
		return false
	}
	n, err := strconv.Atoi(s)
	return err == nil && n > 0
}
