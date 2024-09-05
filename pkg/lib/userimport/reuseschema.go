package userimport

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ReusedSchemaBuilders struct {
	Email             validation.SchemaBuilder
	PhoneNumber       validation.SchemaBuilder
	PreferredUsername validation.SchemaBuilder
	FamilyName        validation.SchemaBuilder
	GivenName         validation.SchemaBuilder
	MiddleName        validation.SchemaBuilder
	Name              validation.SchemaBuilder
	Nickname          validation.SchemaBuilder
	Picture           validation.SchemaBuilder
	Profile           validation.SchemaBuilder
	Website           validation.SchemaBuilder
	Gender            validation.SchemaBuilder
	Birthdate         validation.SchemaBuilder
	Zoneinfo          validation.SchemaBuilder
	Locale            validation.SchemaBuilder
	Address           validation.SchemaBuilder
}

func reuseSchemaBuilders() ReusedSchemaBuilders {
	return ReusedSchemaBuilders{
		Email:             stdattrs.SchemaBuilderEmail(),
		PhoneNumber:       stdattrs.SchemaBuilderPhoneNumber(),
		PreferredUsername: stdattrs.SchemaBuilderPreferredUsername(),
		FamilyName:        stdattrs.SchemaBuilderFamilyName(),
		GivenName:         stdattrs.SchemaBuilderGivenName(),
		MiddleName:        stdattrs.SchemaBuilderMiddleName(),
		Name:              stdattrs.SchemaBuilderName(),
		Nickname:          stdattrs.SchemaBuilderNickName(),
		Picture:           stdattrs.SchemaBuilderPicture(),
		Profile:           stdattrs.SchemaBuilderProfile(),
		Website:           stdattrs.SchemaBuilderWebsite(),
		Gender:            stdattrs.SchemaBuilderGender(),
		Birthdate:         stdattrs.SchemaBuilderBirthdate(),
		Zoneinfo:          stdattrs.SchemaBuilderZoneinfo(),
		Locale:            stdattrs.SchemaBuilderLocale(),
		Address:           stdattrs.SchemaBuilderAddress(),
	}
}
