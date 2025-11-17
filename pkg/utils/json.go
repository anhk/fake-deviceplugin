package utils

import "encoding/json"

func JsonString(v any) string {
	data, err := json.Marshal(v)
	PanicIfError(err)
	return string(data)
}
