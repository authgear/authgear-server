package tester

import "net/url"

type EndpointsProvider interface {
	TesterURL() *url.URL
}
