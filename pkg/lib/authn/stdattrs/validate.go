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
		return schemaBuilderEmail, true
	case "/phone_number":
		return schemaBuilderPhoneNumber, true
	case "/preferred_username":
		return schemaBuilderPreferredUsername, true
	case "/family_name":
		return schemaBuilderFamilyName, true
	case "/given_name":
		return schemaBuilderGivenName, true
	case "/middle_name":
		return schemaBuilderMiddleName, true
	case "/name":
		return schemaBuilderName, true
	case "/nickname":
		return schemaBuilderNickName, true
	case "/picture":
		return schemaBuilderPicture, true
	case "/profile":
		return schemaBuilderProfile, true
	case "/website":
		return schemaBuilderWebsite, true
	case "/gender":
		return schemaBuilderGender, true
	case "/birthdate":
		return schemaBuilderBirthdate, true
	case "/zoneinfo":
		return schemaBuilderZoneinfo, true
	case "/locale":
		return schemaBuilderLocale, true
	case "/address":
		return schemaBuilderAddress, true
	case "/address/formatted":
		return schemaBuilderAddressFormatted, true
	case "/address/street_address":
		return schemaBuilderAddressStreetAddress, true
	case "/address/locality":
		return schemaBuilderAddressLocality, true
	case "/address/region":
		return schemaBuilderAddressRegion, true
	case "/address/postal_code":
		return schemaBuilderAddressPostalCode, true
	case "/address/country":
		return schemaBuilderAddressCountry, true
	default:
		return nil, false
	}
}
