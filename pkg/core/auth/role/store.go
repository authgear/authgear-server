package role

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/utils"
)

type Store interface {
	CreateRoles(roles []string) error
	QueryRoles(roles []string) ([]string, error)
}

func EnsureRole(store Store, logger *logrus.Entry, roles []string) ([]string, error) {
	if roles == nil || len(roles) == 0 {
		return nil, nil
	}
	existedRole, err := store.QueryRoles(roles)
	if err != nil {
		return nil, err
	}
	if len(existedRole) == len(roles) {
		return nil, nil
	}
	logger.Debugf("Diffing the roles not exist in DB")
	absenceRoles := utils.StringSliceExcept(roles, existedRole)
	return absenceRoles, store.CreateRoles(absenceRoles)
}
