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
	return schemaBuilderEmail.Copy()
}
func SchemaBuilderPhoneNumber() validation.SchemaBuilder {
	return schemaBuilderPhoneNumber.Copy()
}
func SchemaBuilderPreferredUsername() validation.SchemaBuilder {
	return schemaBuilderPreferredUsername.Copy()
}
func SchemaBuilderFamilyName() validation.SchemaBuilder {
	return schemaBuilderFamilyName.Copy()
}
func SchemaBuilderGivenName() validation.SchemaBuilder {
	return schemaBuilderGivenName.Copy()
}
func SchemaBuilderMiddleName() validation.SchemaBuilder {
	return schemaBuilderMiddleName.Copy()
}
func SchemaBuilderName() validation.SchemaBuilder {
	return schemaBuilderName.Copy()
}
func SchemaBuilderNickName() validation.SchemaBuilder {
	return schemaBuilderNickName.Copy()
}
func SchemaBuilderPicture() validation.SchemaBuilder {
	return schemaBuilderPicture.Copy()
}
func SchemaBuilderProfile() validation.SchemaBuilder {
	return schemaBuilderProfile.Copy()
}
func SchemaBuilderWebsite() validation.SchemaBuilder {
	return schemaBuilderWebsite.Copy()
}
func SchemaBuilderGender() validation.SchemaBuilder {
	return schemaBuilderGender.Copy()
}
func SchemaBuilderBirthdate() validation.SchemaBuilder {
	return schemaBuilderBirthdate.Copy()
}
func SchemaBuilderZoneinfo() validation.SchemaBuilder {
	return schemaBuilderZoneinfo.Copy()
}
func SchemaBuilderLocale() validation.SchemaBuilder {
	return schemaBuilderLocale.Copy()
}
func SchemaBuilderAddress() validation.SchemaBuilder {
	return schemaBuilderAddress.Copy()
}
func SchemaBuilderAddressFormatted() validation.SchemaBuilder {
	return schemaBuilderAddressFormatted.Copy()
}
func SchemaBuilderAddressStreetAddress() validation.SchemaBuilder {
	return schemaBuilderAddressStreetAddress.Copy()
}
func SchemaBuilderAddressLocality() validation.SchemaBuilder {
	return schemaBuilderAddressLocality.Copy()
}
func SchemaBuilderAddressRegion() validation.SchemaBuilder {
	return schemaBuilderAddressRegion.Copy()
}
func SchemaBuilderAddressPostalCode() validation.SchemaBuilder {
	return schemaBuilderAddressPostalCode.Copy()
}
func SchemaBuilderAddressCountry() validation.SchemaBuilder {
	return schemaBuilderAddressCountry.Copy()
}
