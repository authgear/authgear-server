package pq

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"

	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
	"github.com/skygeario/skygear-server/pkg/gateway/store"
)

func (s *Store) GetDomain(domain string) (*model.Domain, error) {
	builder := s.domainSelectBuilder().
		Where("domain = ?", domain)

	return s.domainQueryAndScan(builder)
}

func (s *Store) GetDefaultDomain(domain string) (*model.Domain, error) {
	builder := s.domainSelectBuilder().
		Where("domain = ?", domain).
		Where("assignment = 'default'")

	return s.domainQueryAndScan(builder)
}

func (s *Store) GetDomainByAppIDAndAssignment(appID string, assignment model.AssignmentType) (*model.Domain, error) {
	builder := s.domainSelectBuilder().
		Where("app_id = ?", appID).
		Where("assignment = ?", string(assignment))

	return s.domainQueryAndScan(builder)
}

func (s *Store) domainSelectBuilder() sq.SelectBuilder {
	return psql.Select(
		"id", "app_id", "domain", "assignment",
	).From(s.tableName("domain"))
}

func (s *Store) domainQueryAndScan(builder sq.SelectBuilder) (*model.Domain, error) {
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
