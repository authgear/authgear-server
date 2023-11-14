package web

import (
	"net/url"
)

func FormToJSON(form url.Values) map[string]interface{} {
	j := make(map[string]interface{})
	// Do not support recurring parameter
	for name := range form {
		value := form.Get(name)
		if value != "" {
			j[name] = value
		}
	}
	return j
}
