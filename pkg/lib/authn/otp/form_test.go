package otp

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/secretcode"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateCode(t *testing.T) {
	testCases := []struct {
		name       string
		form       Form
		cfg        *config.TestModeConfig
		featureCfg *config.TestModeFeatureConfig
		target     string
		userID     string
		expected   string
	}{
		{
			name: "Code - Should prefer config over feature config",
			form: FormCode,
			cfg: &config.TestModeConfig{
				FixedOOBOTP: &config.TestModeOOBOTPConfig{
					Enabled: true,
					Rules: []*config.TestModeOOBOTPRule{
						{
							Regex:     ".*",
							FixedCode: "config_code",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				FixedOOBOTP: &config.TestModeFixedOOBOTPFeatureConfig{
					Enabled: true,
					Code:    "feature_code",
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "config_code",
		},
		{
			name: "Code - Should use feature config if config not enabled",
			form: FormCode,
			cfg: &config.TestModeConfig{
				FixedOOBOTP: &config.TestModeOOBOTPConfig{
					Enabled: false,
					Rules: []*config.TestModeOOBOTPRule{
						{
							Regex:     ".*",
							FixedCode: "config_code",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				FixedOOBOTP: &config.TestModeFixedOOBOTPFeatureConfig{
					Enabled: true,
					Code:    "feature_code",
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "feature_code",
		},
		{
			name: "Code - Should use feature config if config has no fixed code",
			form: FormCode,
			cfg: &config.TestModeConfig{
				FixedOOBOTP: &config.TestModeOOBOTPConfig{
					Enabled: false,
					Rules: []*config.TestModeOOBOTPRule{
						{
							Regex:     ".*",
							FixedCode: "",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				FixedOOBOTP: &config.TestModeFixedOOBOTPFeatureConfig{
					Enabled: true,
					Code:    "feature_code",
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "feature_code",
		},
		{
			name: "Code - Should use feature config if config has no matching rule",
			form: FormCode,
			cfg: &config.TestModeConfig{
				FixedOOBOTP: &config.TestModeOOBOTPConfig{
					Enabled: false,
					Rules: []*config.TestModeOOBOTPRule{
						{
							Regex:     "anothertest@example.com",
							FixedCode: "config_code",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				FixedOOBOTP: &config.TestModeFixedOOBOTPFeatureConfig{
					Enabled: true,
					Code:    "feature_code",
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "feature_code",
		},
		{
			name: "Code - Should generate code if no fixed code",
			form: FormCode,
			cfg: &config.TestModeConfig{
				FixedOOBOTP: &config.TestModeOOBOTPConfig{
					Enabled: false,
					Rules: []*config.TestModeOOBOTPRule{
						{
							Regex:     ".*",
							FixedCode: "",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				FixedOOBOTP: &config.TestModeFixedOOBOTPFeatureConfig{
					Enabled: false,
					Code:    "",
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "[[random_code]]",
		},
		{
			name: "Link - Should respect config",
			form: FormLink,
			cfg: &config.TestModeConfig{
				Email: &config.TestModeEmailConfig{
					Enabled: true,
					Rules: []*config.TestModeEmailRule{
						{
							Regex: ".*",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				DeterministicLinkOTP: &config.TestModeDeterministicLinkOTPFeatureConfig{
					Enabled: true,
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: secretcode.LinkOTPSecretCode.GenerateDeterministic("user1"),
		},
		{
			name: "Link - Should use feature config if config not enabled",
			form: FormLink,
			cfg: &config.TestModeConfig{
				Email: &config.TestModeEmailConfig{
					Enabled: false,
					Rules: []*config.TestModeEmailRule{
						{
							Regex: ".*",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				DeterministicLinkOTP: &config.TestModeDeterministicLinkOTPFeatureConfig{
					Enabled: true,
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: secretcode.LinkOTPSecretCode.GenerateDeterministic("user1"),
		},
		{
			name: "Link - Should use feature config if config has no fixed code",
			form: FormLink,
			cfg: &config.TestModeConfig{
				Email: &config.TestModeEmailConfig{
					Enabled: true,
					Rules: []*config.TestModeEmailRule{
						{
							Regex: ".*",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				DeterministicLinkOTP: &config.TestModeDeterministicLinkOTPFeatureConfig{
					Enabled: true,
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: secretcode.LinkOTPSecretCode.GenerateDeterministic("user1"),
		},
		{
			name: "Link - Should generate code if no deterministic code",
			form: FormLink,
			cfg: &config.TestModeConfig{
				Email: &config.TestModeEmailConfig{
					Enabled: false,
					Rules: []*config.TestModeEmailRule{
						{
							Regex: ".*",
						},
					},
				},
			},
			featureCfg: &config.TestModeFeatureConfig{
				DeterministicLinkOTP: &config.TestModeDeterministicLinkOTPFeatureConfig{
					Enabled: false,
				},
			},
			target:   "test@example.com",
			userID:   "user1",
			expected: "[[random_code]]",
		},
	}

	for _, tc := range testCases {
		Convey(tc.name, t, func() {
			actual := tc.form.GenerateCode(tc.cfg, tc.featureCfg, tc.target, tc.userID)
			if tc.expected == "[[random_code]]" {
				So(len(actual), ShouldEqual, tc.form.CodeLength())
			} else {
				So(actual, ShouldEqual, tc.expected)
			}
		})
	}
}
