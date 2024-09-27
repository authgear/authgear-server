package userexport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/clock"
	libhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

type UserQueries interface {
	GetPageForExport(page uint64, limit uint64) (users []*user.UserForExport, err error)
	CountAll() (count uint64, err error)
}

type UserExportService struct {
	AppDatabase  *appdb.Handle
	Config       *config.UserProfileConfig
	UserQueries  UserQueries
	Logger       Logger
	HTTPOrigin   libhttputil.HTTPOrigin
	CloudStorage UserExportCloudStorage
	Clock        clock.Clock
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("user-export")}
}

func mapGet[T string | bool | map[string]interface{}](m map[string]interface{}, key string) T {
	value, _ := m[key].(T)
	return value
}

func (s *UserExportService) convertDBUserToRecord(user *user.UserForExport) (record *Record, err error) {
	record = &Record{}
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
			opts := secretcode.URIOptions{
				Issuer:      string(s.HTTPOrigin),
				AccountName: user.EndUserAccountID,
			}
			totp, err := secretcode.NewTOTPFromSecret(authenticator.TOTP.Secret)
			if err != nil {
				return nil, err
			}
			otpauthURI := totp.GetURI(opts).String()

			record.Mfa.TOTPs = append(record.Mfa.TOTPs, &MFATOTP{
				Secret: authenticator.TOTP.Secret,
				URI:    otpauthURI,
			})
		case model.AuthenticatorTypeOOBSMS:
			record.Mfa.PhoneNumbers = append(record.Mfa.PhoneNumbers, authenticator.OOBOTP.Phone)
		}
	}

	return record, nil
}

func (s *UserExportService) ExportRecords(ctx context.Context, request *Request, task *redisqueue.Task) (outputFilename string, err error) {
	resultFile, err := os.CreateTemp("", fmt.Sprintf("export-%s.tmp", task.ID))
	if err != nil {
		return
	}
	defer os.Remove(resultFile.Name())

	if request.Format == "csv" {
		err = s.ExportToCSV(resultFile, request, task)
	} else {
		err = s.ExportToNDJson(resultFile, request, task)
	}
	if err != nil {
		return "", err
	}

	key := url.QueryEscape(fmt.Sprintf("%s-%s-%s.%s", task.AppID, task.ID, s.Clock.NowUTC().Format("20060102150405Z"), request.Format))
	_, err = s.UploadResult(key, resultFile, request.Format)

	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *UserExportService) ExportToNDJson(tmpResult *os.File, request *Request, task *redisqueue.Task) (err error) {
	var offset uint64 = uint64(0)
	for {
		s.Logger.Infof("Export ndjson user page offset %v", offset)
		var page []*user.UserForExport = nil

		err = s.AppDatabase.WithTx(func() (e error) {
			result, pageErr := s.UserQueries.GetPageForExport(offset, BatchSize)
			if pageErr != nil {
				return pageErr
			}
			page = result
			return
		})

		if err != nil {
			return err
		}

		s.Logger.Infof("Found number of users: %v", len(page))

		for _, user := range page {
			var record *Record
			record, err = s.convertDBUserToRecord(user)
			if err != nil {
				return
			}

			var recordJson []byte
			recordJson, err = json.Marshal(record)
			if err != nil {
				return
			}

			_, err = tmpResult.Write(recordJson)
			if err != nil {
				return
			}

			_, err = tmpResult.Write([]byte("\n"))
			if err != nil {
				return
			}
		}

		// Exit export loop early when no more record to read
		if len(page) < BatchSize {
			break
		} else {
			offset = offset + BatchSize
		}
	}

	return nil
}

//nolint:gocognit
func (s *UserExportService) ExportToCSV(tmpResult *os.File, request *Request, task *redisqueue.Task) (err error) {
	csvWriter := csv.NewWriter(tmpResult)

	var exportFields []*FieldPointer
	if request.CSV != nil {
		exportFields = request.CSV.Fields
	}
	// Use default CSV field set if no field specified in request
	if exportFields == nil || len(exportFields) == 0 {
		defaultHeader := CSVField{}
		_ = json.Unmarshal([]byte(DefaultCSVExportField), &defaultHeader)
		exportFields = defaultHeader.Fields

		// Append custom_attributes to default pointer set
		for _, attribute := range s.Config.CustomAttributes.Attributes {
			exportFields = append(exportFields, &FieldPointer{
				Pointer: fmt.Sprintf("/custom_attributes%s", attribute.Pointer),
			})
		}
	}

	headerFields, err := ExtractCSVHeaderField(exportFields)
	if err != nil {
		return err
	}

	err = csvWriter.Write(headerFields)
	if err != nil {
		return err
	}

	var offset uint64 = uint64(0)
	for {
		s.Logger.Infof("Export csv user page offset %v", offset)
		var page []*user.UserForExport = nil

		err = s.AppDatabase.WithTx(func() (e error) {
			result, pageErr := s.UserQueries.GetPageForExport(offset, BatchSize)
			if pageErr != nil {
				return pageErr
			}
			page = result
			return
		})

		if err != nil {
			return err
		}

		s.Logger.Infof("Found number of users: %v", len(page))

		for _, user := range page {
			record, err := s.convertDBUserToRecord(user)
			if err != nil {
				return err
			}

			recordJson, err := json.Marshal(record)
			if err != nil {
				return err
			}

			var recordMap interface{}
			err = json.Unmarshal(recordJson, &recordMap)
			if err != nil {
				return err
			}

			var outputLine = make([]string, 0)
			for _, field := range exportFields {
				value, _ := TraverseRecordValue(recordMap, field.Pointer)
				outputLine = append(outputLine, value)
			}
			err = csvWriter.Write(outputLine)
			if err != nil {
				return err
			}
		}

		// Exit export loop early when no more record to read
		if len(page) < BatchSize {
			break
		} else {
			offset = offset + BatchSize
		}
	}

	csvWriter.Flush()

	return nil
}

func (s *UserExportService) UploadResult(key string, resultFile *os.File, format string) (response *http.Response, err error) {
	file, err := os.Open(resultFile.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", key))
	headers.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	if format == "csv" {
		headers.Set("Content-Type", "text/csv")
	} else {
		headers.Set("Content-Type", "application/x-ndjson")
	}

	presignedRequest, err := s.CloudStorage.PresignPutObject(key, headers)
	if err != nil {
		return
	}

	// From library doc,
	// https://cs.opensource.google/go/go/+/refs/tags/go1.23.1:src/net/http/request.go;l=933
	// file pointer does not set `ContentLength` automatically, so we need to explicit set it
	uploadRequest, err := http.NewRequest(http.MethodPut, presignedRequest.URL.String(), file)
	uploadRequest.ContentLength = fileInfo.Size()
	if err != nil {
		return nil, err
	}

	for key, values := range presignedRequest.Header {
		for _, value := range values {
			uploadRequest.Header.Add(key, value)
		}
	}
	client := &http.Client{}
	response, err = client.Do(uploadRequest)

	if err != nil {
		return nil, err
	}

	return response, nil
}
