package fraudprotection

import (
	"context"
	"math"
	"net"
	"regexp"
	"slices"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

var ErrBlockedByFraudProtection = apierrors.Forbidden.WithReason("BlockedByFraudProtection").New("request blocked by fraud protection")

// MetricsQuerier is the interface for querying and writing verified-OTP metrics.
type MetricsQuerier interface {
	RecordVerified(ctx context.Context, ip, phoneCountry string) error
	GetVerifiedByCountry24h(ctx context.Context, country string) (int64, error)
	GetVerifiedByCountry1h(ctx context.Context, country string) (int64, error)
	GetVerifiedByIP24h(ctx context.Context, ip string) (int64, error)
	GetVerifiedByCountryPast14DaysRollingMax(ctx context.Context, country string) (int64, error)
}

// LeakyBucketer is the interface for filling and draining the SMS leaky buckets.
type LeakyBucketer interface {
	RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, error)
	RecordSMSOTPVerified(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error
}

type Service struct {
	Metrics       MetricsQuerier
	LeakyBucket   LeakyBucketer
	Config        *config.FraudProtectionConfig
	FeatureConfig *config.FraudProtectionFeatureConfig
	RemoteIP      httputil.RemoteIP
	Clock         clock.Clock
}

// CheckAndRecord is the main entry point called BEFORE sending an SMS.
// It computes thresholds, fills leaky buckets, evaluates warnings, and returns
// ErrBlockedByFraudProtection if action==deny_if_any_warning and warnings were triggered.
func (s *Service) CheckAndRecord(ctx context.Context, phoneNumber, messageType string) error {
	effectiveCfg := effectiveFraudProtectionConfig(s.Config, s.FeatureConfig)

	if !*effectiveCfg.Enabled {
		return nil
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		// If phone number cannot be parsed, skip fraud protection rather than blocking.
		return nil
	}
	phoneCountry := parsedPhone.Alpha2[0]

	if s.isAlwaysAllowed(effectiveCfg, ip, phoneNumber, phoneCountry) {
		return nil
	}

	thresholds, err := s.ComputeThresholds(ctx, ip, phoneCountry)
	if err != nil {
		// Non-fatal: if threshold computation fails, allow the request.
		return nil
	}

	triggered, err := s.LeakyBucket.RecordSMSOTPSent(ctx, ip, phoneCountry, thresholds)
	if err != nil {
		// Non-fatal: if leaky bucket fails, allow the request.
		return nil
	}

	warnings := s.evaluateWarnings(effectiveCfg, triggered)

	action := effectiveCfg.Decision.Action
	if action == config.FraudProtectionDecisionActionDenyIfAnyWarning && len(warnings) > 0 {
		return ErrBlockedByFraudProtection
	}

	return nil
}

// RecordSMSOTPVerified records a verified OTP event.
// Called from otp.Service.VerifyOTP() when code.OOBChannel==SMS (fire-and-forget).
func (s *Service) RecordSMSOTPVerified(ctx context.Context, phoneNumber string) {
	effectiveCfg := effectiveFraudProtectionConfig(s.Config, s.FeatureConfig)
	if !*effectiveCfg.Enabled {
		return
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		return
	}
	phoneCountry := parsedPhone.Alpha2[0]

	// Write to PostgreSQL metrics (fire-and-forget).
	_ = s.Metrics.RecordVerified(ctx, ip, phoneCountry)

	// Drain leaky buckets (internally, count=1).
	s.RevertSMSOTPSent(ctx, phoneNumber, 1)
}

// RevertSMSOTPSent drains all 4 leaky buckets by count units — no PostgreSQL write.
// Used for alt-auth exclusion (unverified OTPs that should not count against limits).
func (s *Service) RevertSMSOTPSent(ctx context.Context, phoneNumber string, count int) {
	effectiveCfg := effectiveFraudProtectionConfig(s.Config, s.FeatureConfig)
	if !*effectiveCfg.Enabled {
		return
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		return
	}
	phoneCountry := parsedPhone.Alpha2[0]

	thresholds, err := s.ComputeThresholds(ctx, ip, phoneCountry)
	if err != nil {
		return
	}

	_ = s.LeakyBucket.RecordSMSOTPVerified(ctx, ip, phoneCountry, thresholds, count)
}

// ComputeThresholds queries MetricsStore for all 4 adaptive thresholds.
func (s *Service) ComputeThresholds(ctx context.Context, ip, phoneCountry string) (LeakyBucketThresholds, error) {
	// Country-based thresholds require three queries.
	rollingMax, err := s.Metrics.GetVerifiedByCountryPast14DaysRollingMax(ctx, phoneCountry)
	if err != nil {
		return LeakyBucketThresholds{}, err
	}

	verifiedByCountry24h, err := s.Metrics.GetVerifiedByCountry24h(ctx, phoneCountry)
	if err != nil {
		return LeakyBucketThresholds{}, err
	}

	verifiedByCountry1h, err := s.Metrics.GetVerifiedByCountry1h(ctx, phoneCountry)
	if err != nil {
		return LeakyBucketThresholds{}, err
	}

	// IP-based threshold requires one query.
	verifiedByIP24h, err := s.Metrics.GetVerifiedByIP24h(ctx, ip)
	if err != nil {
		return LeakyBucketThresholds{}, err
	}

	// Compute daily threshold for country.
	countryDaily := math.Max(20,
		math.Max(
			float64(rollingMax)*0.2,
			float64(verifiedByCountry24h)*0.2,
		),
	)

	// Compute hourly threshold for country.
	countryHourly := math.Max(3,
		math.Max(
			countryDaily/6,
			float64(verifiedByCountry1h)*0.2,
		),
	)

	// Compute daily threshold for IP.
	ipDaily := math.Max(10, float64(verifiedByIP24h)*0.2)

	// Compute hourly threshold for IP.
	ipHourly := math.Max(5, float64(verifiedByIP24h)*0.2/6)

	return LeakyBucketThresholds{
		CountryHourly: countryHourly,
		CountryDaily:  countryDaily,
		IPHourly:      ipHourly,
		IPDaily:       ipDaily,
	}, nil
}

// evaluateWarnings maps LeakyBucketTriggered to []FraudProtectionWarningType,
// filtered to only include warning types that are enabled in config.
func (s *Service) evaluateWarnings(cfg *config.FraudProtectionConfig, triggered LeakyBucketTriggered) []config.FraudProtectionWarningType {
	enabledTypes := make(map[config.FraudProtectionWarningType]bool)
	for _, w := range cfg.Warnings {
		enabledTypes[w.Type] = true
	}

	var warnings []config.FraudProtectionWarningType
	if triggered.IPCountriesDaily && enabledTypes[config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily] {
		warnings = append(warnings, config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily)
	}
	if triggered.CountryDaily && enabledTypes[config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily] {
		warnings = append(warnings, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily)
	}
	if triggered.CountryHourly && enabledTypes[config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly] {
		warnings = append(warnings, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly)
	}
	if triggered.IPDaily && enabledTypes[config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily] {
		warnings = append(warnings, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily)
	}
	if triggered.IPHourly && enabledTypes[config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly] {
		warnings = append(warnings, config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly)
	}
	return warnings
}

// isAlwaysAllowed checks IP CIDRs, IP geo codes, phone geo codes, and phone regex.
func (s *Service) isAlwaysAllowed(cfg *config.FraudProtectionConfig, ip, phoneNumber, phoneCountry string) bool {
	if cfg.Decision == nil || cfg.Decision.AlwaysAllow == nil {
		return false
	}
	alwaysAllow := cfg.Decision.AlwaysAllow

	// Check IP-based rules.
	if alwaysAllow.IPAddress != nil {
		ipAllow := alwaysAllow.IPAddress

		// Check CIDR ranges.
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			for _, cidr := range ipAllow.CIDRs {
				_, ipNet, err := net.ParseCIDR(cidr)
				if err != nil {
					continue
				}
				if ipNet.Contains(parsedIP) {
					return true
				}
			}
		}

		// Check IP geo codes.
		if len(ipAllow.GeoLocationCodes) > 0 {
			if info, ok := geoip.IPString(ip); ok {
				if slices.Contains(ipAllow.GeoLocationCodes, info.CountryCode) {
					return true
				}
			}
		}
	}

	// Check phone number-based rules.
	if alwaysAllow.PhoneNumber != nil {
		phoneAllow := alwaysAllow.PhoneNumber

		// Check phone geo codes.
		if slices.Contains(phoneAllow.GeoLocationCodes, phoneCountry) {
			return true
		}

		// Check phone regex patterns.
		for _, pattern := range phoneAllow.Regex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			if re.MatchString(phoneNumber) {
				return true
			}
		}
	}

	return false
}

// effectiveFraudProtectionConfig returns the effective config, applying the
// feature flag: when IsModifiable is false, the hardcoded default is used.
func effectiveFraudProtectionConfig(
	appCfg *config.FraudProtectionConfig,
	featureCfg *config.FraudProtectionFeatureConfig,
) *config.FraudProtectionConfig {
	if !*featureCfg.IsModifiable {
		c := &config.FraudProtectionConfig{}
		config.SetFieldDefaults(c)
		return c
	}
	return appCfg
}
