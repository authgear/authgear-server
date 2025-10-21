package userimport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type identityUpdate struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

type claim struct {
	Name  string
	Value string
}

type UserQueries interface {
	GetRaw(ctx context.Context, userID string) (*user.User, error)
}

type UserCommands interface {
	Create(ctx context.Context, userID string) (*user.User, error)
	UpdateAccountStatus(ctx context.Context, userID string, accountStatus user.AccountStatusWithRefTime) error
}

type IdentityService interface {
	New(ctx context.Context, userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	UpdateWithSpec(ctx context.Context, info *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)

	Create(ctx context.Context, info *identity.Info) error
	Delete(ctx context.Context, info *identity.Info) error
	Update(ctx context.Context, oldInfo *identity.Info, newInfo *identity.Info) error
	CheckDuplicated(ctx context.Context, info *identity.Info) (dup *identity.Info, err error)
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	Create(ctx context.Context, info *authenticator.Info, markVerified bool) error
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Delete(ctx context.Context, info *authenticator.Info) error
}

type VerifiedClaimService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim

	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
	GetClaims(ctx context.Context, userID string) ([]*verification.Claim, error)
	DeleteClaim(ctx context.Context, claim *verification.Claim) error
}

type StandardAttributesService interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesService interface {
	UpdateCustomAttributesWithList(ctx context.Context, role accesscontrol.Role, userID string, l attrs.List) error
}

type RolesGroupsCommands interface {
	ResetUserGroup(ctx context.Context, options *rolesgroups.ResetUserGroupOptions) error
	ResetUserRole(ctx context.Context, options *rolesgroups.ResetUserRoleOptions) error
}

type SearchReindexService interface {
	MarkUsersAsReindexRequiredInTx(ctx context.Context, userIDs []string) error
	EnqueueReindexUserTask(ctx context.Context, userID string) error
}

type UserImportService struct {
	AppDatabase          *appdb.Handle
	LoginIDConfig        *config.LoginIDConfig
	Identities           IdentityService
	Authenticators       AuthenticatorService
	UserCommands         UserCommands
	UserQueries          UserQueries
	VerifiedClaims       VerifiedClaimService
	StandardAttributes   StandardAttributesService
	CustomAttributes     CustomAttributesService
	RolesGroupsCommands  RolesGroupsCommands
	SearchReindexService SearchReindexService
	Clock                clock.Clock
}

var UserImportLogger = slogutil.NewLogger("user-import")

func (s *UserImportService) ImportRecords(ctx context.Context, request *Request) *Result {
	logger := UserImportLogger.GetLogger(ctx)
	total := len(request.Records)
	result := &Result{
		Summary: &Summary{
			Total: total,
		},
	}

	options := &Options{
		Upsert:     request.Upsert,
		Identifier: request.Identifier,
	}

	err := s.AppDatabase.WithPrepareStatementsHandle(ctx, func(ctx context.Context, h db.PreparedStatementsHandle) error {
		for idx, rawMessage := range request.Records {
			s.importRecordInConn(ctx, h, options, idx, rawMessage, total, result)
		}
		return nil
	})
	if err != nil {
		logger.WithError(err).Error(ctx, "encountered unexpected error in using prepared statement handle")
	}

	return result
}

