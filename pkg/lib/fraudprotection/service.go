package fraudprotection

import (
	"context"
	"log/slog"
	"math"
	"net"
	"regexp"
	"slices"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type EventService interface {
	DispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error
}

type DatabaseHandle interface {
	IsInTx(ctx context.Context) bool
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) error
}

var ServiceLogger = slogutil.NewLogger("fraudprotection")

var ErrBlockedByFraudProtection = apierrors.TooManyRequest.WithReason("BlockedByFraudProtection").New("request blocked by fraud protection")

const (
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
	RecordUnverifiedSMSOTPCountDrained(ctx context.Context, ip, phoneCountry string, count int) error
	GetVerifiedByCountry24h(ctx context.Context, country string) (int64, error)
	GetVerifiedByCountry1h(ctx context.Context, country string) (int64, error)
	GetVerifiedByIP24h(ctx context.Context, ip string) (int64, error)
	GetVerifiedByCountryPast14DaysRollingMax(ctx context.Context, country string) (int64, error)
}

// LeakyBucketer is the interface for filling and draining the SMS leaky buckets.
type LeakyBucketer interface {
	RecordUnverifiedSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, LeakyBucketLevels, error)
	DrainUnverifiedSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error
	RecordSMSOTPVerifiedCountry(ctx context.Context, ip, phoneCountry string) error
}

type Service struct {
	AppID           config.AppID
	Metrics         MetricsQuerier
	LeakyBucket     LeakyBucketer
	Config          *config.FraudProtectionConfig
	RemoteIP        httputil.RemoteIP
	UserAgentString httputil.UserAgentString
	HTTPRequestURL  httputil.HTTPRequestURL
	HTTPReferer     httputil.HTTPReferer
	Clock           clock.Clock
	Database        DatabaseHandle
	EventService    EventService
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

	triggered, levels, err := s.LeakyBucket.RecordUnverifiedSMSOTPSent(ctx, ip, phoneCountry, thresholds)
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
			otelutil.IntCounterAddOne(ctx,
				otelauthgear.CounterFraudProtectionWarningCount,
				otelauthgear.WithFraudProtectionWarningType(string(w)),
			)
		}
	}

	decision := model.FraudProtectionDecisionAllowed
	action := s.Config.Decision.Action
	if action == config.FraudProtectionDecisionActionDenyIfAnyWarning && len(warnings) > 0 {
		decision = model.FraudProtectionDecisionBlocked
	}

	triggeredWarningStrings := make([]string, len(warnings))
	for i, w := range warnings {
		triggeredWarningStrings[i] = string(w)
	}

	var geoCode string
	if info, ok := geoip.IPString(ip); ok {
		geoCode = info.CountryCode
	}

	var userID string
	if uid := session.GetUserID(ctx); uid != nil {
		userID = *uid
	}

	payload := &nonblocking.FraudProtectionDecisionRecordedEventPayload{
		Record: model.FraudProtectionDecisionRecord{
			Timestamp: s.Clock.NowUTC(),
			Decision:  decision,
			Action:    model.FraudProtectionActionSendSMS,
			ActionDetail: model.FraudProtectionDecisionActionDetail{
				Recipient:              phoneNumber,
				Type:                   messageType,
				PhoneNumberCountryCode: phoneCountry,
			},
			TriggeredWarnings: triggeredWarningStrings,
			UserAgent:         string(s.UserAgentString),
			IPAddress:         ip,
			HTTPUrl:           string(s.HTTPRequestURL),
			HTTPReferer:       string(s.HTTPReferer),
			UserID:            userID,
			GeoLocationCode:   geoCode,
		},
	}
	if err := s.dispatchEventImmediately(ctx, payload); err != nil {
		ServiceLogger.GetLogger(ctx).WithError(err).Error(ctx, "failed to dispatch fraud protection decision_recorded event")
	}

	if decision == model.FraudProtectionDecisionBlocked {
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

	if err := s.LeakyBucket.RecordSMSOTPVerifiedCountry(ctx, ip, phoneCountry); err != nil {
		return err
	}

	// Drain leaky buckets.
	return s.RevertSMSOTPSent(ctx, phoneNumber, 1)
}

