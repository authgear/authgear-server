package plan

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type Service struct {
	Handle *globaldb.Handle
	Store  *plan.Store
}

func (s *Service) CreatePlan(name string) error {
	return s.Handle.WithTx(func() (err error) {
		p := model.NewPlan(name)
		return s.Store.Create(p)
	})
}