func (s *UserImportService) importRecordInConn(
	ctx context.Context,
	h db.PreparedStatementsHandle,
	options *Options,
	idx int,
	rawMessage json.RawMessage,
	total int,
	result *Result,
) {
	logger := UserImportLogger.GetLogger(ctx)
	detail := Detail{
		Index: idx,
		// Assume the outcome is failed.
		Outcome: OutcomeFailed,
	}

	var record Record
	shouldReindexUser := false
	err := h.WithTx(ctx, func(ctx context.Context) error {
		var err error
		record, err = s.importRecordInTxn(ctx, &detail, options, rawMessage)
		if err != nil {
			return err
		}
		switch detail.Outcome {
		case OutcomeInserted:
			fallthrough
		case OutcomeUpdated:
			shouldReindexUser = true
			err = s.SearchReindexService.MarkUsersAsReindexRequiredInTx(ctx, []string{detail.UserID})
			if err != nil {
				return err
			}
		default:
			// Reindex is not required for other cases
		}
		return nil
	})
	if record != nil {
		record.Redact()
		detail.Record = record
	}

	logger.Info(ctx, "processed record", slog.Int("index", idx+1), slog.Int("total", total), slog.String("outcome", string(detail.Outcome)))

	if err != nil {
		if !apierrors.IsAPIError(err) {
			logger.WithError(err).Error(ctx, err.Error())
		}
		detail.Errors = []*apierrors.APIError{apierrors.AsAPIError(err)}
	}

	result.Details = append(result.Details, detail)

	switch detail.Outcome {
	case OutcomeInserted:
		result.Summary.Inserted += 1
	case OutcomeUpdated:
		result.Summary.Updated += 1
	case OutcomeSkipped:
		result.Summary.Skipped += 1
	case OutcomeFailed:
		result.Summary.Failed += 1
	default:
		result.Summary.Failed += 1
	}

	if shouldReindexUser {
		// Do it after the transaction has committed to ensure the user can be queried
		err = s.SearchReindexService.EnqueueReindexUserTask(ctx, detail.UserID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to enqueue reindex user task")
		}
	}
}

func (s *UserImportService) importRecordInTxn(ctx context.Context, detail *Detail, options *Options, rawMessage json.RawMessage) (record Record, err error) {
	err = options.RecordSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &record)
	if err != nil {
		// If validation failed, still try to unmarshal the input as a record to provide context to the api caller
		_ = json.Unmarshal(rawMessage, &record)
		return
	}

	var infos []*identity.Info
	switch options.Identifier {
	case IdentifierEmail:
		emailPtr, _ := record.Email()
		infos, err = s.Identities.ListByClaim(ctx, IdentifierEmail, *emailPtr)
	case IdentifierPhoneNumber:
		phoneNumberPtr, _ := record.PhoneNumber()
		infos, err = s.Identities.ListByClaim(ctx, IdentifierPhoneNumber, *phoneNumberPtr)
	case IdentifierPreferredUsername:
		preferredUsernamePtr, _ := record.PreferredUsername()
		infos, err = s.Identities.ListByClaim(ctx, IdentifierPreferredUsername, *preferredUsernamePtr)
	default:
		err = fmt.Errorf("unknown identifier: %v", options.Identifier)
	}
	if err != nil {
		return
	}

	switch len(infos) {
	case 0:
		err = s.insertRecordInTxn(ctx, detail, record)
		if err != nil {
			return
		}
		return
	case 1:
		info := infos[0]
		if options.Upsert {
			err = s.upsertRecordInTxn(ctx, detail, options, record, info)
			if err != nil {
				return
			}
			return
		} else {
			detail.UserID = info.UserID
			detail.Outcome = OutcomeSkipped
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "skipping because upsert = false and user exists",
			})
			return
		}
	default:
		err = fmt.Errorf("unexpected number of identities found: %v", len(infos))
		return
	}
}

func (s *UserImportService) checkIdentityDuplicate(ctx context.Context, info *identity.Info) (err error) {
	_, err = s.Identities.CheckDuplicated(ctx, info)
	if err != nil {
		return
	}
	return
}

func (s *UserImportService) insertRecordInTxn(ctx context.Context, detail *Detail, record Record) (err error) {
	userID := uuid.New()
	u, err := s.UserCommands.Create(ctx, userID)
	if err != nil {
		return
	}

	infos, err := s.insertIdentitiesInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertVerifiedClaimsInTxn(ctx, detail, record, userID, infos)
	if err != nil {
		return
	}

	err = s.insertStandardAttributesInTxn(ctx, detail, record, u)
	if err != nil {
		return
	}

	err = s.insertCustomAttributesInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertAccountStatusInTxn(ctx, detail, record, u)
	if err != nil {
		return
	}

	err = s.insertRolesInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertGroupsInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertPasswordInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAPasswordInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAOOBOTPEmailInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAOOBOTPPhoneInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFATOTPInTxn(ctx, detail, record, userID)
	if err != nil {
		return
	}

	detail.UserID = userID
	detail.Outcome = OutcomeInserted
	return
}

