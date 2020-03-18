package protocol

type ErrorResponse map[string]string

func (r ErrorResponse) Error(v string)            { r["error"] = v }
func (r ErrorResponse) ErrorDescription(v string) { r["error_description"] = v }
func (r ErrorResponse) State(v string)            { r["state"] = v }
