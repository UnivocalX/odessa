package core

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Secret wraps a sensitive string value and redacts itself in logs/JSON.
type Secret string

func (s Secret) String() string {
	return "[REDACTED]"
}

func (s Secret) GoString() string {
	return "[REDACTED]"
}

func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = Secret(str)
	return nil
}

func (s Secret) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *Secret) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		b, ok := value.([]byte)
		if !ok {
			return errors.New("secret: expected string or []byte")
		}
		str = string(b)
	}
	*s = Secret(str)
	return nil
}

func (s Secret) Expose() string {
	return string(s)
}
