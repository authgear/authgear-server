package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	SchemaBuilderAddress = validation.SchemaBuilder{}.
		Type(validation.TypeObject)
	addressProperties := SchemaBuilderAddress.Properties()
	addressProperties.Property("formatted", SchemaBuilderAddressFormatted)
	addressProperties.Property("street_address", SchemaBuilderAddressStreetAddress)
	addressProperties.Property("locality", SchemaBuilderAddressLocality)
	addressProperties.Property("region", SchemaBuilderAddressRegion)
	addressProperties.Property("postal_code", SchemaBuilderAddressPostalCode)
	addressProperties.Property("country", SchemaBuilderAddressCountry)

	SchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		AdditionalPropertiesFalse()

	schemaProperties := SchemaBuilder.Properties()
	schemaProperties.Property("email", SchemaBuilderEmail)
	schemaProperties.Property("phone_number", SchemaBuilderPhoneNumber)
	schemaProperties.Property("preferred_username", SchemaBuilderPreferredUsername)
	schemaProperties.Property("family_name", SchemaBuilderFamilyName)
	schemaProperties.Property("given_name", SchemaBuilderGivenName)
	schemaProperties.Property("middle_name", SchemaBuilderMiddleName)
	schemaProperties.Property("name", SchemaBuilderName)
	schemaProperties.Property("nickname", SchemaBuilderNickName)
	schemaProperties.Property("picture", SchemaBuilderPicture)
	schemaProperties.Property("profile", SchemaBuilderProfile)
	schemaProperties.Property("website", SchemaBuilderWebsite)
	schemaProperties.Property("gender", SchemaBuilderGender)
	schemaProperties.Property("birthdate", SchemaBuilderBirthdate)
	schemaProperties.Property("zoneinfo", SchemaBuilderZoneinfo)
	schemaProperties.Property("locale", SchemaBuilderLocale)
	schemaProperties.Property("address", SchemaBuilderAddress)

	Schema = SchemaBuilder.ToSimpleSchema()
}

var SchemaBuilderAddress validation.SchemaBuilder
var SchemaBuilder validation.SchemaBuilder
var Schema *validation.SimpleSchema

var SchemaBuilderEmail = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("email")

var SchemaBuilderPhoneNumber = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("phone")

var SchemaBuilderPreferredUsername = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderFamilyName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderGivenName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderMiddleName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderNickName = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderPicture = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("x_picture")

var SchemaBuilderProfile = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("uri")

var SchemaBuilderWebsite = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("uri")

var SchemaBuilderGender = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderBirthdate = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("birthdate")

var SchemaBuilderZoneinfo = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("timezone")

var SchemaBuilderLocale = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	Format("bcp47")

var SchemaBuilderAddressFormatted = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderAddressStreetAddress = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderAddressLocality = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderAddressRegion = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderAddressPostalCode = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

var SchemaBuilderAddressCountry = validation.SchemaBuilder{}.
	Type(validation.TypeString).
	MinLength(1)

func Validate(t T) error {
	a := t.ToClaims()
	return Schema.Validator().ValidateValue(a)
}

func SchemaBuilderForPointerString(ptrStr string) (validation.SchemaBuilder, bool) {
	switch ptrStr {
	case "/email":
		return SchemaBuilderEmail, true
	case "/phone_number":
		return SchemaBuilderPhoneNumber, true
	case "/preferred_username":
		return SchemaBuilderPreferredUsername, true
	case "/family_name":
		return SchemaBuilderFamilyName, true
	case "/given_name":
		return SchemaBuilderGivenName, true
	case "/middle_name":
		return SchemaBuilderMiddleName, true
	case "/name":
		return SchemaBuilderName, true
	case "/nickname":
		return SchemaBuilderNickName, true
	case "/picture":
		return SchemaBuilderPicture, true
	case "/profile":
		return SchemaBuilderProfile, true
	case "/website":
		return SchemaBuilderWebsite, true
	case "/gender":
		return SchemaBuilderGender, true
	case "/birthdate":
		return SchemaBuilderBirthdate, true
	case "/zoneinfo":
		return SchemaBuilderZoneinfo, true
	case "/locale":
		return SchemaBuilderLocale, true
	case "/address":
		return SchemaBuilderAddress, true
	case "/address/formatted":
		return SchemaBuilderAddressFormatted, true
	case "/address/street_address":
		return SchemaBuilderAddressStreetAddress, true
	case "/address/locality":
		return SchemaBuilderAddressLocality, true
	case "/address/region":
		return SchemaBuilderAddressRegion, true
	case "/address/postal_code":
		return SchemaBuilderAddressPostalCode, true
	case "/address/country":
		return SchemaBuilderAddressCountry, true
	default:
		return nil, false
	}
}
