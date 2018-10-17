package cloud

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// cloudStoreSigner models the signer of skygear cloud asset store
type cloudStoreSigner struct {
	token         string
	extra         string
	expiredAt     time.Time
	refreshTicker *time.Ticker
	mutex         *sync.RWMutex
}

func newCloudStoreSigner(
	refreshInterval time.Duration,
	refreshTickerFunc func(),
	logger *logrus.Entry,
) *cloudStoreSigner {

	s := &cloudStoreSigner{
		mutex:         &sync.RWMutex{},
		refreshTicker: time.NewTicker(refreshInterval),
	}

	go func() {
		for tickerTime := range s.refreshTicker.C {
			logger.
				WithField("time", tickerTime).
				Info("Cloud Asset Signer Refresh Ticker Trigger")
			refreshTickerFunc()
		}
	}()

	return s
}

func (s *cloudStoreSigner) update(token, extra string, expiredAt time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.token = token
	s.extra = extra
	s.expiredAt = expiredAt
}

func (s cloudStoreSigner) get() (token, extra string, expiredAt time.Time) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.token, s.extra, s.expiredAt
}
