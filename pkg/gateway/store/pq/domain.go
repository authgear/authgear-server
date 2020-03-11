package pq

import (
	"database/sql"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func (s *Store) GetDomain(domain string) (*model.Domain, error) {
	builder := psql.Select(
		"id", "app_id", "domain", "assignment",
	).From(s.tableName("domain")).
		Where("domain = ?", domain)
	scanner, err := s.QueryRowWith(builder)
	if err != nil {
		return nil, err
	}

	d := &model.Domain{}
	err = scanner.StructScan(
		d,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.NewNotFoundError("domain")
	}
	if err != nil {
		return nil, err
	}

	return d, nil
}