func (s *UserImportService) insertIdentitiesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (infos []*identity.Info, err error) {
	var specs []*identity.Spec

	if emailPtr, ok := record.Email(); ok {
		if emailPtr == nil {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "email = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypeEmail)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				detail.Warnings = append(detail.Warnings, Warning{
					Message: "email is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypeEmail,
						Key:   key,
						Value: stringutil.NewUserInputString(*emailPtr),
					},
				})
			}
		}
	}

	if phoneNumberPtr, ok := record.PhoneNumber(); ok {
		if phoneNumberPtr == nil {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "phone_number = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypePhone)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				detail.Warnings = append(detail.Warnings, Warning{
					Message: "phone_number is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypePhone,
						Key:   key,
						Value: stringutil.NewUserInputString(*phoneNumberPtr),
					},
				})
			}

		}
	}

	if preferredUsernamePtr, ok := record.PreferredUsername(); ok {
		if preferredUsernamePtr == nil {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "preferred_username = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypeUsername)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				detail.Warnings = append(detail.Warnings, Warning{
					Message: "preferred_username is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypeUsername,
						Key:   key,
						Value: stringutil.NewUserInputString(*preferredUsernamePtr),
					},
				})
			}
		}
	}

	for _, spec := range specs {
		var info *identity.Info
		info, err = s.Identities.New(ctx, userID, spec, identity.NewIdentityOptions{
			// Allow the developer to bypass blocklist.
			LoginIDEmailByPassBlocklistAllowlist: true,
		})
		if err != nil {
			return
		}
		infos = append(infos, info)
	}

	for _, info := range infos {
		err = s.checkIdentityDuplicate(ctx, info)
		if err != nil {
			return
		}

		err = s.Identities.Create(ctx, info)
		if err != nil {
			return
		}
	}

	return
}

