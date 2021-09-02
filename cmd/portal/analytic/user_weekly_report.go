package analytic

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type UserWeeklyReportOptions struct {
	Year        int
	Week        int
	PortalAppID string
}

type UserWeeklyReport struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *analytic.GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *analytic.AppDBStore
}

func (r *UserWeeklyReport) Run(options *UserWeeklyReportOptions) (err error) {
	rangeFrom := timeutil.FirstDayOfISOWeek(options.Year, options.Week, time.UTC)
	rangeTo := rangeFrom.AddDate(0, 0, 7)

	var appOwners []*analytic.AppCollaborator
	if err = r.GlobalHandle.WithTx(func() (e error) {
		appOwners, err = r.GlobalDBStore.GetNewAppOwners(&rangeFrom, &rangeTo)
		return err
	}); err != nil {
		err = fmt.Errorf("failed to fetch new apps: %w", err)
		return
	}

	var newUserIDs []string
	if err = r.AppDBHandle.WithTx(func() (e error) {
		newUserIDs, err = r.AppDBStore.GetNewUserIDs(options.PortalAppID, &rangeFrom, &rangeTo)
		return err
	}); err != nil {
		err = fmt.Errorf("failed to fetch new apps: %w", err)
		return
	}

	// userIDSet store the users who have created app
	userIDsSet := map[string]interface{}{}
	for _, perAppOwners := range appOwners {
		userIDsSet[perAppOwners.UserID] = struct{}{}
	}

	// count number of new users who have created app
	haveCreatedAppCount := 0
	for _, userID := range newUserIDs {
		if _, ok := userIDsSet[userID]; ok {
			haveCreatedAppCount++
		}
	}

	entry := []interface{}{
		options.Year,
		options.Week,
		rangeFrom.UTC().Format(time.RFC3339),
		rangeTo.UTC().Format(time.RFC3339),
		len(newUserIDs),
		haveCreatedAppCount,
	}

	// TODO: support output type
	fmt.Println("entry", entry)

	return nil
}
