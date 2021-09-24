package analytic

import (
	"fmt"
	"math"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	periodicalutil "github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type Chart struct {
	DataSet []*DataPoint `json:"dataset"`
}

type SignupConversionRateData struct {
	TotalSignup               int     `json:"totalSignup"`
	TotalSignupUniquePageView int     `json:"totalSignupUniquePageView"`
	ConversionRate            float64 `json:"conversionRate"`
}

// ChartService provides method for the portal to get data for charts
type ChartService struct {
	Database       *auditdb.ReadHandle
	AuditStore     *AuditDBReadStore
	Clock          clock.Clock
	AnalyticConfig *config.AnalyticConfig
}

func (s *ChartService) GetActiveUserChat(
	appID string,
	periodical string,
	rangeFrom time.Time,
	rangeTo time.Time,
) (*Chart, error) {
	if s.Database == nil {
		return &Chart{}, nil
	}

	countType := ""
	periodicalType := periodicalutil.Type(periodical)
	switch periodicalType {
	case periodicalutil.Weekly:
		countType = WeeklyActiveUserCountType
	case periodicalutil.Monthly:
		countType = MonthlyActiveUserCountType
	default:
		return nil, fmt.Errorf("unknown periodical: %s", periodical)
	}

	rangeFrom, rangeTo, err := s.GetBoundedRange(periodicalType, rangeFrom, rangeTo)
	if err != nil {
		// invalid range, return empty chart
		return &Chart{}, nil
	}

	dataset, err := s.getDataPointsByCountType(appID, countType, periodicalType, rangeFrom, rangeTo)
	if err != nil {
		return nil, err
	}

	return &Chart{
		DataSet: dataset,
	}, nil
}

func (s *ChartService) GetTotalUserCountChat(appID string, rangeFrom time.Time, rangeTo time.Time) (*Chart, error) {
	if s.Database == nil {
		return &Chart{}, nil
	}

	rangeFrom, rangeTo, err := s.GetBoundedRange(periodicalutil.Daily, rangeFrom, rangeTo)
	if err != nil {
		// invalid range, return empty chart
		return &Chart{}, nil
	}

	dataset, err := s.getDataPointsByCountType(appID, CumulativeUserCountType, periodicalutil.Daily, rangeFrom, rangeTo)
	if err != nil {
		return nil, err
	}
	return &Chart{
		DataSet: dataset,
	}, nil
}

func (s *ChartService) GetSignupConversionRate(appID string, rangeFrom time.Time, rangeTo time.Time) (*SignupConversionRateData, error) {
	if s.Database == nil {
		return &SignupConversionRateData{}, nil
	}

	rangeFrom, rangeTo, err := s.GetBoundedRange(periodicalutil.Daily, rangeFrom, rangeTo)
	if err != nil {
		// invalid range, return empty chart
		return &SignupConversionRateData{}, nil
	}

	totalSignupCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailySignupCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total signup count: %w", err)
	}

	totalSignupUniquePageCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailySignupUniquePageViewCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total signup unique page view count: %w", err)
	}

	conversionRate := float64(0)
	if totalSignupUniquePageCount > 0 {
		rate := float64(totalSignupCount) / float64(totalSignupUniquePageCount)
		conversionRate = math.Round(rate*100*100) / 100
	}

	return &SignupConversionRateData{
		TotalSignup:               totalSignupCount,
		TotalSignupUniquePageView: totalSignupUniquePageCount,
		ConversionRate:            conversionRate,
	}, nil
}

func (s *ChartService) GetSignupByMethodsChart(appID string, rangeFrom time.Time, rangeTo time.Time) (*Chart, error) {
	if s.Database == nil {
		return &Chart{}, nil
	}

	rangeFrom, rangeTo, err := s.GetBoundedRange(periodicalutil.Daily, rangeFrom, rangeTo)
	if err != nil {
		// invalid range, return empty chart
		return &Chart{}, nil
	}

	// SignupByMethodsChart are the data points for signup by method pie chart
	signupByMethodsChart := []*DataPoint{}
	for _, method := range DailySignupCountTypeByMethods {
		c, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, method.CountType, &rangeFrom, &rangeTo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch signup count for method: %s: %w", method.MethodName, err)
		}
		signupByMethodsChart = append(signupByMethodsChart, &DataPoint{
			Label: method.MethodName,
			Data:  c,
		})
	}
	return &Chart{
		DataSet: signupByMethodsChart,
	}, nil
}

// GetBoundedRange returns if the given range is valid and the bounded range
// The range is bounded by the analytic epoch ane the current date
func (s *ChartService) GetBoundedRange(
	periodical periodicalutil.Type,
	rangeFrom time.Time,
	rangeTo time.Time,
) (newRangeFrom time.Time, newRangeTo time.Time, err error) {
	today := timeutil.TruncateToDate(s.Clock.NowUTC())
	newRangeFrom = rangeFrom
	newRangeTo = rangeTo
	if !s.AnalyticConfig.Epoch.IsZero() {
		epoch := time.Time(s.AnalyticConfig.Epoch)
		if newRangeFrom.Before(epoch) {
			newRangeFrom = epoch
		}
	}

	var limitRangeTo time.Time
	switch periodical {
	case periodicalutil.Weekly:
		// adjust range to monday
		newRangeFrom = timeutil.MondayOfTheWeek(newRangeFrom)
		newRangeTo = timeutil.MondayOfTheWeek(newRangeTo)
		// monday of last week
		limitRangeTo = timeutil.MondayOfTheWeek(today.AddDate(0, 0, -7))
	case periodicalutil.Monthly:
		// adjust range to first day of the month
		newRangeFrom = timeutil.FirstDayOfTheMonth(newRangeFrom)
		newRangeTo = timeutil.FirstDayOfTheMonth(newRangeTo)
		// first day of last month
		limitRangeTo = timeutil.FirstDayOfTheMonth(today.AddDate(0, -1, 0))
	case periodicalutil.Daily:
		// yesterday
		limitRangeTo = today.AddDate(0, 0, -1)
	default:
		panic(fmt.Sprintf("unknown periodical: %s", periodical))
	}
	if newRangeTo.After(limitRangeTo) {
		newRangeTo = limitRangeTo
	}

	if newRangeFrom.After(newRangeTo) {
		err = fmt.Errorf("invalid range")
		return
	}

	return
}

func (s *ChartService) getDataPointsByCountType(
	appID string,
	countType string,
	periodical periodicalutil.Type,
	rangeFrom time.Time,
	rangeTo time.Time,
) ([]*DataPoint, error) {
	counts, err := s.AuditStore.GetAnalyticCountsByType(appID, countType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, err
	}

	countsMap := map[string]int{}
	dataPoints := []*DataPoint{}
	for _, c := range counts {
		countsMap[c.Date.Format(timeutil.LayoutISODate)] = c.Count
	}

	dateLists := GetDateListByRangeInclusive(rangeFrom, rangeTo, periodical)
	for _, date := range dateLists {
		dateStr := date.Format(timeutil.LayoutISODate)
		count := countsMap[dateStr]
		dataPoints = append(dataPoints, &DataPoint{
			Label: dateStr,
			Data:  count,
		})
	}

	return dataPoints, nil
}