func (s *UserImportService) insertVerifiedClaimsInTxn(ctx context.Context, detail *Detail, record Record, userID string, infos []*identity.Info) (err error) {
	if emailVerified, emailVerifiedOK := record.EmailVerified(); emailVerifiedOK {
		if !emailVerified {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "email_verified = false has no effect in insert.",
			})
		} else {
			var email string
			var emailOK bool
			for _, info := range infos {
				claims := info.AllStandardClaims()
				email, emailOK = claims["email"].(string)
				if emailOK {
					break
				}
			}

			if !emailOK {
				detail.Warnings = append(detail.Warnings, Warning{
					Message: "email_verified = true has no effect when email is absent.",
				})
			} else {
				claim := s.VerifiedClaims.NewVerifiedClaim(ctx, userID, "email", email)
				err = s.VerifiedClaims.MarkClaimVerified(ctx, claim)
				if err != nil {
					return
				}
			}
		}
	}

	if phoneNumberVerified, phoneNumberVerifiedOK := record.PhoneNumberVerified(); phoneNumberVerifiedOK {
		if !phoneNumberVerified {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "phone_number_verified = false has no effect in insert.",
			})
		} else {
			var phoneNumber string
			var phoneNumberOK bool
			for _, info := range infos {
				claims := info.AllStandardClaims()
				phoneNumber, phoneNumberOK = claims["phone_number"].(string)
				if phoneNumberOK {
					break
				}
			}

			if !phoneNumberOK {
				detail.Warnings = append(detail.Warnings, Warning{
					Message: "phone_number_verified = true has no effect when phone_number is absent.",
				})
			} else {
				claim := s.VerifiedClaims.NewVerifiedClaim(ctx, userID, "phone_number", phoneNumber)
				err = s.VerifiedClaims.MarkClaimVerified(ctx, claim)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (s *UserImportService) insertStandardAttributesInTxn(ctx context.Context, detail *Detail, record Record, u *user.User) (err error) {
	stdAttrsList := record.NonIdentityAwareStandardAttributesList()

	stdAttrs, err := stdattrs.T(u.StandardAttributes).MergedWithList(stdAttrsList)
	if err != nil {
		return
	}

	err = s.StandardAttributes.UpdateStandardAttributes(ctx, accesscontrol.RoleGreatest, u.ID, stdAttrs)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertCustomAttributesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	customAttrsList := record.CustomAttributesList()
	err = s.CustomAttributes.UpdateCustomAttributesWithList(ctx, accesscontrol.RoleGreatest, userID, customAttrsList)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertAccountStatusInTxn(ctx context.Context, detail *Detail, record Record, u *user.User) (err error) {
	now := s.Clock.NowUTC()
	var accountStatus *user.AccountStatusWithRefTime
	{
		accountStatusStruct := u.AccountStatus(now)
		accountStatus = &accountStatusStruct
	}

	needUpdate := false

	if disabled, ok := record.Disabled(); ok {
		if !disabled {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "disabled = false has no effect in insert.",
			})
		} else {
			accountStatus, err = accountStatus.DisableIndefinitely(nil)
			if err != nil {
				return
			}
			needUpdate = true
		}
	}

	accountValidFrom, accountValidFromOK := record.AccountValidFrom()
	accountValidUntil, accountValidUntilOK := record.AccountValidUntil()
	switch {
	case accountValidFromOK && accountValidUntilOK:
		accountStatus, err = accountStatus.SetAccountValidPeriod(accountValidFrom, accountValidUntil)
		if err != nil {
			return
		}
		needUpdate = true
	case accountValidFromOK:
		accountStatus, err = accountStatus.SetAccountValidFrom(accountValidFrom)
		if err != nil {
			return
		}
		needUpdate = true
	case accountValidUntilOK:
		accountStatus, err = accountStatus.SetAccountValidUntil(accountValidUntil)
		if err != nil {
			return
		}
		needUpdate = true
	default:
		// Nothing to do.
	}

	if needUpdate {
		err = s.UserCommands.UpdateAccountStatus(ctx, u.ID, *accountStatus)
		if err != nil {
			return
		}
	}

	return
}

func (s *UserImportService) insertRolesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	roleKeys, ok := record.Roles()
	if !ok {
		return
	}

	err = s.RolesGroupsCommands.ResetUserRole(ctx, &rolesgroups.ResetUserRoleOptions{
		UserID:   userID,
		RoleKeys: roleKeys,
	})
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertGroupsInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	groupKeys, ok := record.Groups()
	if !ok {
		return
	}

	err = s.RolesGroupsCommands.ResetUserGroup(ctx, &rolesgroups.ResetUserGroupOptions{
		UserID:    userID,
		GroupKeys: groupKeys,
	})
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertPasswordInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	pw, ok := record.Password()
	if !ok {
		return
	}
	password := Password(pw)
	passwordHash := password.PasswordHash()
	passwordExpireAfter := password.ExpireAfter()

	spec := &authenticator.Spec{
		UserID:    userID,
		Type:      model.AuthenticatorTypePassword,
		IsDefault: false,
		Kind:      authenticator.KindPrimary,
		Password: &authenticator.PasswordSpec{
			PasswordHash: passwordHash,
			ExpireAfter:  passwordExpireAfter,
		},
	}

	info, err := s.Authenticators.New(ctx, spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(ctx, info, false)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAPasswordInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	mfaPasswordObj, ok := mfa.Password()
	if !ok {
		return
	}
	mfaPassword := Password(mfaPasswordObj)

	passwordHash := mfaPassword.PasswordHash()
	passwordExpireAfter := mfaPassword.ExpireAfter()

	spec := &authenticator.Spec{
		UserID:    userID,
		Type:      model.AuthenticatorTypePassword,
		IsDefault: false,
		Kind:      authenticator.KindSecondary,
		Password: &authenticator.PasswordSpec{
			PasswordHash: passwordHash,
			ExpireAfter:  passwordExpireAfter,
		},
	}

	info, err := s.Authenticators.New(ctx, spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(ctx, info, false)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAOOBOTPEmailInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	emailPtr, ok := mfa.Email()
	if !ok {
		return
	}

	if emailPtr == nil {
		detail.Warnings = append(detail.Warnings, Warning{
			Message: "mfa.email = null has no effect in insert.",
		})
		return
	}

	spec := &authenticator.Spec{
		UserID: userID,
		Type:   model.AuthenticatorTypeOOBEmail,
		Kind:   model.AuthenticatorKindSecondary,
		OOBOTP: &authenticator.OOBOTPSpec{
			Email: *emailPtr,
		},
	}

	info, err := s.Authenticators.New(ctx, spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(ctx, info, false)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAOOBOTPPhoneInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	phoneNumberPtr, ok := mfa.PhoneNumber()
	if !ok {
		return
	}

	if phoneNumberPtr == nil {
		detail.Warnings = append(detail.Warnings, Warning{
			Message: "mfa.phone_number = null has no effect in insert.",
		})
		return
	}

	spec := &authenticator.Spec{
		UserID: userID,
		Type:   model.AuthenticatorTypeOOBSMS,
		Kind:   model.AuthenticatorKindSecondary,
		OOBOTP: &authenticator.OOBOTPSpec{
			Phone: *phoneNumberPtr,
		},
	}

	info, err := s.Authenticators.New(ctx, spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(ctx, info, false)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFATOTPInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	totpObj, ok := mfa.TOTP()
	if !ok {
		return
	}

	secret := TOTP(totpObj).Secret()

	spec := &authenticator.Spec{
		UserID: userID,
		Type:   model.AuthenticatorTypeTOTP,
		Kind:   model.AuthenticatorKindSecondary,
		TOTP: &authenticator.TOTPSpec{
			DisplayName: "Imported",
			Secret:      secret,
		},
	}

	info, err := s.Authenticators.New(ctx, spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(ctx, info, false)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertRecordInTxn(ctx context.Context, detail *Detail, options *Options, record Record, info *identity.Info) (err error) {
	err = s.upsertIdentitiesInTxn(ctx, detail, options, record, info)
	if err != nil {
		return
	}

	err = s.upsertVerifiedClaimsInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertStandardAttributesInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertCustomAttributesInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertAccountStatusInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertRolesInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertGroupsInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	// password update behavior is IGNORED.
	// mfa.password update behavior is IGNORED.
	// mfa.totp update behavior is IGNORED.

	err = s.upsertMFAOOBOTPEmailInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertMFAOOBOTPPhoneInTxn(ctx, detail, record, info.UserID)
	if err != nil {
		return
	}

	detail.UserID = info.UserID
	detail.Outcome = OutcomeUpdated
	return
}

func (s *UserImportService) upsertIdentitiesInTxnHelper(ctx context.Context, detail *Detail, userID string, infos []*identity.Info, typ model.LoginIDKeyType, ptr *string) (err error) {
	if ptr == nil {
		err := s.removeIdentityInTxn(ctx, detail, infos, typ)
		if err != nil {
			return err
		}
	} else {
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Type:  typ,
				Key:   string(typ),
				Value: stringutil.NewUserInputString(*ptr),
			},
		}
		err := s.upsertIdentityInTxn(ctx, detail, userID, infos, spec)
		if err != nil {
			return err
		}
	}
	return nil
}

// nolint: gocognit
func (s *UserImportService) upsertIdentitiesInTxn(ctx context.Context, detail *Detail, options *Options, record Record, info *identity.Info) (err error) {
	userID := info.UserID
	infos, err := s.Identities.ListByUser(ctx, userID)
	if err != nil {
		return
	}

	switch options.Identifier {
	case IdentifierEmail:
		if phoneNumberPtr, phoneNumberOK := record.PhoneNumber(); phoneNumberOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypePhone, phoneNumberPtr)
			if err != nil {
				return
			}
		}

		if preferredUsernamePtr, preferredUsernameOK := record.PreferredUsername(); preferredUsernameOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypeUsername, preferredUsernamePtr)
			if err != nil {
				return
			}
		}
	case IdentifierPhoneNumber:
		if emailPtr, emailOK := record.Email(); emailOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypeEmail, emailPtr)
			if err != nil {
				return
			}
		}

		if preferredUsernamePtr, preferredUsernameOK := record.PreferredUsername(); preferredUsernameOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypeUsername, preferredUsernamePtr)
			if err != nil {
				return
			}
		}
	case IdentifierPreferredUsername:
		if emailPtr, emailOK := record.Email(); emailOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypeEmail, emailPtr)
			if err != nil {
				return
			}
		}

		if phoneNumberPtr, phoneNumberOK := record.PhoneNumber(); phoneNumberOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, detail, userID, infos, model.LoginIDKeyTypePhone, phoneNumberPtr)
			if err != nil {
				return
			}
		}
	default:
		err = fmt.Errorf("unknown identifier: %v", options.Identifier)
	}

	return
}

