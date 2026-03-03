package fraudprotection

import (
	"context"
	"log/slog"
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
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ServiceLogger = slogutil.NewLogger("fraudprotection")

var ErrBlockedByFraudProtection = apierrors.Forbidden.WithReason("BlockedByFraudProtection").New("request blocked by fraud protection")

const (
	// thresholdScaleFactor is the fraction of historical verified-OTP counts
	// used to set adaptive rate-limit thresholds (20%).
	thresholdScaleFactor = 0.2

	// thresholdHoursPerDay is the denominator for converting a daily threshold
	// to an hourly one.
	thresholdHoursPerDay = 6.0

	// Minimum floor values for each leaky-bucket threshold.
	thresholdMinCountryDaily  = 20.0
	thresholdMinCountryHourly = 3.0
	thresholdMinIPDaily       = 10.0
	thresholdMinIPHourly      = 5.0
)

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
	RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, LeakyBucketLevels, error)
	RecordSMSOTPVerified(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error
}

type Service struct {
	AppID       config.AppID
	Metrics     MetricsQuerier
	LeakyBucket LeakyBucketer
	Config      *config.FraudProtectionConfig
	RemoteIP    httputil.RemoteIP
	Clock       clock.Clock
}

// CheckAndRecord is the main entry point called BEFORE sending an SMS.
// It computes thresholds, fills leaky buckets, evaluates warnings, and returns
// ErrBlockedByFraudProtection if action==deny_if_any_warning and warnings were triggered.
func (s *Service) CheckAndRecord(ctx context.Context, phoneNumber, messageType string) error {
	if !*s.Config.Enabled {
		return nil
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		// If phone number cannot be parsed, skip fraud protection rather than blocking.
		return nil
	}
	phoneCountry := parsedPhone.Alpha2[0]

	if s.isAlwaysAllowed(s.Config, ip, phoneNumber, phoneCountry) {
		return nil
	}

	thresholds, err := s.ComputeThresholds(ctx, ip, phoneCountry)
	if err != nil {
		return err
	}

	triggered, levels, err := s.LeakyBucket.RecordSMSOTPSent(ctx, ip, phoneCountry, thresholds)
	if err != nil {
		return err
	}

	warnings := s.evaluateWarnings(s.Config, triggered)

	if len(warnings) > 0 {
		logger := ServiceLogger.GetLogger(ctx)
		for _, w := range warnings {
			threshold, level := s.warningThresholdAndLevel(w, thresholds, levels)
			logger.Warn(ctx, "fraud protection warning triggered",
				slog.String("app_id", string(s.AppID)),
				slog.String("warning_type", string(w)),
				slog.String("ip", ip),
				slog.String("phone_country", phoneCountry),
				slog.Float64("threshold", threshold),
				slog.Float64("level", level),
			)
		}
	}

	action := s.Config.Decision.Action
	if action == config.FraudProtectionDecisionActionDenyIfAnyWarning && len(warnings) > 0 {
		return ErrBlockedByFraudProtection
	}

	return nil
}

// RecordSMSOTPVerified records a verified OTP event.
// Called from otp.Service.VerifyOTP() when code.OOBChannel==SMS.
func (s *Service) RecordSMSOTPVerified(ctx context.Context, phoneNumber string) error {
	if !*s.Config.Enabled {
		return nil
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		// If phone number cannot be parsed, skip fraud protection rather than failing.
		return nil
	}
	phoneCountry := parsedPhone.Alpha2[0]

	// Write to PostgreSQL metrics.
	if err := s.Metrics.RecordVerified(ctx, ip, phoneCountry); err != nil {
		return err
	}

	// Drain leaky buckets.
	return s.RevertSMSOTPSent(ctx, phoneNumber, 1)
}

// RevertSMSOTPSent drains all 4 leaky buckets by count units — no PostgreSQL write.
// Used for alt-auth exclusion (unverified OTPs that should not count against limits).
func (s *Service) RevertSMSOTPSent(ctx context.Context, phoneNumber string, count int) error {
	if !*s.Config.Enabled {
		return nil
	}

	ip := string(s.RemoteIP)

	parsedPhone, err := phone.ParsePhoneNumberWithUserInput(phoneNumber)
	if err != nil || len(parsedPhone.Alpha2) == 0 {
		// If phone number cannot be parsed, skip fraud protection rather than failing.
		return nil
	}
	phoneCountry := parsedPhone.Alpha2[0]

	thresholds, err := s.ComputeThresholds(ctx, ip, phoneCountry)
	if err != nil {
		return err
	}

	return s.LeakyBucket.RecordSMSOTPVerified(ctx, ip, phoneCountry, thresholds, count)
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
	countryDaily := math.Max(thresholdMinCountryDaily,
		math.Max(
			float64(rollingMax)*thresholdScaleFactor,
			float64(verifiedByCountry24h)*thresholdScaleFactor,
		),
	)

	// Compute hourly threshold for country.
	countryHourly := math.Max(thresholdMinCountryHourly,
		math.Max(
			countryDaily/thresholdHoursPerDay,
			float64(verifiedByCountry1h)*thresholdScaleFactor,
		),
	)

	// Compute daily threshold for IP.
	ipDaily := math.Max(thresholdMinIPDaily, float64(verifiedByIP24h)*thresholdScaleFactor)

	// Compute hourly threshold for IP.
	ipHourly := math.Max(thresholdMinIPHourly, float64(verifiedByIP24h)*thresholdScaleFactor/thresholdHoursPerDay)

	return LeakyBucketThresholds{
		CountryHourly: countryHourly,
		CountryDaily:  countryDaily,
		IPHourly:      ipHourly,
		IPDaily:       ipDaily,
	}, nil
}

// warningThresholdAndLevel returns the threshold and current bucket level for a given warning type.
// For IPCountriesDaily the "level" is the distinct-country count and "threshold" is the fixed constant.
func (s *Service) warningThresholdAndLevel(w config.FraudProtectionWarningType, thresholds LeakyBucketThresholds, levels LeakyBucketLevels) (threshold, level float64) {
	switch w {
	case config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly:
		return thresholds.CountryHourly, levels.CountryHourly
	case config.FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily:
		return thresholds.CountryDaily, levels.CountryDaily
	case config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly:
		return thresholds.IPHourly, levels.IPHourly
	case config.FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily:
		return thresholds.IPDaily, levels.IPDaily
	case config.FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily:
		return ipCountriesThreshold, float64(levels.IPCountriesCount)
	default:
		return 0, 0
	}
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
