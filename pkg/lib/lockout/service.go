package lockout

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("lockout")}
}

type Service struct {
	Logger  Logger
	Storage Storage
}

func (l *Service) MakeAttempt(spec BucketSpec, attempts int) (lockedUntil *time.Time, err error) {
	if !spec.Enabled {
		return nil, nil
	}

	logger := l.Logger.
		WithField("key", spec.Key())

	isSuccess, lockedUntil, err := l.Storage.Update(spec, attempts)
	if err != nil {
		return nil, err
	}
	if lockedUntil != nil {
		logger = logger.
			WithField("lockedUntil", *lockedUntil)
	}
	if !isSuccess {
		logger.Debug("make attempt failed")
		return lockedUntil, NewErrLocked(spec.Name, *lockedUntil)
	}

	logger.Debug("make attempt success")

	return lockedUntil, nil
}
