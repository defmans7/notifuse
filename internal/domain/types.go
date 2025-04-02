package domain

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
)

// MapOfAny is persisted as JSON in the database
type MapOfAny map[string]any

// Scan implements the sql.Scanner interface
func (m *MapOfAny) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, m)
}

// Value implements the driver.Valuer interface
func (m MapOfAny) Value() (driver.Value, error) {
	return json.Marshal(m)
}
