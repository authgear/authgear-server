package otp

import (
	"context"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/clock"

	. "github.com/smartystreets/goconvey/convey"
)

type testCodeStore struct {
	codes map[string]*Code
}

func newTestCodeStore() *testCodeStore {
	return &testCodeStore{codes: map[string]*Code{}}
}

func (s *testCodeStore) key(p Purpose, target string) string {
	return string(p) + ":" + target
}

func (s *testCodeStore) Create(ctx context.Context, purpose Purpose, code *Code) error {
	s.codes[s.key(purpose, code.Target)] = code
	return nil
}

func (s *testCodeStore) Get(ctx context.Context, purpose Purpose, target string) (*Code, error) {
	code, ok := s.codes[s.key(purpose, target)]
	if !ok {
		return nil, ErrCodeNotFound
	}
	return code, nil
}

func (s *testCodeStore) Update(ctx context.Context, purpose Purpose, code *Code) error {
	s.codes[s.key(purpose, code.Target)] = code
	return nil
}

func (s *testCodeStore) Delete(ctx context.Context, purpose Purpose, target string) error {
	delete(s.codes, s.key(purpose, target))
	return nil
}

type testLookupStore struct{}

func (s *testLookupStore) Create(ctx context.Context, purpose Purpose, code string, target string, expireAt time.Time) error {
	return nil
}

func (s *testLookupStore) Get(ctx context.Context, purpose Purpose, code string) (string, error) {
	return "", ErrCodeNotFound
}

func (s *testLookupStore) Delete(ctx context.Context, purpose Purpose, code string) error {
	return nil
}

type testAttemptTracker struct{}

func (s *testAttemptTracker) ResetFailedAttempts(ctx context.Context, kind Kind, target string) error {
	return nil
}

func (s *testAttemptTracker) GetFailedAttempts(ctx context.Context, kind Kind, target string) (int, error) {
	return 0, nil
}

func (s *testAttemptTracker) IncrementFailedAttempts(ctx context.Context, kind Kind, target string) (int, error) {
	return 1, nil
}

type testRateLimiter struct {
	usedKeys       map[string]int
	reservedKeys   map[*ratelimit.Reservation]string
	oneShotBuckets map[ratelimit.BucketName]bool
	timeToAct      map[string]time.Time
}

func newTestRateLimiter() *testRateLimiter {
	return &testRateLimiter{
		usedKeys:     map[string]int{},
		reservedKeys: map[*ratelimit.Reservation]string{},
		oneShotBuckets: map[ratelimit.BucketName]bool{
			ratelimit.VerificationCooldownSMS:                  true,
			ratelimit.VerificationCooldownWhatsapp:             true,
			ratelimit.VerificationCooldownSMSPerSession:        true,
			ratelimit.VerificationCooldownWhatsappPerSession:   true,
			ratelimit.OOBOTPCooldownSMS:                        true,
			ratelimit.OOBOTPCooldownWhatsapp:                   true,
			ratelimit.OOBOTPCooldownSMSPerSession:              true,
			ratelimit.OOBOTPCooldownWhatsappPerSession:         true,
			ratelimit.ForgotPasswordCooldownSMS:                true,
			ratelimit.ForgotPasswordCooldownWhatsapp:           true,
			ratelimit.ForgotPasswordCooldownSMSPerSession:      true,
			ratelimit.ForgotPasswordCooldownWhatsappPerSession: true,
		},
		timeToAct: map[string]time.Time{},
	}
}

func (l *testRateLimiter) GetTimeToAct(ctx context.Context, spec ratelimit.BucketSpec) (*time.Time, error) {
	if t, ok := l.timeToAct[spec.Key()]; ok {
		return &t, nil
	}
	zero := time.Unix(0, 0).UTC()
	return &zero, nil
}

func (l *testRateLimiter) Allow(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.FailedReservation, error) {
	if l.oneShotBuckets[spec.Name] {
		if l.usedKeys[spec.Key()] > 0 {
			return ratelimit.NewFailedReservation(spec), nil
		}
		l.usedKeys[spec.Key()]++
	}
	return nil, nil
}

func (l *testRateLimiter) Reserve(ctx context.Context, spec ratelimit.BucketSpec) (*ratelimit.Reservation, *ratelimit.FailedReservation, error) {
	if l.oneShotBuckets[spec.Name] {
		if l.usedKeys[spec.Key()] > 0 {
			return nil, ratelimit.NewFailedReservation(spec), nil
		}
		l.usedKeys[spec.Key()]++
	}
	resv := &ratelimit.Reservation{}
	l.reservedKeys[resv] = spec.Key()
	return resv, nil, nil
}

func (l *testRateLimiter) Cancel(ctx context.Context, r *ratelimit.Reservation) {
	if r == nil {
		return
	}
	key, ok := l.reservedKeys[r]
	if !ok {
		return
	}
	if l.usedKeys[key] > 0 {
		l.usedKeys[key]--
	}
	delete(l.reservedKeys, r)
}

