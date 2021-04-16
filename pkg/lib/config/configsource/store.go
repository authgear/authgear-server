package configsource

import (
	"database/sql"
	"errors"

	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
)

type Store struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *Store) GetAppIDByDomain(domain string) (string, error) {
	builder := s.SQLBuilder.Global().
		Select(
			"app_id",
		).
		From(s.SQLBuilder.TableName("_portal_domain")).
		Where("domain = ?", domain)

	scanner, err := s.SQLExecutor.QueryRowWith(builder)
	if err != nil {
		return "", err
	}

	var appID string
	if err = scanner.Scan(&appID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrAppNotFound
		}
		return "", err
	}

	return appID, nil
}
