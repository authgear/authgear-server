package analytic

import (
	"fmt"
	"time"

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
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
}

func (r *UserWeeklyReport) Run(options *UserWeeklyReportOptions) (data *ReportData, err error) {
	rangeFormPtr, err := timeutil.FirstDayOfISOWeek(options.Year, options.Week, time.UTC)
	if err != nil {
		err = fmt.Errorf("invalid year or week number: %w", err)
		return
	}
	rangeFrom := *rangeFormPtr
	rangeTo := rangeFrom.AddDate(0, 0, 7)

	var appOwners []*AppCollaborator
	if err = r.GlobalHandle.WithTx(func() (e error) {
		appOwners, err = r.GlobalDBStore.GetAppOwners(&rangeFrom, &rangeTo)
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

	data = &ReportData{
		Header: []interface{}{
			"Year",
			"Week",
			"Range From",
			"Range To",
			"Number of new users",
			"Number of new users who have created project",
		},
		Values: [][]interface{}{
			entry,
		},
	}

	return
}
