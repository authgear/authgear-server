package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStandardAttributesConfig(t *testing.T) {
	Convey("StandardAttributesConfig", t, func() {
		c := &StandardAttributesConfig{}
		c.SetDefaults()

		accessControl := c.GetAccessControl()

		So(accessControl.GetLevel("/name", RoleEndUser, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RoleEndUser, 0), ShouldEqual, AccessControlLevelReadwrite)

		So(accessControl.GetLevel("/name", RoleBearer, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RoleBearer, 0), ShouldEqual, AccessControlLevelReadonly)

		So(accessControl.GetLevel("/name", RolePortalUI, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/family_name", RolePortalUI, 0), ShouldEqual, AccessControlLevelReadwrite)
	})
}

func TestCustomAttributesConfig(t *testing.T) {
	Convey("CustomAttributesConfig", t, func() {
		c := &CustomAttributesConfig{
			Attributes: []*CustomAttributesAttributeConfig{
				&CustomAttributesAttributeConfig{
					Type:    CustomAttributeTypeString,
					Pointer: "/a",
					AccessControl: &UserProfileAttributesAccessControl{
						EndUser:  AccessControlLevelStringHidden,
						Bearer:   AccessControlLevelStringHidden,
						PortalUI: AccessControlLevelStringHidden,
					},
				},
			},
		}

		accessControl := c.GetAccessControl()

		So(accessControl.GetLevel("/a", RoleEndUser, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/a", RoleBearer, 0), ShouldEqual, AccessControlLevelHidden)
		So(accessControl.GetLevel("/a", RolePortalUI, 0), ShouldEqual, AccessControlLevelHidden)
	})
}

func TestCustomAttributesAttributeConfig(t *testing.T) {
	newFloat := func(f float64) *float64 {
		return &f
	}

	newInt64 := func(i int64) int64 {
		return i
	}

	Convey("CustomAttributesAttributeConfig", t, func() {
		test := func(c *CustomAttributesAttributeConfig, schema map[string]interface{}) {
			actual, err := c.ToJSONSchema()
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, schema)
		}

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeString,
		}, map[string]interface{}{
			"type":      "string",
			"minLength": 1,
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeNumber,
		}, map[string]interface{}{
			"type": "number",
		})

		test(&CustomAttributesAttributeConfig{
			Type:    CustomAttributeTypeNumber,
			Minimum: newFloat(0),
		}, map[string]interface{}{
			"type":    "number",
			"minimum": 0.0,
		})

		test(&CustomAttributesAttributeConfig{
			Type:    CustomAttributeTypeNumber,
			Maximum: newFloat(1),
		}, map[string]interface{}{
			"type":    "number",
			"maximum": 1.0,
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeInteger,
		}, map[string]interface{}{
			"type": "integer",
		})

		test(&CustomAttributesAttributeConfig{
			Type:    CustomAttributeTypeInteger,
			Minimum: newFloat(0),
		}, map[string]interface{}{
			"type":    "integer",
			"minimum": newInt64(0),
		})

		test(&CustomAttributesAttributeConfig{
			Type:    CustomAttributeTypeInteger,
			Maximum: newFloat(1),
		}, map[string]interface{}{
			"type":    "integer",
			"maximum": newInt64(1),
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeEnum,
			Enum: []string{"a", "b"},
		}, map[string]interface{}{
			"type": "string",
			"enum": []string{"a", "b"},
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypePhoneNumber,
		}, map[string]interface{}{
			"type":   "string",
			"format": "phone",
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeEmail,
		}, map[string]interface{}{
			"type":   "string",
			"format": "email",
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeURL,
		}, map[string]interface{}{
			"type":   "string",
			"format": "uri",
		})

		test(&CustomAttributesAttributeConfig{
			Type: CustomAttributeTypeCountryCode,
		}, map[string]interface{}{
			"type":   "string",
			"format": "iso3166-1-alpha-2",
		})
	})
}
