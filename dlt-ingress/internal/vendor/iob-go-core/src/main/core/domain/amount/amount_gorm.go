package amount

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func (a Amount) Value() (driver.Value, error) {
	return a.String(), nil
}

func (a *Amount) Scan(src any) error {
	if src == nil {
		*a = Amount{}
		return nil
	}
	switch v := src.(type) {
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil
		}
		amt, err := NewFromString(s)
		if err != nil {
			return err
		}
		*a = *amt
		return nil

	case []byte:
		amt, err := NewFromString(string(v))
		if err != nil {
			return err
		}
		*a = *amt
		return nil
	default:
		return fmt.Errorf("unsupported type for Amount Scan: %T", src)
	}
}
