package declarative

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/fraudprotection"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type testFraudProtectionMetrics struct{}

func (m *testFraudProtectionMetrics) RecordVerified(_ context.Context, _, _ string) error { return nil }
func (m *testFraudProtectionMetrics) RecordUnverifiedSMSOTPCountDrained(_ context.Context, _, _ string, _ int) error {
	return nil
}
func (m *testFraudProtectionMetrics) GetVerifiedByCountry24h(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *testFraudProtectionMetrics) GetVerifiedByCountry1h(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *testFraudProtectionMetrics) GetVerifiedByIP24h(_ context.Context, _ string) (int64, error) {
	return 0, nil
}
func (m *testFraudProtectionMetrics) GetVerifiedByCountryPast14DaysRollingMax(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

type testLeakyBucketDrainCall struct {
	phoneCountry string
	count        int
}

type testFraudProtectionLeakyBucket struct {
	drainCalls []testLeakyBucketDrainCall
}

func (l *testFraudProtectionLeakyBucket) RecordUnverifiedSMSOTPSent(_ context.Context, _, _ string, _ fraudprotection.LeakyBucketThresholds) (fraudprotection.LeakyBucketTriggered, fraudprotection.LeakyBucketLevels, error) {
	return fraudprotection.LeakyBucketTriggered{}, fraudprotection.LeakyBucketLevels{}, nil
}
func (l *testFraudProtectionLeakyBucket) DrainUnverifiedSMSOTPSent(_ context.Context, _, phoneCountry string, _ fraudprotection.LeakyBucketThresholds, count int) error {
	l.drainCalls = append(l.drainCalls, testLeakyBucketDrainCall{phoneCountry: phoneCountry, count: count})
	return nil
}
func (l *testFraudProtectionLeakyBucket) RecordSMSOTPVerifiedCountry(_ context.Context, _, _ string) error {
	return nil
}

func TestRevertUnverifiedSMSOTPs(t *testing.T) {
	Convey("revertUnverifiedSMSOTPs", t, func() {
		cfg := &config.FraudProtectionConfig{}
		config.SetFieldDefaults(cfg)

		leakyBucket := &testFraudProtectionLeakyBucket{}
		deps := &authflow.Dependencies{
			HTTPRequest: &http.Request{Header: http.Header{}},
			FraudProtection: &fraudprotection.Service{
				Config:      cfg,
				RemoteIP:    httputil.RemoteIP("1.2.3.4"),
				Metrics:     &testFraudProtectionMetrics{},
				LeakyBucket: leakyBucket,
			},
		}

		session := &authflow.Session{
			SMSOTPSentCountByPhone: map[string]int{
				"+6591230001": 3,
				"+6591230002": 1,
			},
			SMSOTPVerifiedCountByPhone: map[string]int{
				"+6591230001": 1,
				"+6591230002": 1,
			},
		}
		ctx := session.MakeContext(context.Background(), deps)

		root := &authflow.Flow{}
		effect := revertUnverifiedSMSOTPs(authflow.NewFlows(root))
		onCommit, ok := effect.(authflow.OnCommitEffect)
		So(ok, ShouldBeTrue)

		err := onCommit(ctx, deps)
		So(err, ShouldBeNil)
		So(leakyBucket.drainCalls, ShouldHaveLength, 1)
		So(leakyBucket.drainCalls[0], ShouldResemble, testLeakyBucketDrainCall{
			phoneCountry: "SG",
			count:        2,
		})
	})
}
