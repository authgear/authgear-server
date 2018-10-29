package pq

import (
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type safeAuthInfoStore struct {
	impl      authinfo.Store
	txContext db.SafeTxContext
}

func NewSafeAuthInfoStore(
	builder db.SQLBuilder,
	executor db.SQLExecutor,
	logger *logrus.Entry,
	txContext db.SafeTxContext,
) authinfo.Store {
	return &safeAuthInfoStore{
		impl:      NewAuthInfoStore(builder, executor, logger),
		txContext: txContext,
	}
}

func (s *safeAuthInfoStore) CreateAuth(authinfo *authinfo.AuthInfo) error {
	s.txContext.EnsureTx()
	return s.impl.CreateAuth(authinfo)
}

func (s *safeAuthInfoStore) GetAuth(id string, authinfo *authinfo.AuthInfo) error {
	s.txContext.EnsureTx()
	return s.impl.GetAuth(id, authinfo)
}

func (s *safeAuthInfoStore) UpdateAuth(authinfo *authinfo.AuthInfo) error {
	s.txContext.EnsureTx()
	return s.impl.UpdateAuth(authinfo)
}

func (s *safeAuthInfoStore) DeleteAuth(id string) error {
	s.txContext.EnsureTx()
	return s.impl.DeleteAuth(id)
}

func (s *safeAuthInfoStore) AssignRoles(userIDs []string, roles []string) error {
	s.txContext.EnsureTx()
	return s.impl.AssignRoles(userIDs, roles)
}

func (s *safeAuthInfoStore) GetRoles(userIDs []string) (map[string][]string, error) {
	s.txContext.EnsureTx()
	return s.impl.GetRoles(userIDs)
}

func (s *safeAuthInfoStore) RevokeRoles(userIDs []string, roles []string) error {
	s.txContext.EnsureTx()
	return s.impl.RevokeRoles(userIDs, roles)
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authinfo.Store = &safeAuthInfoStore{}
)
