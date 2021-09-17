package analytic

import (
	"fmt"
	"time"

	periodicalutil "github.com/authgear/authgear-server/pkg/util/periodical"
)

type Chart struct {
	DataSet []*DataPoint `json:"dataset"`
}

// ChartService provides method for the portal to get data for charts
type ChartService struct {
	AuditStore *AuditDBReadStore
}

func (s *ChartService) GetActiveUserChat(
	appID string,
	periodical string,
	rangeFrom time.Time,
	rangeTo time.Time,
) (*Chart, error) {
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
		countsMap[c.Date.Format("2006-01-02")] = c.Count
	}

	dateLists := GetDateListByRangeInclusive(rangeFrom, rangeTo, periodical)
	for _, date := range dateLists {
		dateStr := date.Format("2006-01-02")
		count := countsMap[dateStr]
		dataPoints = append(dataPoints, &DataPoint{
			Label: dateStr,
			Data:  count,
		})
	}

	return dataPoints, nil
}
