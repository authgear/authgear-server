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

type SignupSummary struct {
	TotalUserCount            int     `json:"totalUserCount"`
	TotalSignup               int     `json:"totalSignup"`
	TotalSignupPageCount      int     `json:"totalSignupPageCount"`
	TotalSignupUniquePageView int     `json:"totalSignupUniquePageView"`
	TotalLoginPageView        int     `json:"totalLoginPageView"`
	TotalLoginUniquePageView  int     `json:"totalLoginUniquePageView"`
	ConversionRate            float64 `json:"conversionRate"`
	SignupByChannelChart      *Chart  `json:"signupByChannelChart"`
	TotalUserCountChart       *Chart  `json:"totalUserCountChart"`
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

func (s *ChartService) GetSignupSummary(
	appID string,
	rangeFrom time.Time,
	rangeTo time.Time,
) (*SignupSummary, error) {
	if s.Database == nil {
		return &SignupSummary{
			SignupByChannelChart: &Chart{},
			TotalUserCountChart:  &Chart{},
		}, nil
	}
	var err error
	totalUserCounts, err := s.getDataPointsByCountType(appID, CumulativeUserCountType, periodicalutil.Daily, rangeFrom, rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total user count")
	}

	totalUserCount := 0
	if len(totalUserCounts) > 0 {
		totalUserCount = totalUserCounts[len(totalUserCounts)-1].Data
	}

	totalSignupCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailySignupCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total signup count: %w", err)
	}

	totalSignupPageCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailySignupPageViewCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total signup page view count: %w", err)
	}

	totalSignupUniquePageCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailySignupUniquePageViewCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total signup unique page view count: %w", err)
	}

	totalLoginPageCount, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailyLoginPageViewCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total login page view count: %w", err)
	}

	totalLoginUniquePageView, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, DailyLoginUniquePageViewCountType, &rangeFrom, &rangeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch total login unique page view count: %w", err)
	}

	conversionRate := float64(0)
	if totalSignupUniquePageCount > 0 {
		rate := float64(totalSignupCount) / float64(totalLoginUniquePageView)
		conversionRate = math.Round(rate*100*100) / 100
	}

	// SignupByChannelChart are the data points for signup by channel pie chart
	signupByChannelChart := []*DataPoint{}
	for _, channel := range DailySignupCountTypeByChannels {
		c, err := s.AuditStore.GetSumOfAnalyticCountsByType(appID, channel.CountType, &rangeFrom, &rangeTo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch signup count for channel: %s: %w", channel.ChannelName, err)
		}
		signupByChannelChart = append(signupByChannelChart, &DataPoint{
			Label: channel.ChannelName,
			Data:  c,
		})
	}

	return &SignupSummary{
		TotalUserCount:            totalUserCount,
		TotalSignup:               totalSignupCount,
		TotalSignupPageCount:      totalSignupPageCount,
		TotalSignupUniquePageView: totalSignupUniquePageCount,
		TotalLoginPageView:        totalLoginPageCount,
		TotalLoginUniquePageView:  totalLoginUniquePageView,
		ConversionRate:            conversionRate,
		SignupByChannelChart: &Chart{
			DataSet: signupByChannelChart,
		},
		TotalUserCountChart: &Chart{
			DataSet: totalUserCounts,
		},
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
