package analytic

import (
	"errors"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type ProjectWeeklyReportOptions struct {
	Year        int
	Week        int
	PortalAppID string
}

type ProjectWeeklyReport struct {
	GlobalHandle      *globaldb.Handle
	GlobalDBStore     *GlobalDBStore
	AppDBHandle       *appdb.Handle
	AppDBStore        *AppDBStore
	AuditDBHandle     *auditdb.ReadHandle
	MeterAuditDBStore MeterAuditDBReadStore
	AuditDBStore      *AuditDBReadStore
}

func (r *ProjectWeeklyReport) Run(options *ProjectWeeklyReportOptions) (data *ReportData, err error) {
	rangeFormPtr, err := timeutil.FirstDayOfISOWeek(options.Year, options.Week, time.UTC)
	if err != nil {
		err = fmt.Errorf("invalid year or week number: %w", err)
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

	values := make([][]interface{}, len(appIDs))
	for i, appID := range appIDs {
		ownerID := appToOwnerUserIDMap[appID]
		email := userIDToEmailMap[ownerID]

		countMap := map[string]int{
			string(nonblocking.UserAuthenticated): 0,
			string(nonblocking.UserCreated):       0,
			string(nonblocking.EmailSent):         0,
			string(nonblocking.SMSSent):           0,
		}
		weeklyActiveUserCount := 0

		err = r.AuditDBHandle.ReadOnly(func() (e error) {
			for activityType := range countMap {
				countMap[activityType], err = r.MeterAuditDBStore.GetCountByActivityType(
					appID, activityType, &rangeFrom, &rangeTo)
				if err != nil {
					err = fmt.Errorf("failed to fetch count for activityType %s: %w", activityType, err)
					return err
				}
			}

			c, err := r.AuditDBStore.GetAnalyticCountByType(
				appID,
				WeeklyActiveUserCountType,
				&rangeFrom,
			)
			if err != nil {
				if !errors.Is(err, ErrAnalyticCountNotFound) {
					return fmt.Errorf("failed to fetch weekly active user: %w", err)
				}
			} else {
				weeklyActiveUserCount = c.Count
			}

			return nil
		})
		if err != nil {
			return
		}

		values[i] = []interface{}{
			options.Year,
			options.Week,
			rangeFrom.UTC().Format(time.RFC3339),
			rangeTo.UTC().Format(time.RFC3339),
			appID,
			email,
			countMap[string(nonblocking.UserCreated)],
			countMap[string(nonblocking.UserAuthenticated)],
			countMap[string(nonblocking.EmailSent)],
			countMap[string(nonblocking.SMSSent)],
			weeklyActiveUserCount,
		}
	}

	data = &ReportData{
		Header: []interface{}{
			"Year",
			"Week",
			"Range From",
			"Range To",
			"Project ID",
			"Owner email",
			"Number of signup",
			"Number of login",
			"Email sent",
			"SMS sent",
			"Weekly active user",
		},
		Values: values,
	}

	return
}