func loadTestAppConfig() *config.AppConfig {
	boolPtr := func(v bool) *bool { return &v }
	return &config.AppConfig{
		Verification: &config.VerificationConfig{
			CodeValidPeriod: "5m",
			RateLimits: &config.VerificationRateLimitsConfig{
				Email: &config.VerificationRateLimitsEmailConfig{
					TriggerCooldown: "1m",
					TriggerPerIP:    &config.RateLimitConfig{Enabled: boolPtr(false)},
					TriggerPerUser:  &config.RateLimitConfig{Enabled: boolPtr(false)},
					ValidatePerIP:   &config.RateLimitConfig{Enabled: boolPtr(false)},
				},
				SMS: &config.VerificationRateLimitsSMSConfig{
					TriggerCooldown: "1m",
					TriggerPerIP:    &config.RateLimitConfig{Enabled: boolPtr(false)},
					TriggerPerUser:  &config.RateLimitConfig{Enabled: boolPtr(false)},
					ValidatePerIP:   &config.RateLimitConfig{Enabled: boolPtr(false)},
				},
			},
		},
	}
}

func newTestService(rateLimiter *testRateLimiter) *Service {
	return &Service{
		Clock: clock.NewMockClockAtTime(time.Unix(1700000000, 0).UTC()),
		AppID: "app",
		TestModeConfig: &config.TestModeConfig{
			FixedOOBOTP: &config.TestModeOOBOTPConfig{},
			Email:       &config.TestModeEmailConfig{},
		},
		TestModeFeatureConfig: &config.TestModeFeatureConfig{
			FixedOOBOTP:          &config.TestModeFixedOOBOTPFeatureConfig{},
			DeterministicLinkOTP: &config.TestModeDeterministicLinkOTPFeatureConfig{},
		},
		CodeStore:      newTestCodeStore(),
		LookupStore:    &testLookupStore{},
		AttemptTracker: &testAttemptTracker{},
		RateLimiter:    rateLimiter,
		FeatureConfig:  &config.FeatureConfig{},
		EnvConfig:      &config.RateLimitsEnvironmentConfig{},
	}
}

func TestGenerateOTPWithSessionCooldown(t *testing.T) {
	Convey("GenerateOTP applies authflow session cooldown independently from target cooldown", t, func() {
		cfg := loadTestAppConfig()
		rateLimiter := newTestRateLimiter()
		svc := newTestService(rateLimiter)
		kindSMS := KindVerification(cfg, "sms")
		kindWhatsapp := KindVerification(cfg, "whatsapp")

		_, err := svc.GenerateOTP(context.Background(), kindSMS, "+85265000001", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(err, ShouldBeNil)

		_, err = svc.GenerateOTP(context.Background(), kindSMS, "+85265000002", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(ratelimit.IsRateLimitErrorWithBucketName(err, ratelimit.VerificationCooldownSMSPerSession), ShouldBeTrue)

		_, err = svc.GenerateOTP(context.Background(), kindSMS, "+85265000003", FormCode, &GenerateOptions{})
		So(err, ShouldBeNil)

		_, err = svc.GenerateOTP(context.Background(), kindWhatsapp, "+85265000004", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(err, ShouldBeNil)
	})
}

func TestGenerateOTPWithSessionCooldownDoesNotConsumeTargetCooldownOnFailure(t *testing.T) {
	Convey("GenerateOTP cancels target cooldown reservation when session cooldown is blocked", t, func() {
		cfg := loadTestAppConfig()
		rateLimiter := newTestRateLimiter()
		svc := newTestService(rateLimiter)
		kind := KindVerification(cfg, "sms")

		_, err := svc.GenerateOTP(context.Background(), kind, "+85265000001", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(err, ShouldBeNil)

		_, err = svc.GenerateOTP(context.Background(), kind, "+85265000002", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(ratelimit.IsRateLimitErrorWithBucketName(err, ratelimit.VerificationCooldownSMSPerSession), ShouldBeTrue)

		_, err = svc.GenerateOTP(context.Background(), kind, "+85265000002", FormCode, &GenerateOptions{
			AuthenticationFlowID: "flow-2",
		})
		So(err, ShouldBeNil)
	})
}

func TestInspectStateWithSessionCooldown(t *testing.T) {
	Convey("InspectState returns the later of target and authflow session cooldown", t, func() {
		cfg := loadTestAppConfig()
		rateLimiter := newTestRateLimiter()
		svc := newTestService(rateLimiter)
		kind := KindVerification(cfg, "sms")
		targetSpec := kind.RateLimitTriggerCooldown("+85265000001")
		sessionSpec := kind.RateLimitTriggerCooldownPerSession("flow-1")
		targetTime := time.Unix(1700000010, 0).UTC()
		sessionTime := time.Unix(1700000100, 0).UTC()
		rateLimiter.timeToAct[targetSpec.Key()] = targetTime
		rateLimiter.timeToAct[sessionSpec.Key()] = sessionTime

		state, err := svc.InspectState(context.Background(), kind, "+85265000001", &InspectStateOptions{
			AuthenticationFlowID: "flow-1",
		})
		So(err, ShouldBeNil)
		So(state.CanResendAt, ShouldResemble, sessionTime)

		state, err = svc.InspectState(context.Background(), kind, "+85265000001", nil)
		So(err, ShouldBeNil)
		So(state.CanResendAt, ShouldResemble, targetTime)
	})
}
