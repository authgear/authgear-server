package userexport

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type UserQueries interface {
	GetPageForExport(ctx context.Context, page uint64, limit uint64) (users []*user.UserForExport, err error)
	CountAll(ctx context.Context) (count uint64, err error)
}

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return HTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

type UserExportService struct {
	AppDatabase  *appdb.Handle
	Config       *config.UserProfileConfig
	UserQueries  UserQueries
	HTTPOrigin   httputil.HTTPOrigin
	HTTPClient   HTTPClient
	CloudStorage UserExportCloudStorage
	Clock        clock.Clock
}

var UserExportLogger = slogutil.NewLogger("user-export")

func mapGet[T string | bool | map[string]interface{}](m map[string]interface{}, key string) T {
	value, _ := m[key].(T)
	return value
}

func (s *UserExportService) convertDBUserToRecord(user *user.UserForExport) (record *Record, err error) {
	record = &Record{}
	record.Sub = user.ID
	record.CreatedAt = user.CreatedAt

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
	record.DeleteAt = user.DeleteAt
	record.LastLoginAt = user.LastLoginAt
	record.TemporarilyDisabledFrom = user.TemporarilyDisabledFrom
	record.TemporarilyDisabledUntil = user.TemporarilyDisabledUntil
	record.AccountValidFrom = user.AccountValidFrom
	record.AccountValidUntil = user.AccountValidUntil
	record.IsAnonymized = user.IsAnonymized
	record.AnonymizeAt = user.AnonymizeAt

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
				LoginID: &IdentityLoginID{
					Key:           identity.LoginID.LoginIDKey,
					Type:          string(identity.LoginID.LoginIDType),
					Value:         identity.LoginID.LoginID,
					OriginalValue: identity.LoginID.OriginalLoginID,
				},
				Claims: identity.LoginID.Claims,
			})
		case model.IdentityTypeLDAP:
			lastLoginUsername := ""
			if identity.LDAP.LastLoginUserName != nil {
				lastLoginUsername = *identity.LDAP.LastLoginUserName
			}
			record.Identities = append(record.Identities, &Identity{
				Type: model.IdentityTypeLDAP,
				LDAP: &IdentityLDAP{
					ServerName:           identity.LDAP.ServerName,
					LastLoginUsername:    lastLoginUsername,
					UserIDAttributeName:  identity.LDAP.UserIDAttributeName,
					UserIDAttributeValue: identity.LDAP.UserIDAttributeValueDisplayValue(),
					Attributes:           identity.LDAP.RawEntryJSON,
				},
				Claims: identity.LDAP.Claims,
			})
		case model.IdentityTypeOAuth:
			record.Identities = append(record.Identities, &Identity{
				Type: model.IdentityTypeOAuth,
				OAuth: &IdentityOAuth{
					ProviderAlias:     identity.OAuth.ProviderAlias,
					ProviderType:      identity.OAuth.ProviderID.Type,
					ProviderSubjectID: identity.OAuth.ProviderSubjectID,
					UserProfile:       identity.OAuth.UserProfile,
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
	logger := UserExportLogger.GetLogger(ctx)

	defer func() {
		if err != nil {
			logger.WithError(err).Error(ctx, "export failed")
		}
	}()

	resultFile, err := os.CreateTemp("", fmt.Sprintf("export-%s.tmp", task.ID))
	if err != nil {
		return
	}
	defer os.Remove(resultFile.Name())

	if request.Format == "csv" {
		err = s.ExportToCSV(ctx, resultFile, request, task)
	} else {
		err = s.ExportToNDJson(ctx, resultFile, request, task)
	}
	if err != nil {
		return "", err
	}

	key := url.QueryEscape(fmt.Sprintf("%s-%s-%s.%s", task.AppID, task.ID, s.Clock.NowUTC().Format("20060102150405Z"), request.Format))
	_, err = s.UploadResult(ctx, key, resultFile, request.Format)

	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *UserExportService) ExportToNDJson(ctx context.Context, tmpResult *os.File, request *Request, task *redisqueue.Task) (err error) {
	logger := UserExportLogger.GetLogger(ctx)
	var offset uint64 = uint64(0)
	for {
		logger.Info(ctx, "Export ndjson user page offset", slog.Uint64("offset", offset))
		var page []*user.UserForExport = nil

		err = s.AppDatabase.WithTx(ctx, func(ctx context.Context) (e error) {
			result, pageErr := s.UserQueries.GetPageForExport(ctx, offset, BatchSize)
			if pageErr != nil {
				return pageErr
			}
			page = result
			return
		})

		if err != nil {
			return err
		}

		logger.Info(ctx, "Found number of users", slog.Int("count", len(page)))

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
func (s *UserExportService) ExportToCSV(ctx context.Context, tmpResult *os.File, request *Request, task *redisqueue.Task) (err error) {
	logger := UserExportLogger.GetLogger(ctx)
	csvWriter := csv.NewWriter(tmpResult)

	var exportFields []*FieldPointer
	if request.CSV != nil {
		exportFields = request.CSV.Fields
	}
	// Use default CSV field set if no field specified in request
	if exportFields == nil || len(exportFields) == 0 {
		exportFields = defaultCSVExportFields

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
		logger.Info(ctx, "Export csv user page offset", slog.Uint64("offset", offset))
		var page []*user.UserForExport = nil

		err = s.AppDatabase.WithTx(ctx, func(ctx context.Context) (e error) {
			result, pageErr := s.UserQueries.GetPageForExport(ctx, offset, BatchSize)
			if pageErr != nil {
				return pageErr
			}
			page = result
			return
		})

		if err != nil {
			return err
		}

		logger.Info(ctx, "Found number of users", slog.Int("count", len(page)))

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

func (s *UserExportService) UploadResult(ctx context.Context, key string, resultFile *os.File, format string) (response *http.Response, err error) {
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

	presignedRequest, err := s.CloudStorage.PresignPutObject(ctx, key, headers)
	if err != nil {
		return
	}

	// From library doc,
	// https://cs.opensource.google/go/go/+/refs/tags/go1.23.1:src/net/http/request.go;l=933
	// file pointer does not set `ContentLength` automatically, so we need to explicit set it
	uploadRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedRequest.URL.String(), file)
	uploadRequest.ContentLength = fileInfo.Size()
	if err != nil {
		return nil, err
	}

	for key, values := range presignedRequest.Header {
		for _, value := range values {
			uploadRequest.Header.Add(key, value)
		}
	}
	response, err = s.HTTPClient.Do(uploadRequest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserExportService) makeRequestSchema() validation.SchemaBuilder {
	// Currently we only check the pointer is a valid JSON pointer.
	// But we do not check whether the pointer can possibly point to something.
	// For example, /nonsense points to nothing, and /identities/0/type can possibly points to something.
	// But the set of valid pointers are infinite, so we do not attempt to valid them here.
	pointer := validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Format("json-pointer")

	field := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("pointer")
	field.Properties().
		Property(
			"field_name",
			validation.SchemaBuilder{}.
				Type(validation.TypeString).
				MinLength(1),
		).
		Property("pointer", pointer)

	csv := validation.SchemaBuilder{}.
		Type(validation.TypeObject)
	csv.Properties().
		Property(
			"fields",
			validation.SchemaBuilder{}.
				Type(validation.TypeArray).
				MinItems(1).
				Items(field),
		)

	root := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		AdditionalPropertiesFalse().
		Required("format")

	root.Properties().
		Property(
			"format",
			validation.SchemaBuilder{}.
				Type(validation.TypeString).
				Enum("ndjson", "csv"),
		).
		Property("csv", csv)

	return root
}

func (s *UserExportService) ParseExportRequest(w http.ResponseWriter, r *http.Request) (*Request, error) {
	var request Request
	schema := s.makeRequestSchema().ToSimpleSchema()
	err := httputil.BindJSONBody(r, w, schema.Validator(), &request)
	if err != nil {
		return nil, err
	}

	var fields []*FieldPointer
	if request.CSV != nil {
		fields = request.CSV.Fields
	}
	_, err = ExtractCSVHeaderField(fields)
	if err != nil {
		return nil, err
	}

	return &request, nil
}
