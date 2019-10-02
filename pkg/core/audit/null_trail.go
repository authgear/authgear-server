package audit

import "net/http"

type NullTrail struct{}

func (t NullTrail) WithRequest(request *http.Request) Trail { return t }
func (t NullTrail) Log(entry Entry)                         {}
