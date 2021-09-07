package analytic

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type projectWeeklyReportRecord struct {
	AppID       string
	OwnerUserID string
	OwnerEmail  string
}

type ProjectWeeklyReportOptions struct {
	Year        int
	Week        int
	PortalAppID string
}

type ProjectWeeklyReport struct {
	GlobalHandle  *globaldb.Handle
	GlobalDBStore *GlobalDBStore
	AppDBHandle   *appdb.Handle
	AppDBStore    *AppDBStore
}

func (r *ProjectWeeklyReport) Run(options *ProjectWeeklyReportOptions) (data *ReportData, err error) {
	rangeFormPtr, err := timeutil.FirstDayOfISOWeek(options.Year, options.Week, time.UTC)
	if err != nil {
		err = fmt.Errorf("invald year or week number: %w", err)
		return
	}
	rangeFrom := *rangeFormPtr
	rangeTo := rangeFrom.AddDate(0, 0, 7)

	var appIDs []string
	var appOwners []*AppCollaborator
	if err = r.GlobalHandle.WithTx(func() error {
		appIDs, err = r.GlobalDBStore.GetAppIDs()
		if err != nil {
			return fmt.Errorf("failed to fetch app ids: %w", err)
		}

		appOwners, err = r.GlobalDBStore.GetAppOwners(nil, nil)
		if err != nil {
			return fmt.Errorf("failed to fetch app owners: %w", err)
		}
		return nil
	}); err != nil {
		return
	}

	userIDs := []string{}
	userIDsSet := map[string]interface{}{}
	appToOwnerUserIDMap := map[string]string{}
	for _, ap := range appOwners {
		appToOwnerUserIDMap[ap.AppID] = ap.UserID
		if _, ok := userIDsSet[ap.UserID]; !ok {
			userIDsSet[ap.UserID] = struct{}{}
			userIDs = append(userIDs, ap.UserID)
		}
	}

	var userIDToEmailMap map[string]string
	if err = r.AppDBHandle.WithTx(func() (e error) {
		userIDToEmailMap, err = r.AppDBStore.GetUserVerifiedEmails(options.PortalAppID, userIDs)
		return err
	}); err != nil {
		err = fmt.Errorf("failed to fetch owner's email: %w", err)
		return
	}

	values := [][]interface{}{}
	records := make([]*projectWeeklyReportRecord, len(appIDs))
	for i, appID := range appIDs {
		ownerID := appToOwnerUserIDMap[appID]
		email := userIDToEmailMap[ownerID]
		records[i] = &projectWeeklyReportRecord{
			AppID:       appID,
			OwnerUserID: ownerID,
			OwnerEmail:  email,
		}

		values = append(values, []interface{}{
			options.Year,
			options.Week,
			rangeFrom.UTC().Format(time.RFC3339),
			rangeTo.UTC().Format(time.RFC3339),
			appID,
			email,
		})
	}

	data = &ReportData{
		Header: []interface{}{
			"Year",
			"Week",
			"Range From",
			"Range To",
			"Project ID",
			"Owner email",
		},
		Values: values,
	}

	return
}
