package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// MockSecretItemData is a mock implementation of SecretItemData for testing
type MockSecretItemData struct {
	sensitiveStrings []string
}

func (m *MockSecretItemData) SensitiveStrings() []string {
	return m.sensitiveStrings
}

func TestNewMaskPatternFromSecretConfig(t *testing.T) {
	Convey("NewMaskPatternFromSecretConfig", t, func() {
		Convey("should return empty slice for empty config", func() {
			cfg := &SecretConfig{
				Secrets: []SecretItem{},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(patterns, ShouldBeEmpty)
		})

		Convey("should return empty slice for nil config", func() {
			patterns := NewMaskPatternFromSecretConfig(nil)
			So(patterns, ShouldBeEmpty)
		})

		Convey("should create mask patterns from sensitive strings", func() {
			mockData := &MockSecretItemData{
				sensitiveStrings: []string{"secret1", "secret2", "secret3"},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 3)
		})

		Convey("should skip empty sensitive strings", func() {
			mockData := &MockSecretItemData{
				sensitiveStrings: []string{"secret1", "", "secret2", ""},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 2)
		})

		Convey("should handle multiple secret items", func() {
			mockData1 := &MockSecretItemData{
				sensitiveStrings: []string{"secret1", "secret2"},
			}
			mockData2 := &MockSecretItemData{
				sensitiveStrings: []string{"secret3", "secret4"},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData1},
					{Data: mockData2},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 4)
		})

		Convey("should handle mix of empty and non-empty sensitive strings", func() {
			mockData1 := &MockSecretItemData{
				sensitiveStrings: []string{"secret1", ""},
			}
			mockData2 := &MockSecretItemData{
				sensitiveStrings: []string{"", "secret2"},
			}
			mockData3 := &MockSecretItemData{
				sensitiveStrings: []string{},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData1},
					{Data: mockData2},
					{Data: mockData3},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 2)
		})

		Convey("should create PlainMaskPattern instances", func() {
			mockData := &MockSecretItemData{
				sensitiveStrings: []string{"test-secret"},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 1)

			// Test that the pattern works correctly
			result := patterns[0].Mask("password=test-secret", "***")
			So(result, ShouldEqual, "password=***")
		})

		Convey("should handle special characters in sensitive strings", func() {
			mockData := &MockSecretItemData{
				sensitiveStrings: []string{"$pecial@char#", "regex[pattern]"},
			}
			cfg := &SecretConfig{
				Secrets: []SecretItem{
					{Data: mockData},
				},
			}
			patterns := NewMaskPatternFromSecretConfig(cfg)
			So(len(patterns), ShouldEqual, 2)

			// Test that special characters are handled as plain text
			result1 := patterns[0].Mask("key=$pecial@char#", "***")
			So(result1, ShouldEqual, "key=***")

			result2 := patterns[1].Mask("pattern=regex[pattern]", "***")
			So(result2, ShouldEqual, "pattern=***")
		})
	})
}
