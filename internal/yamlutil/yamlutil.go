package yamlutil

import "encoding/json"

// QuoteString returns a double-quoted YAML scalar. JSON string escaping is also
// valid YAML double-quoted scalar escaping for the plain text values we emit.
func QuoteString(value string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return `""`
	}
	return string(data)
}
