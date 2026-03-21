package dbutil

import "encoding/json"

// MustMarshal marshals v to JSON. Panics if marshaling fails.
// Use for struct-to-JSONB conversions that should never fail (e.g., Go structs with json tags).
func MustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic("dbutil.MustMarshal: " + err.Error())
	}
	return b
}
