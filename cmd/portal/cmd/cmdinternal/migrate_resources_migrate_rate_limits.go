package cmdinternal

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

var cmdInternalMigrateRateLimits = &cobra.Command{
	Use:   "migrate-rate-limits",
	Short: "Migrate rate limits config",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}
		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		internal.MigrateResources(cmd.Context(), &internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateRateLimits,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateRateLimits(appID string, configSourceData map[string]string, DryRun bool) error {
	encodedData := configSourceData["authgear.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	err = MigrateConfigRateLimits(m)
	if err != nil {
		return fmt.Errorf("failed to migrate config: %w", err)
	}

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	_, err = config.Parse(migrated)
	if err != nil {
		return fmt.Errorf("invalid config after migration: %w", err)
	}

	configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

//nolint:gocognit
func MigrateConfigRateLimits(config map[string]any) error {
	// Password failed attempts rate limit
	if m, ok := mapGet[map[string]any](config, "authenticator", "password", "ratelimit", "failed_attempt"); ok {
		passwordBurst := float64(10)
		passwordPeriod := "1m"
		if size, ok := mapGet[float64](m, "size"); ok {
			passwordBurst = size
		}
		if resetPeriod, ok := mapGet[string](m, "reset_period"); ok {
			passwordPeriod = resetPeriod
		}

		mapSetIfNotFound(config, map[string]any{
			"enabled": true,
			"period":  passwordPeriod,
			"burst":   passwordBurst,
		}, "authentication", "rate_limits", "password", "per_user_per_ip")
	}

	// Forgot password code valid period
	forgotPasswordCodeExpirySeconds, ok := mapGet[float64](config, "forgot_password", "reset_code_expiry_seconds")
	if ok {
		codeValidPeriod := time.Second * time.Duration(forgotPasswordCodeExpirySeconds)
		mapSetIfNotFound(config, codeValidPeriod.String(), "forgot_password", "code_valid_period")
	}

	// Verification code valid period
	verificationCodeExpirySeconds, ok := mapGet[float64](config, "verification", "code_expiry_seconds")
	if ok {
		codeValidPeriod := time.Second * time.Duration(verificationCodeExpirySeconds)
		mapSetIfNotFound(config, codeValidPeriod.String(), "verification", "code_valid_period")
		mapSetIfNotFound(config, codeValidPeriod.String(), "authenticator", "oob_otp", "sms", "code_valid_period")
		mapSetIfNotFound(config, codeValidPeriod.String(), "authenticator", "oob_otp", "email", "code_valid_period")
	}

	// OTP failed attempts revocation
	otpFailedAttempts, ok := mapGet[map[string]any](config, "otp", "ratelimit", "failed_attempt")
	if ok {
		if enabled, ok := mapGet[bool](otpFailedAttempts, "enabled"); ok && enabled {
			size, ok := mapGet[float64](otpFailedAttempts, "size")
			if !ok {
				size = 5
			}
			mapSetIfNotFound(config, size, "authentication", "rate_limits", "oob_otp", "sms", "max_failed_attempts_revoke_otp")
			mapSetIfNotFound(config, size, "authentication", "rate_limits", "oob_otp", "email", "max_failed_attempts_revoke_otp")
			mapSetIfNotFound(config, size, "verification", "rate_limits", "sms", "max_failed_attempts_revoke_otp")
			mapSetIfNotFound(config, size, "verification", "rate_limits", "email", "max_failed_attempts_revoke_otp")
		}
	}

	// SMS rate limits
	smsPerPhone, ok := mapGet[map[string]any](config, "messaging", "sms", "ratelimit", "per_phone")
	if ok {
		enabled, found := mapGet[bool](smsPerPhone, "enabled")
		if found {
			size, ok := mapGet[float64](smsPerPhone, "size")
			if !ok {
				size = 10
			}
			resetPeriod, ok := mapGet[string](smsPerPhone, "reset_period")
			if !ok {
				resetPeriod = "24h"
			}

			mapSetIfNotFound(config, map[string]any{
				"enabled": enabled,
				"period":  resetPeriod,
				"burst":   size,
			}, "messaging", "rate_limits", "sms_per_target")
		}
	}
	smsPerIP, ok := mapGet[map[string]any](config, "messaging", "sms", "ratelimit", "per_ip")
	if ok {
		enabled, found := mapGet[bool](smsPerIP, "enabled")
		if found {
			size, ok := mapGet[float64](smsPerIP, "size")
			if !ok {
				size = 120
			}
			resetPeriod, ok := mapGet[string](smsPerIP, "reset_period")
			if !ok {
				resetPeriod = "1m"
			}

			mapSetIfNotFound(config, map[string]any{
				"enabled": enabled,
				"period":  resetPeriod,
				"burst":   size,
			}, "messaging", "rate_limits", "sms_per_ip")
		}
	}

	// SMS resend cooldown
	smsCooldownSeconds, ok := mapGet[float64](config, "messaging", "sms", "ratelimit", "resend_cooldown_seconds")
	if ok {
		triggerCooldown := time.Second * time.Duration(smsCooldownSeconds)
		mapSetIfNotFound(config, triggerCooldown.String(), "verification", "rate_limits", "sms", "trigger_cooldown")
		mapSetIfNotFound(config, triggerCooldown.String(), "forgot_password", "rate_limits", "sms", "trigger_cooldown")
		mapSetIfNotFound(config, triggerCooldown.String(), "authentication", "rate_limits", "oob_otp", "sms", "trigger_cooldown")
	}

	// Email resend cooldown
	emailCooldownSeconds, ok := mapGet[float64](config, "messaging", "email", "ratelimit", "resend_cooldown_seconds")
	if ok {
		triggerCooldown := time.Second * time.Duration(emailCooldownSeconds)
		mapSetIfNotFound(config, triggerCooldown.String(), "verification", "rate_limits", "email", "trigger_cooldown")
		mapSetIfNotFound(config, triggerCooldown.String(), "forgot_password", "rate_limits", "email", "trigger_cooldown")
		mapSetIfNotFound(config, triggerCooldown.String(), "authentication", "rate_limits", "oob_otp", "email", "trigger_cooldown")
	}

	// Welcome message
	mapDelete(config, "welcome_message")

	// Cleanup deprecated config
	mapDelete(config, "otp")
	mapDelete(config, "authenticator", "password", "ratelimit")
	mapDelete(config, "forgot_password", "reset_code_expiry_seconds")
	mapDelete(config, "verification", "code_expiry_seconds")
	mapDelete(config, "messaging", "sms")
	mapDelete(config, "messaging", "email")

	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateRateLimits)
}
