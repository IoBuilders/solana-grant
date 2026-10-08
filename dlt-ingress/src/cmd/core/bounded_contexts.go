package cmdcore

import "strings"

type BoundedContext string

const (
	DLT_INGRESS = "dltingress"
)

var BOUNDED_CONTEXTS = []string{
	DLT_INGRESS,
}

// HANDLER_GEN_BCS are the bounded contexts whose HTTP endpoints and event listeners use
// the generated HttpEndpoint/RegistrableListener methods (genhttpendpoint, genlistener).
// Add a context here once it adopts that pattern.
var HANDLER_GEN_BCS = []string{
	DLT_INGRESS,
}

var OPENAPI_TAGS = map[string]string{
	DLT_INGRESS: "DLT Ingress",
}

// API_SEGMENTS is the first path segment each bounded context serves its endpoints under, for the
// contexts whose package name and public segment differ.
var API_SEGMENTS = map[string]string{}

func ApiSegment(bc string) string {
	if segment, found := API_SEGMENTS[bc]; found {
		return segment
	}
	return bc
}

func OpenApiTag(bc string) string {
	if tag, found := OPENAPI_TAGS[bc]; found {
		return tag
	}
	if bc == "" {
		return bc
	}
	return strings.ToUpper(bc[:1]) + bc[1:]
}

var MIGRATED_BCS = map[string]bool{
	DLT_INGRESS: true,
}

var DISABLED_BCS = map[string]bool{}

var CALLBACK_SKIP_BCS = map[string]bool{
	DLT_INGRESS: true,
}

var ERROR_REFACTOR_BCS = map[string]bool{
	DLT_INGRESS: true,
}

func IsDisabled(bc string) bool {
	return DISABLED_BCS[bc]
}

func IsMigrated(bc string) bool {
	return MIGRATED_BCS[bc]
}

func SkipCallback(bc string) bool {
	return CALLBACK_SKIP_BCS[bc]
}

func IsErrorRefactor(bc string) bool {
	return ERROR_REFACTOR_BCS[bc]
}
