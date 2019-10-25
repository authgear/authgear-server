package skyerr

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

const (
	APIErrorDetail errors.DetailTag = "api"
	TenantDetail   errors.DetailTag = "tenant"
)

type APIErrorString string
type TenantString string

func (APIErrorString) IsTagged(tag errors.DetailTag) bool { return tag == APIErrorDetail }
func (TenantString) IsTagged(tag errors.DetailTag) bool   { return tag == TenantDetail }
