package httputil

import (
	"fmt"
)

type HTTPOrigin string

func MakeHTTPOrigin(proto HTTPProto, host HTTPHost) HTTPOrigin {
	return HTTPOrigin(fmt.Sprintf("%v://%v", proto, host))
}
