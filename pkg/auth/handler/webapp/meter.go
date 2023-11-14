package webapp

import (
	"github.com/authgear/authgear-server/pkg/lib/meter"
)

type MeterService interface {
	TrackPageView(VisitorID string, pageType meter.PageType) error
}
