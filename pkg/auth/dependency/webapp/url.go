package webapp

import (
	"net/url"
	"strings"
)

func RemoveX(q url.Values) {
	for name := range q {
		if strings.HasPrefix(name, "x_") {
			delete(q, name)
		}
	}
}
