package analytic

import (
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type UserWeeklyReport struct {
	Handle        *globaldb.Handle
	GlobalDBStore *analytic.GlobalDBStore
}

func (r *UserWeeklyReport) Run(year int, week int) {
	// TODO: implement the report
}
