// Package errtypes contains custom error types
package errtypes

import (
	"fmt"
	"strings"
)

const (
	UnknownLycheeKeyErrMsg = "unknown lychee key"
	InvalidModelNameErrMsg = "invalid model name"
)

// TODO: This should have a structured response from the API
type UnknownLycheeKey struct {
	Key string
}

func (e *UnknownLycheeKey) Error() string {
	return fmt.Sprintf("unauthorized: %s %q", UnknownLycheeKeyErrMsg, strings.TrimSpace(e.Key))
}
