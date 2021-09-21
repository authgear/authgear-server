package analytic

import (
	"fmt"
	"math"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
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
	Database   *auditdb.ReadHandle
	AuditStore *AuditDBReadStore
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
