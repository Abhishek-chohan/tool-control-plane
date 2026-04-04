package storage

import (
	"encoding/json"
)

func toJSON(value interface{}) []byte {
	if value == nil {
		return []byte("null")
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return []byte("null")
	}
	return bytes
}

func fromJSON(data []byte, out interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
