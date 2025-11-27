package netutil

import (
	"net/http"
)

type CrossOriginProtection struct{}

func NewCrossOriginProtection() *CrossOriginProtection {
	return &CrossOriginProtection{}
}

func (m *CrossOriginProtection) Check(r *http.Request) error {
	return http.NewCrossOriginProtection().Check(r)
}
