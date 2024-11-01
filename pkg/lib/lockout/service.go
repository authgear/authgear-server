package lockout

import (
	"context"

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

func (s *Service) MakeAttempts(ctx context.Context, spec LockoutSpec, contributor string, attempts int) (result *MakeAttemptResult, err error) {
	if !spec.Enabled {
		return &MakeAttemptResult{
			spec:        spec,
			LockedUntil: nil,
		}, nil
	}

	logger := s.Logger.
		WithField("key", spec.Key())

	isSuccess, lockedUntil, err := s.Storage.Update(ctx, spec, contributor, attempts)
	if err != nil {
		return nil, err
	}
	if lockedUntil != nil {
		logger = logger.
			WithField("lockedUntil", *lockedUntil)
	}

	result = &MakeAttemptResult{
		spec:        spec,
		LockedUntil: lockedUntil,
	}

	if !isSuccess {
		logger.Debug("make attempt failed")
		return result, result.ErrorIfLocked()
	}

	logger.Debug("make attempt success")

	return result, nil
}

func (s *Service) ClearAttempts(ctx context.Context, spec LockoutSpec, contributor string) error {
	if !spec.Enabled {
		return nil
	}

	return s.Storage.Clear(ctx, spec, contributor)
}
