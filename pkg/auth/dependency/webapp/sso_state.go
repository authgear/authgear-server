package webapp

type SSOState map[string]string

func (c SSOState) CallbackURL() string {
	if s, ok := c["callback_url"]; ok {
		return s
	}
	return ""
}

func (c SSOState) SetCallbackURL(s string) {
	c["callback_url"] = s
}

func (c SSOState) WebAppStateID() string {
	if s, ok := c["webapp_sid"]; ok {
		return s
	}
	return ""
}

func (c SSOState) SetWebAppStateID(s string) {
	c["webapp_sid"] = s
}
