package analytic

import (
	"time"
)

// ChartService provides method for the portal to get data for charts
type ChartService struct {
	AuditStore *AuditDBReadStore
}

func (s *ChartService) GetActiveUserChat(
	appID string,
	periodical string,
	rangeFrom time.Time,
	rangeTo time.Time,
) ([]*DataPoint, error) {
	return nil, nil
}
