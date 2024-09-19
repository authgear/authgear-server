package userexport

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserQueries interface {
	GetPageForExport(page uint64, limit uint64) (users []*user.UserForExport, err error)
	CountAll() (count uint64, err error)
}

type UserExportService struct {
	AppDatabase *appdb.Handle
	UserQueries UserQueries
	Logger      Logger
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("user-export")}
}

func mapGet[T string | bool | map[string]interface{}](m map[string]interface{}, key string) T {
	value, _ := m[key].(T)
	return value
}

func convertDBUserToRecord(user *user.UserForExport, record *Record) {
	record.Sub = user.ID

	record.PreferredUsername = mapGet[string](user.StandardAttributes, "preferred_username")
	record.Email = mapGet[string](user.StandardAttributes, "email")
	record.PhoneNumber = mapGet[string](user.StandardAttributes, "phone_number")

	record.EmailVerified = mapGet[bool](user.StandardAttributes, "email_verified")
	record.PhoneNumberVerified = mapGet[bool](user.StandardAttributes, "phone_number_verified")

	record.Name = mapGet[string](user.StandardAttributes, "name")
	record.GivenName = mapGet[string](user.StandardAttributes, "given_name")
	record.FamilyName = mapGet[string](user.StandardAttributes, "family_name")
	record.MiddleName = mapGet[string](user.StandardAttributes, "middle_name")
	record.Nickname = mapGet[string](user.StandardAttributes, "nickname")
	record.Profile = mapGet[string](user.StandardAttributes, "profile")
	record.Picture = mapGet[string](user.StandardAttributes, "picture")
	record.Website = mapGet[string](user.StandardAttributes, "website")
	record.Gender = mapGet[string](user.StandardAttributes, "gender")
	record.Birthdate = mapGet[string](user.StandardAttributes, "birthdate")
	record.Zoneinfo = mapGet[string](user.StandardAttributes, "zoneinfo")
	record.Locale = mapGet[string](user.StandardAttributes, "locale")

	address := mapGet[map[string]interface{}](user.StandardAttributes, "address")
	record.Address = &Address{
		Formatted:     mapGet[string](address, "formatted"),
		StreetAddress: mapGet[string](address, "street_address"),
		Locality:      mapGet[string](address, "locality"),
		Region:        mapGet[string](address, "region"),
		PostalCode:    mapGet[string](address, "postal_code"),
		Country:       mapGet[string](address, "country"),
	}

	record.CustomAttributes = user.CustomAttributes

	record.Roles = user.Roles
	record.Groups = user.Groups

	record.Disabled = user.IsDisabled

	record.Identities = make([]*Identity, 0)
	record.BiometricCount = 0
	record.PasskeyCount = 0
	for _, identity := range user.Identities {
		switch identityType := identity.Type; identityType {
		case model.IdentityTypeBiometric:
			record.BiometricCount = record.BiometricCount + 1
		case model.IdentityTypePasskey:
			record.PasskeyCount = record.PasskeyCount + 1
		case model.IdentityTypeLoginID:
			record.Identities = append(record.Identities, &Identity{
				Type: model.IdentityTypeLoginID,
				LoginID: map[string]interface{}{
					"key":            identity.LoginID.LoginIDKey,
					"type":           identity.LoginID.LoginIDType,
					"value":          identity.LoginID.LoginID,
					"original_value": identity.LoginID.OriginalLoginID,
				},
				Claims: identity.LoginID.Claims,
			})
		case model.IdentityTypeLDAP:
			record.Identities = append(record.Identities, &Identity{
				Type: model.IdentityTypeLDAP,
				LoginID: map[string]interface{}{
					"server_name":             identity.LDAP.ServerName,
					"last_login_username":     identity.LDAP.LastLoginUserName,
					"user_id_attribute_name":  identity.LDAP.UserIDAttributeName,
					"user_id_attribute_value": identity.LDAP.UserIDAttributeValue,
					"attributes":              identity.LDAP.RawEntryJSON,
				},
				Claims: identity.LDAP.Claims,
			})
		case model.IdentityTypeOAuth:
			record.Identities = append(record.Identities, &Identity{
				Type: model.IdentityTypeOAuth,
				LoginID: map[string]interface{}{
					"provider_alias":      identity.OAuth.ProviderAlias,
					"provider_type":       identity.OAuth.ProviderID.Type,
					"provider_subject_id": identity.OAuth.ProviderSubjectID,
					"user_profile":        identity.OAuth.UserProfile,
				},
				Claims: identity.OAuth.Claims,
			})
		}
	}

	record.Mfa = &MFA{
		Emails:       make([]string, 0),
		PhoneNumbers: make([]string, 0),
		TOTPs:        make([]*MFATOTP, 0),
	}
	for _, authenticator := range user.Authenticators {
		switch authenticatorType := authenticator.Type; authenticatorType {
		case model.AuthenticatorTypeOOBEmail:
			record.Mfa.Emails = append(record.Mfa.Emails, authenticator.OOBOTP.Email)
		case model.AuthenticatorTypeTOTP:
			record.Mfa.TOTPs = append(record.Mfa.TOTPs, &MFATOTP{
				Secret: authenticator.TOTP.Secret,
				URI:    authenticator.TOTP.UserID,
			})
		case model.AuthenticatorTypeOOBSMS:
			record.Mfa.PhoneNumbers = append(record.Mfa.PhoneNumbers, authenticator.OOBOTP.Phone)
		}
	}
}

func (s *UserExportService) ExportRecords(ctx context.Context, request *Request) (outputFilename string, err error) {
	// TODO: write to a tmp file
	writer := io.MultiWriter(io.Discard, os.Stdout)

	// Bound export loop maximum count
	const maxPageToGet = 10000
	for pageNumber := 0; pageNumber < maxPageToGet; pageNumber += 1 {
		s.Logger.Infof("Export user page %v", pageNumber)
		var page []*user.UserForExport = nil
		var offset uint64 = uint64(pageNumber * BatchSize)
		err = s.AppDatabase.WithTx(func() (e error) {
			result, pageErr := s.UserQueries.GetPageForExport(offset, BatchSize)
			if pageErr != nil {
				return pageErr
			}
			page = result
			return
		})

		if err != nil {
			return "", err
		}

		s.Logger.Infof("Found number of users: %v", len(page))

		for _, user := range page {
			var record Record
			convertDBUserToRecord(user, &record)

			recordJson, jsonErr := json.Marshal(record)
			if jsonErr != nil {
				return "", jsonErr
			}
			recordBytes := make([]byte, 0)
			recordBytes = append(recordBytes, []byte(recordJson)...)
			recordBytes = append(recordBytes, []byte("\n")...)
			writer.Write(recordBytes)
		}

		// Exit export loop early when no more record to read
		if len(page) < BatchSize {
			break
		}
	}

	// TODO: Upload tmp result output to cloud storage

	// TODO: Return output file name
	return "dummy_output_filename", nil
}
