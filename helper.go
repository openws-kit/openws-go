package openws

import "encoding/json"

// mustMarshal is a wrapper around json.Marshal that panics on error. It is
// intended for use in situations where you are confident that the input data
// can be marshaled without error.
func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic("error marshaling JSON: " + err.Error())
	}

	return data
}
