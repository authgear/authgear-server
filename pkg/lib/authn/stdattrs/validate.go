package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// SchemaBuilders in this file should be private to avoid unexpected mutation,
// due to the fact that SchemaBuilder is just a map
// We only expose public functions to create SchemaBuilder
func init() {
	schemaBuilderAddress = validation.SchemaBuilder{}.
		Type(validation.TypeObject)
	addressProperties := schemaBuilderAddress.Properties()
	addressProperties.Property("formatted", schemaBuilderAddressFormatted)
	addressProperties.Property("street_address", schemaBuilderAddressStreetAddress)
	addressProperties.Property("locality", schemaBuilderAddressLocality)
	addressProperties.Property("region", schemaBuilderAddressRegion)
	addressProperties.Property("postal_code", schemaBuilderAddressPostalCode)
	addressProperties.Property("country", schemaBuilderAddressCountry)

	schemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		AdditionalPropertiesFalse()

	schemaProperties := schemaBuilder.Properties()
	schemaProperties.Property("email", schemaBuilderEmail)
	schemaProperties.Property("phone_number", schemaBuilderPhoneNumber)
	schemaProperties.Property("preferred_username", schemaBuilderPreferredUsername)
	schemaProperties.Property("family_name", schemaBuilderFamilyName)
	schemaProperties.Property("given_name", schemaBuilderGivenName)
	schemaProperties.Property("middle_name", schemaBuilderMiddleName)
	schemaProperties.Property("name", schemaBuilderName)
	schemaProperties.Property("nickname", schemaBuilderNickName)
	schemaProperties.Property("picture", schemaBuilderPicture)
	schemaProperties.Property("profile", schemaBuilderProfile)
	schemaProperties.Property("website", schemaBuilderWebsite)
	schemaProperties.Property("gender", schemaBuilderGender)
	schemaProperties.Property("birthdate", schemaBuilderBirthdate)
	schemaProperties.Property("zoneinfo", schemaBuilderZoneinfo)
	schemaProperties.Property("locale", schemaBuilderLocale)
	schemaProperties.Property("address", schemaBuilderAddress)

	schema = schemaBuilder.ToSimpleSchema()
}

var schemaBuilderAddress validation.SchemaBuilder
var schemaBuilder validation.SchemaBuilder
var schema *validation.SimpleSchema

var schemaBuilderEmail = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("email")

var schemaBuilderPhoneNumber = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("phone")

var schemaBuilderPreferredUsername = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderFamilyName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderGivenName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderMiddleName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderNickName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderPicture = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("x_picture")

var schemaBuilderProfile = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("uri")

var schemaBuilderWebsite = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("uri")

var schemaBuilderGender = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderBirthdate = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("birthdate")

var schemaBuilderZoneinfo = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("timezone")

var schemaBuilderLocale = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("bcp47")

var schemaBuilderAddressFormatted = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderAddressStreetAddress = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderAddressLocality = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderAddressRegion = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderAddressPostalCode = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var schemaBuilderAddressCountry = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

func Validate(t T) error {
	a := t.ToClaims()
	return schema.Validator().ValidateValue(a)
}

func SchemaBuilderForPointerString(ptrStr string) (validation.SchemaBuilder, bool) {
	switch ptrStr {
	case "/email":
		return SchemaBuilderEmail(), true
	case "/phone_number":
		return SchemaBuilderPhoneNumber(), true
	case "/preferred_username":
		return SchemaBuilderPreferredUsername(), true
	case "/family_name":
		return SchemaBuilderFamilyName(), true
	case "/given_name":
		return SchemaBuilderGivenName(), true
	case "/middle_name":
		return SchemaBuilderMiddleName(), true
	case "/name":
		return SchemaBuilderName(), true
	case "/nickname":
		return SchemaBuilderNickName(), true
	case "/picture":
		return SchemaBuilderPicture(), true
	case "/profile":
		return SchemaBuilderProfile(), true
	case "/website":
		return SchemaBuilderWebsite(), true
	case "/gender":
		return SchemaBuilderGender(), true
	case "/birthdate":
		return SchemaBuilderBirthdate(), true
	case "/zoneinfo":
		return SchemaBuilderZoneinfo(), true
	case "/locale":
		return SchemaBuilderLocale(), true
	case "/address":
		return SchemaBuilderAddress(), true
	case "/address/formatted":
		return SchemaBuilderAddressFormatted(), true
	case "/address/street_address":
		return SchemaBuilderAddressStreetAddress(), true
	case "/address/locality":
		return SchemaBuilderAddressLocality(), true
	case "/address/region":
		return SchemaBuilderAddressRegion(), true
	case "/address/postal_code":
		return SchemaBuilderAddressPostalCode(), true
	case "/address/country":
		return SchemaBuilderAddressCountry(), true
	default:
		return nil, false
	}
}

func SchemaBuilderEmail() validation.SchemaBuilder {
	return schemaBuilderEmail.Clone()
}
func SchemaBuilderPhoneNumber() validation.SchemaBuilder {
	return schemaBuilderPhoneNumber.Clone()
}
func SchemaBuilderPreferredUsername() validation.SchemaBuilder {
	return schemaBuilderPreferredUsername.Clone()
}
func SchemaBuilderFamilyName() validation.SchemaBuilder {
	return schemaBuilderFamilyName.Clone()
}
func SchemaBuilderGivenName() validation.SchemaBuilder {
	return schemaBuilderGivenName.Clone()
}
func SchemaBuilderMiddleName() validation.SchemaBuilder {
	return schemaBuilderMiddleName.Clone()
}
func SchemaBuilderName() validation.SchemaBuilder {
	return schemaBuilderName.Clone()
}
func SchemaBuilderNickName() validation.SchemaBuilder {
	return schemaBuilderNickName.Clone()
}
func SchemaBuilderPicture() validation.SchemaBuilder {
	return schemaBuilderPicture.Clone()
}
func SchemaBuilderProfile() validation.SchemaBuilder {
	return schemaBuilderProfile.Clone()
}
func SchemaBuilderWebsite() validation.SchemaBuilder {
	return schemaBuilderWebsite.Clone()
}
func SchemaBuilderGender() validation.SchemaBuilder {
	return schemaBuilderGender.Clone()
}
func SchemaBuilderBirthdate() validation.SchemaBuilder {
	return schemaBuilderBirthdate.Clone()
}
func SchemaBuilderZoneinfo() validation.SchemaBuilder {
	return schemaBuilderZoneinfo.Clone()
}
func SchemaBuilderLocale() validation.SchemaBuilder {
	return schemaBuilderLocale.Clone()
}
func SchemaBuilderAddress() validation.SchemaBuilder {
	return schemaBuilderAddress.Clone()
}
func SchemaBuilderAddressFormatted() validation.SchemaBuilder {
	return schemaBuilderAddressFormatted.Clone()
}
func SchemaBuilderAddressStreetAddress() validation.SchemaBuilder {
	return schemaBuilderAddressStreetAddress.Clone()
}
func SchemaBuilderAddressLocality() validation.SchemaBuilder {
	return schemaBuilderAddressLocality.Clone()
}
func SchemaBuilderAddressRegion() validation.SchemaBuilder {
	return schemaBuilderAddressRegion.Clone()
}
func SchemaBuilderAddressPostalCode() validation.SchemaBuilder {
	return schemaBuilderAddressPostalCode.Clone()
}
func SchemaBuilderAddressCountry() validation.SchemaBuilder {
	return schemaBuilderAddressCountry.Clone()
}