// RevertSMSOTPSent drains all 4 leaky buckets by count units and records the
// drain count in PostgreSQL audit metrics.
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

	if err := s.LeakyBucket.DrainUnverifiedSMSOTPSent(ctx, ip, phoneCountry, thresholds, count); err != nil {
		return err
	}
	return s.Metrics.RecordUnverifiedSMSOTPCountDrained(ctx, ip, phoneCountry, count)
}

func (s *Service) resolveSMSUnverifiedOTPBudget(phoneCountry string) (globalDaily, globalHourly, effectiveDaily, effectiveHourly float64) {
	budget := s.Config.SMS.UnverifiedOTPBudget
	globalDaily = *budget.DailyRatio
	globalHourly = *budget.HourlyRatio
	effectiveDaily = globalDaily
	effectiveHourly = globalHourly

	for _, override := range budget.ByPhoneCountry {
		if slices.Contains(override.CountryCodes, phoneCountry) {
			if override.DailyRatio != nil {
				effectiveDaily = *override.DailyRatio
			}
			if override.HourlyRatio != nil {
				effectiveHourly = *override.HourlyRatio
			}
			break
		}
	}

	return
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

	globalDailyRatio, globalHourlyRatio, effectiveDailyRatio, effectiveHourlyRatio := s.resolveSMSUnverifiedOTPBudget(phoneCountry)

	// Compute daily threshold for country.
	countryDaily := math.Max(thresholdMinCountryDaily,
		math.Max(
			float64(rollingMax)*effectiveDailyRatio,
			float64(verifiedByCountry24h)*effectiveDailyRatio,
		),
	)

	// Compute hourly threshold for country.
	countryHourly := math.Max(thresholdMinCountryHourly,
		math.Max(
			float64(rollingMax)/thresholdHoursPerDay*effectiveHourlyRatio,
			float64(verifiedByCountry1h)*effectiveHourlyRatio,
		),
	)

	// Compute daily threshold for IP.
	ipDaily := math.Max(thresholdMinIPDaily, float64(verifiedByIP24h)*globalDailyRatio)

	// Compute hourly threshold for IP.
	ipHourly := math.Max(thresholdMinIPHourly, float64(verifiedByIP24h)/thresholdHoursPerDay*globalHourlyRatio)

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
	return isIPAlwaysAllowed(alwaysAllow.IPAddress, ip) ||
		isPhoneAlwaysAllowed(alwaysAllow.PhoneNumber, phoneNumber, phoneCountry)
}

func isIPAlwaysAllowed(ipAllow *config.FraudProtectionIPAlwaysAllow, ip string) bool {
	if ipAllow == nil {
		return false
	}
	if isIPInCIDRs(ip, ipAllow.CIDRs) {
		return true
	}
	if len(ipAllow.GeoLocationCodes) > 0 {
		if info, ok := geoip.IPString(ip); ok {
			return slices.Contains(ipAllow.GeoLocationCodes, info.CountryCode)
		}
	}
	return false
}

func isIPInCIDRs(ip string, cidrs []string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func isPhoneAlwaysAllowed(phoneAllow *config.FraudProtectionPhoneNumberAlwaysAllow, phoneNumber, phoneCountry string) bool {
	if phoneAllow == nil {
		return false
	}
	if slices.Contains(phoneAllow.GeoLocationCodes, phoneCountry) {
		return true
	}
	for _, pattern := range phoneAllow.Regex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(phoneNumber) {
			return true
		}
	}
	return false
}

// dispatchEventImmediately dispatches an event, opening a read-only transaction
// if the caller is not already inside one (same pattern as messaging.Sender).
func (s *Service) dispatchEventImmediately(ctx context.Context, payload event.NonBlockingPayload) error {
	if s.Database.IsInTx(ctx) {
		return s.EventService.DispatchEventImmediately(ctx, payload)
	}
	return s.Database.ReadOnly(ctx, func(ctx context.Context) error {
		return s.EventService.DispatchEventImmediately(ctx, payload)
	})
}