func (s *UserImportService) removeIdentityInTxn(ctx context.Context, detail *Detail, infos []*identity.Info, typ model.LoginIDKeyType) error {
	var toBeRemoved []*identity.Info
	for _, info := range infos {
		info := info
		if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == typ && info.LoginID.LoginIDKey == string(typ) {
			toBeRemoved = append(toBeRemoved, info)
		}
	}

	for _, info := range toBeRemoved {
		err := s.Identities.Delete(ctx, info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserImportService) upsertIdentityInTxn(ctx context.Context, detail *Detail, userID string, infos []*identity.Info, spec *identity.Spec) error {
	var toBeUpdated []identityUpdate
	var toBeInserted []*identity.Info

	isUpdated := false
	for _, info := range infos {
		info := info

		if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == spec.LoginID.Type && info.LoginID.LoginIDKey == spec.LoginID.Key {
			isUpdated = true
			updatedInfo, err := s.Identities.UpdateWithSpec(ctx, info, spec, identity.NewIdentityOptions{
				// Allow the developer to bypass blocklist.
				LoginIDEmailByPassBlocklistAllowlist: true,
			})
			if err != nil {
				return err
			}
			toBeUpdated = append(toBeUpdated, identityUpdate{
				OldInfo: info,
				NewInfo: updatedInfo,
			})
		}
	}
	if !isUpdated {
		info, err := s.Identities.New(ctx, userID, spec, identity.NewIdentityOptions{
			// Allow the developer to bypass blocklist.
			LoginIDEmailByPassBlocklistAllowlist: true,
		})
		if err != nil {
			return err
		}
		toBeInserted = append(toBeInserted, info)
	}

	for _, identityUpdate := range toBeUpdated {
		err := s.checkIdentityDuplicate(ctx, identityUpdate.NewInfo)
		if err != nil {
			return err
		}

		err = s.Identities.Update(ctx, identityUpdate.OldInfo, identityUpdate.NewInfo)
		if err != nil {
			return err
		}
	}

	for _, info := range toBeInserted {
		err := s.checkIdentityDuplicate(ctx, info)
		if err != nil {
			return err
		}

		err = s.Identities.Create(ctx, info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserImportService) setVerifiedInTxn(ctx context.Context, userID string, verifiedClaims []*verification.Claim, c *claim, verified bool) error {
	if verified {
		for _, verifiedClaim := range verifiedClaims {
			// Claim is verified already.
			if verifiedClaim.Name == c.Name && verifiedClaim.Value == c.Value {
				return nil
			}
		}

		verifiedClaim := s.VerifiedClaims.NewVerifiedClaim(ctx, userID, c.Name, c.Value)
		err := s.VerifiedClaims.MarkClaimVerified(ctx, verifiedClaim)
		if err != nil {
			return err
		}
	} else {
		var toBeDeleted *verification.Claim
		for _, verifiedClaim := range verifiedClaims {
			if verifiedClaim.Name == c.Name && verifiedClaim.Value == c.Value {
				toBeDeleted = verifiedClaim
			}
		}
		if toBeDeleted == nil {
			return nil
		}

		err := s.VerifiedClaims.DeleteClaim(ctx, toBeDeleted)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserImportService) upsertEmailVerifiedInTxn(ctx context.Context, detail *Detail, record Record, userID string, infos []*identity.Info, verifiedClaims []*verification.Claim) (err error) {
	if emailVerified, emailVerifiedOK := record.EmailVerified(); emailVerifiedOK {
		emailPtr, emailOK := record.Email()
		if !emailOK {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "email_verified has no effect when email is absent.",
			})
		} else if emailPtr == nil {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "email_verified has no effect when email = null.",
			})
		} else {
			var c *claim
			for _, info := range infos {
				if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == model.LoginIDKeyTypeEmail && info.LoginID.LoginIDKey == string(model.LoginIDKeyTypeEmail) {
					claims := info.AllStandardClaims()
					if email, ok := claims["email"].(string); ok && email == *emailPtr {
						c = &claim{
							Name:  "email",
							Value: email,
						}
					}
				}
			}
			if c != nil {
				err = s.setVerifiedInTxn(ctx, userID, verifiedClaims, c, emailVerified)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (s *UserImportService) upsertPhoneNumberVerifiedInTxn(ctx context.Context, detail *Detail, record Record, userID string, infos []*identity.Info, verifiedClaims []*verification.Claim) (err error) {
	if phoneNumberVerified, phoneNumberVerifiedOK := record.PhoneNumberVerified(); phoneNumberVerifiedOK {
		phoneNumberPtr, phoneNumberOK := record.PhoneNumber()
		if !phoneNumberOK {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "phone_number_verified has no effect when phone_number is absent.",
			})
		} else if phoneNumberPtr == nil {
			detail.Warnings = append(detail.Warnings, Warning{
				Message: "phone_number_verified has no effect when phone_number = null.",
			})
		} else {
			var c *claim
			for _, info := range infos {
				if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == model.LoginIDKeyTypePhone && info.LoginID.LoginIDKey == string(model.LoginIDKeyTypePhone) {
					claims := info.AllStandardClaims()
					if phoneNumber, ok := claims["phone_number"].(string); ok && phoneNumber == *phoneNumberPtr {
						c = &claim{
							Name:  "phone_number",
							Value: phoneNumber,
						}
					}
				}
			}
			if c != nil {
				err = s.setVerifiedInTxn(ctx, userID, verifiedClaims, c, phoneNumberVerified)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (s *UserImportService) upsertVerifiedClaimsInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	infos, err := s.Identities.ListByUser(ctx, userID)
	if err != nil {
		return
	}

	verifiedClaims, err := s.VerifiedClaims.GetClaims(ctx, userID)
	if err != nil {
		return
	}

	err = s.upsertEmailVerifiedInTxn(ctx, detail, record, userID, infos, verifiedClaims)
	if err != nil {
		return
	}

	err = s.upsertPhoneNumberVerifiedInTxn(ctx, detail, record, userID, infos, verifiedClaims)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertStandardAttributesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	u, err := s.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return
	}

	err = s.insertStandardAttributesInTxn(ctx, detail, record, u)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertCustomAttributesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	return s.insertCustomAttributesInTxn(ctx, detail, record, userID)
}

func (s *UserImportService) upsertAccountStatusInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	disabled, disabledOK := record.Disabled()
	accountValidFrom, accountValidFromOK := record.AccountValidFrom()
	accountValidUntil, accountValidUntilOK := record.AccountValidUntil()

	if !disabledOK && !accountValidFromOK && !accountValidUntilOK {
		return
	}

	u, err := s.UserQueries.GetRaw(ctx, userID)
	if err != nil {
		return
	}

	now := s.Clock.NowUTC()
	var accountStatus *user.AccountStatusWithRefTime
	{
		accountStatusStruct := u.AccountStatus(now)
		accountStatus = &accountStatusStruct
	}

	needUpdate := false

	if disabledOK {
		if disabled {
			newAccountStatus, accountStatusErr := accountStatus.DisableIndefinitely(nil)
			if accountStatusErr != nil {
				// Treat invalid account status transition as warning.
				detail.Warnings = append(detail.Warnings, Warning{
					Message: accountStatusErr.Error(),
				})
			} else {
				accountStatus = newAccountStatus
				needUpdate = true
			}
		} else {
			newAccountStatus, accountStatusErr := accountStatus.Reenable()
			if accountStatusErr != nil {
				// Treat invalid account status transition as warning.
				detail.Warnings = append(detail.Warnings, Warning{
					Message: accountStatusErr.Error(),
				})
			} else {
				accountStatus = newAccountStatus
				needUpdate = true
			}
		}
	}

	switch {
	case accountValidFromOK && accountValidUntilOK:
		accountStatus, err = accountStatus.SetAccountValidPeriod(accountValidFrom, accountValidUntil)
		if err != nil {
			return
		}
		needUpdate = true
	case accountValidFromOK:
		accountStatus, err = accountStatus.SetAccountValidFrom(accountValidFrom)
		if err != nil {
			return
		}
		needUpdate = true
	case accountValidUntilOK:
		accountStatus, err = accountStatus.SetAccountValidUntil(accountValidUntil)
		if err != nil {
			return
		}
		needUpdate = true
	default:
		// Nothing to do.
	}

	if needUpdate {
		err = s.UserCommands.UpdateAccountStatus(ctx, userID, *accountStatus)
		if err != nil {
			return
		}
	}

	return
}

func (s *UserImportService) upsertRolesInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	return s.insertRolesInTxn(ctx, detail, record, userID)
}

func (s *UserImportService) upsertGroupsInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	return s.insertGroupsInTxn(ctx, detail, record, userID)
}

func (s *UserImportService) upsertMFAOOBOTPEmailInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	emailPtr, ok := mfa.Email()
	if !ok {
		return
	}

	infos, err := s.Authenticators.List(
		ctx,
		userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(model.AuthenticatorTypeOOBEmail),
	)
	if err != nil {
		return
	}

	if emailPtr == nil {
		for _, info := range infos {
			err = s.Authenticators.Delete(ctx, info)
			if err != nil {
				return
			}
		}
	} else {
		spec := &authenticator.Spec{
			UserID: userID,
			Type:   model.AuthenticatorTypeOOBEmail,
			Kind:   model.AuthenticatorKindSecondary,
			OOBOTP: &authenticator.OOBOTPSpec{
				Email: *emailPtr,
			},
		}

		var expected *authenticator.Info
		expected, err = s.Authenticators.New(ctx, spec)
		if err != nil {
			return
		}

		var found *authenticator.Info
		for _, info := range infos {
			if info.Equal(expected) {
				found = info
			}
		}

		// Not found. We delete all and create again.
		if found == nil {
			for _, info := range infos {
				err = s.Authenticators.Delete(ctx, info)
				if err != nil {
					return
				}
			}

			err = s.Authenticators.Create(ctx, expected, false)
			if err != nil {
				return
			}
		}

		// Otherwise it is found. Nothing to do.
	}

	return
}

func (s *UserImportService) upsertMFAOOBOTPPhoneInTxn(ctx context.Context, detail *Detail, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	phoneNumberPtr, ok := mfa.PhoneNumber()
	if !ok {
		return
	}

	infos, err := s.Authenticators.List(
		ctx,
		userID,
		authenticator.KeepKind(authenticator.KindSecondary),
		authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
	)
	if err != nil {
		return
	}

	if phoneNumberPtr == nil {
		for _, info := range infos {
			err = s.Authenticators.Delete(ctx, info)
			if err != nil {
				return
			}
		}
	} else {
		spec := &authenticator.Spec{
			UserID: userID,
			Type:   model.AuthenticatorTypeOOBSMS,
			Kind:   model.AuthenticatorKindSecondary,
			OOBOTP: &authenticator.OOBOTPSpec{
				Phone: *phoneNumberPtr,
			},
		}

		var expected *authenticator.Info
		expected, err = s.Authenticators.New(ctx, spec)
		if err != nil {
			return
		}

		var found *authenticator.Info
		for _, info := range infos {
			if info.Equal(expected) {
				found = info
			}
		}

		// Not found. We delete all and create again.
		if found == nil {
			for _, info := range infos {
				err = s.Authenticators.Delete(ctx, info)
				if err != nil {
					return
				}
			}

			err = s.Authenticators.Create(ctx, expected, false)
			if err != nil {
				return
			}
		}

		// Otherwise it is found. Nothing to do.
	}

	return
}
