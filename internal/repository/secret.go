package repository

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Secret wraps a sensitive string value (passwords, tokens, API keys).
// It prevents accidental leakage via logging, %v/%+v, JSON, etc.,
// while still working transparently with GORM for DB storage.
type Secret string

// String prevents leakage via fmt.Println, %v, %s, error messages, etc.
func (s Secret) String() string {
	return "[REDACTED]"
}

// GoString prevents leakage via %#v and debuggers.
func (s Secret) GoString() string {
	return "[REDACTED]"
}

// MarshalJSON prevents leakage when the struct is serialized to JSON
// (e.g. logged, or accidentally returned in an API response).
func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

// UnmarshalJSON still allows the real value to come in from requests/config.
func (s *Secret) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = Secret(str)
	return nil
}

// Value implements driver.Valuer so GORM can write the real value to the DB.
func (s Secret) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner so GORM can read the real value from the DB.
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

// ExposeSecret provides explicit, searchable access to the real value.
// Prefer this over string(s) so all unwrapping sites are grep-able.
func (s Secret) Expose() string {
	return string(s)
}