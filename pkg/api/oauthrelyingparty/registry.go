package oauthrelyingparty

import (
	"fmt"
)

var registry = map[string]Provider{}

func RegisterProvider(typ string, provider Provider) {
	_, ok := registry[typ]
	if ok {
		panic(fmt.Errorf("oauth provider is already registered: %v", typ))
	}
	registry[typ] = provider
}
