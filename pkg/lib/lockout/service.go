package lockout

import (
	"context"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ServiceLogger = slogutil.NewLogger("lockout")

type Service struct {
	Storage Storage
}

func (s *Service) MakeAttempts(ctx context.Context, spec LockoutSpec, contributor string, attempts int) (result *MakeAttemptResult, err error) {
	logger := ServiceLogger.GetLogger(ctx)

	if !spec.Enabled {
		return &MakeAttemptResult{
			spec:        spec,
			LockedUntil: nil,
		}, nil
	}

	logger = logger.With(
		slog.String("key", spec.Key()),
	)

	isSuccess, lockedUntil, err := s.Storage.Update(ctx, spec, contributor, attempts)
	if err != nil {
		return nil, err
	}

	if lockedUntil != nil {
		logger = logger.With(
			slog.Time("lockedUntil", *lockedUntil),
		)
	}

	result = &MakeAttemptResult{
		spec:        spec,
		LockedUntil: lockedUntil,
	}

	if !isSuccess {
		logger.Debug(ctx, "make attempt failed")
		return result, result.ErrorIfLocked()
	}

	logger.Debug(ctx, "make attempt success")

	return result, nil
}

func (s *Service) ClearAttempts(ctx context.Context, spec LockoutSpec, contributor string) error {
	if !spec.Enabled {
		return nil
	}

	return s.Storage.Clear(ctx, spec, contributor)
}
