package lockout

import "github.com/authgear/authgear-server/pkg/util/log"

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("lockout")}
}

type Service struct {
	Logger  Logger
	Storage Storage
}

func (l *Service) MakeAttempt(spec BucketSpec, attempts int) error {
	if !spec.Enabled {
		return nil
	}

	logger := l.Logger.
		WithField("key", spec.Key())

	lockedUntil, err := l.Storage.Update(spec, attempts)
	if err != nil {
		return err
	}
	if lockedUntil != nil {
		logger.
			WithField("lockedUntil", *lockedUntil).Debug("locked out")
		return NewErrLocked(spec.Name, *lockedUntil)
	}

	logger.
		Debug("checked not locked out")

	return nil
}
