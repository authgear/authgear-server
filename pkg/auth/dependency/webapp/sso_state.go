package webapp

type SSOState map[string]string

func (c SSOState) RequestQuery() string {
	if s, ok := c["request_query"]; ok {
		return s
	}
	return ""
}

func (c SSOState) SetRequestQuery(s string) {
	c["request_query"] = s
}
