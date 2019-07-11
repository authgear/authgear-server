package password

import "github.com/skygeario/skygear-server/pkg/core/utils"

const DefaultRealm string = "default"

type realmChecker interface {
	isValid(realm string) bool
}

type defaultRealmChecker struct {
	allowedRealms []string
}

func (c defaultRealmChecker) isValid(realm string) bool {
	return utils.StringSliceContains(c.allowedRealms, realm)
}

var (
	_ realmChecker = &defaultRealmChecker{}
)
