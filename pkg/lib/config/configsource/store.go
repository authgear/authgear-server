package configsource

import (
	globaldb "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
)

type Store struct {
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
}

func (s *Store) GetAppIDByDomain(domain string) (string, error) {
	return "", nil
}
