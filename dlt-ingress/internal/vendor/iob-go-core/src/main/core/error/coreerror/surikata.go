package coreerror

import "fmt"

func NewContractInstanceError(name string, err error) error {
	return fmt.Errorf("create contract %s instance failed: %w", name, err)
}

func NewDecodingEventError(name string, err error) error {
	return fmt.Errorf("decoding of event %s failed: %w", name, err)
}
