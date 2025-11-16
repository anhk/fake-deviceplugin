package utils

import "encoding/json"

func JsonString(v any) string {
	data, err := json.Marshal(v)
	Must(err)
	return string(data)
}
