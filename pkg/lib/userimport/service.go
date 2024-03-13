package userimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type importResult struct {
	Outcome  Outcome
	Warnings []Warning
}

type identityUpdate struct {
	OldInfo *identity.Info
	NewInfo *identity.Info
}

type claim struct {
	Name  string
	Value string
}

type UserQueries interface {
	GetRaw(userID string) (*user.User, error)
}

type UserCommands interface {
	Create(userID string) (*user.User, error)
	UpdateAccountStatus(userID string, accountStatus user.AccountStatus) error
}

type IdentityService interface {
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(info *identity.Info) error
	Delete(info *identity.Info) error
	Update(oldInfo *identity.Info, newInfo *identity.Info) error
	UpdateWithSpec(info *identity.Info, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	CheckDuplicated(info *identity.Info) (dup *identity.Info, err error)
	ListByClaim(name string, value string) ([]*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	Create(info *authenticator.Info) error
}

type VerifiedClaimService interface {
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	GetClaims(userID string) ([]*verification.Claim, error)
	DeleteClaim(claim *verification.Claim) error
}

type StandardAttributesService interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type CustomAttributesService interface {
	UpdateCustomAttributesWithList(role accesscontrol.Role, userID string, l attrs.List) error
}

type RolesGroupsCommands interface {
	ResetUserGroup(options *rolesgroups.ResetUserGroupOptions) error
	ResetUserRole(options *rolesgroups.ResetUserRoleOptions) error
}

type UserImportService struct {
	AppDatabase         *appdb.Handle
	LoginIDConfig       *config.LoginIDConfig
	Identities          IdentityService
	Authenticators      AuthenticatorService
	UserCommands        UserCommands
	UserQueries         UserQueries
	VerifiedClaims      VerifiedClaimService
	StandardAttributes  StandardAttributesService
	CustomAttributes    CustomAttributesService
	RolesGroupsCommands RolesGroupsCommands
	Logger              Logger
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("user-import")}
}

func (s *UserImportService) ImportRecords(ctx context.Context, request *Request) (*Summary, []Detail) {
	summary := Summary{}

	options := &Options{
		Upsert:     request.Upsert,
		Identifier: request.Identifier,
	}
	var details []Detail

	for idx, rawMessage := range request.Records {
		summary.Total += 1

		detail := Detail{
			Index:  idx,
			Record: rawMessage,
		}

		hasDetail := false
		var result importResult
		err := s.AppDatabase.WithTx(func() error {
			err := s.ImportRecordInTxn(ctx, &result, options, rawMessage)
			if err != nil {
				return err
			}
			return nil
		})

		if len(result.Warnings) > 0 {
			detail.Warnings = result.Warnings
			hasDetail = true
		}
		if err != nil {
			if !apierrors.IsAPIError(err) {
				s.Logger.WithError(err).Error(err.Error())
			}
			detail.Errors = []*apierrors.APIError{apierrors.AsAPIError(err)}
			hasDetail = true
		}

		if hasDetail {
			details = append(details, detail)
		}

		switch result.Outcome {
		case OutcomeInserted:
			summary.Inserted += 1
		case OutcomeUpdated:
			summary.Updated += 1
		case OutcomeSkipped:
			summary.Skipped += 1
		case OutcomeFailed:
			summary.Failed += 1
		default:
			summary.Failed += 1
		}
	}

	return &summary, details
}

func (s *UserImportService) ImportRecordInTxn(ctx context.Context, result *importResult, options *Options, rawMessage json.RawMessage) (err error) {
	var record Record

	err = options.RecordSchema().Validator().ParseJSONRawMessage(rawMessage, &record)
	if err != nil {
		return
	}

	var infos []*identity.Info
	switch options.Identifier {
	case IdentifierEmail:
		emailPtr, _ := record.Email()
		infos, err = s.Identities.ListByClaim(IdentifierEmail, *emailPtr)
	case IdentifierPhoneNumber:
		phoneNumberPtr, _ := record.PhoneNumber()
		infos, err = s.Identities.ListByClaim(IdentifierPhoneNumber, *phoneNumberPtr)
	case IdentifierPreferredUsername:
		preferredUsernamePtr, _ := record.PreferredUsername()
		infos, err = s.Identities.ListByClaim(IdentifierPreferredUsername, *preferredUsernamePtr)
	default:
		err = fmt.Errorf("unknown identifier: %v", options.Identifier)
	}
	if err != nil {
		return
	}

	switch len(infos) {
	case 0:
		return s.insertRecordInTxn(ctx, result, record)
	case 1:
		if options.Upsert {
			return s.upsertRecordInTxn(ctx, result, options, record, infos[0])
		} else {
			result.Outcome = OutcomeSkipped
			result.Warnings = append(result.Warnings, Warning{
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
	dupe, err := s.Identities.CheckDuplicated(info)
	if errors.Is(err, identity.ErrIdentityAlreadyExists) {
		err = api.NewInvariantViolated("DuplicatedIdentity", "identity already exists", map[string]interface{}{
			"login_id": dupe.LoginID.LoginID,
		})
		return
	}
	if err != nil {
		return
	}
	return
}

func (s *UserImportService) insertRecordInTxn(ctx context.Context, result *importResult, record Record) (err error) {
	userID := uuid.New()
	u, err := s.UserCommands.Create(userID)
	if err != nil {
		return
	}

	infos, err := s.insertIdentitiesInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertVerifiedClaimsInTxn(ctx, result, record, userID, infos)
	if err != nil {
		return
	}

	err = s.insertStandardAttributesInTxn(ctx, result, record, u)
	if err != nil {
		return
	}

	err = s.insertCustomAttributesInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertDisabledInTxn(ctx, result, record, u)
	if err != nil {
		return
	}

	err = s.insertRolesInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertGroupsInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertPasswordInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAPasswordInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAOOBOTPEmailInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFAOOBOTPPhoneInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	err = s.insertMFATOTPInTxn(ctx, result, record, userID)
	if err != nil {
		return
	}

	result.Outcome = OutcomeInserted
	return
}

func (s *UserImportService) insertIdentitiesInTxn(ctx context.Context, result *importResult, record Record, userID string) (infos []*identity.Info, err error) {
	var specs []*identity.Spec

	if emailPtr, ok := record.Email(); ok {
		if emailPtr == nil {
			result.Warnings = append(result.Warnings, Warning{
				Message: "email = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypeEmail)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				result.Warnings = append(result.Warnings, Warning{
					Message: "email is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypeEmail,
						Key:   key,
						Value: *emailPtr,
					},
				})
			}
		}
	}

	if phoneNumberPtr, ok := record.PhoneNumber(); ok {
		if phoneNumberPtr == nil {
			result.Warnings = append(result.Warnings, Warning{
				Message: "phone_number = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypePhone)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				result.Warnings = append(result.Warnings, Warning{
					Message: "phone_number is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypePhone,
						Key:   key,
						Value: *phoneNumberPtr,
					},
				})
			}

		}
	}

	if preferredUsernamePtr, ok := record.PreferredUsername(); ok {
		if preferredUsernamePtr == nil {
			result.Warnings = append(result.Warnings, Warning{
				Message: "preferred_username = null has no effect in insert.",
			})
		} else {
			key := string(model.LoginIDKeyTypeUsername)
			_, ok := s.LoginIDConfig.GetKeyConfig(key)
			if !ok {
				result.Warnings = append(result.Warnings, Warning{
					Message: "preferred_username is ignored because it is not an allowed login ID.",
				})
			} else {
				specs = append(specs, &identity.Spec{
					Type: model.IdentityTypeLoginID,
					LoginID: &identity.LoginIDSpec{
						Type:  model.LoginIDKeyTypeUsername,
						Key:   key,
						Value: *preferredUsernamePtr,
					},
				})
			}
		}
	}

	for _, spec := range specs {
		var info *identity.Info
		info, err = s.Identities.New(userID, spec, identity.NewIdentityOptions{
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

		err = s.Identities.Create(info)
		if err != nil {
			return
		}
	}

	return
}

func (s *UserImportService) insertVerifiedClaimsInTxn(ctx context.Context, result *importResult, record Record, userID string, infos []*identity.Info) (err error) {
	if emailVerified, emailVerifiedOK := record.EmailVerified(); emailVerifiedOK {
		if !emailVerified {
			result.Warnings = append(result.Warnings, Warning{
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
				result.Warnings = append(result.Warnings, Warning{
					Message: "email_verified = true has no effect when email is absent.",
				})
			} else {
				claim := s.VerifiedClaims.NewVerifiedClaim(userID, "email", email)
				err = s.VerifiedClaims.MarkClaimVerified(claim)
				if err != nil {
					return
				}
			}
		}
	}

	if phoneNumberVerified, phoneNumberVerifiedOK := record.PhoneNumberVerified(); phoneNumberVerifiedOK {
		if !phoneNumberVerified {
			result.Warnings = append(result.Warnings, Warning{
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
				result.Warnings = append(result.Warnings, Warning{
					Message: "phone_number_verified = true has no effect when phone_number is absent.",
				})
			} else {
				claim := s.VerifiedClaims.NewVerifiedClaim(userID, "phone_number", phoneNumber)
				err = s.VerifiedClaims.MarkClaimVerified(claim)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (s *UserImportService) insertStandardAttributesInTxn(ctx context.Context, result *importResult, record Record, u *user.User) (err error) {
	stdAttrsList := record.StandardAttributesList()

	stdAttrs, err := stdattrs.T(u.StandardAttributes).MergedWithList(stdAttrsList)
	if err != nil {
		return
	}

	err = s.StandardAttributes.UpdateStandardAttributes(accesscontrol.RoleGreatest, u.ID, stdAttrs)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertCustomAttributesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	customAttrsList := record.CustomAttributesList()
	err = s.CustomAttributes.UpdateCustomAttributesWithList(accesscontrol.RoleGreatest, userID, customAttrsList)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertDisabledInTxn(ctx context.Context, result *importResult, record Record, u *user.User) (err error) {
	disabled, ok := record.Disabled()
	if !ok {
		return
	}

	if !disabled {
		result.Warnings = append(result.Warnings, Warning{
			Message: "disabled = false has no effect in insert.",
		})
		return
	}

	accountStatus, err := u.AccountStatus().Disable(nil)
	if err != nil {
		return
	}

	err = s.UserCommands.UpdateAccountStatus(u.ID, *accountStatus)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertRolesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	roleKeys, ok := record.Roles()
	if !ok {
		return
	}

	err = s.RolesGroupsCommands.ResetUserRole(&rolesgroups.ResetUserRoleOptions{
		UserID:   userID,
		RoleKeys: roleKeys,
	})
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertGroupsInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	groupKeys, ok := record.Groups()
	if !ok {
		return
	}

	err = s.RolesGroupsCommands.ResetUserGroup(&rolesgroups.ResetUserGroupOptions{
		UserID:    userID,
		GroupKeys: groupKeys,
	})
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertPasswordInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	password, ok := record.Password()
	if !ok {
		return
	}
	passwordHash := Password(password).PasswordHash()

	spec := &authenticator.Spec{
		UserID:    userID,
		Type:      model.AuthenticatorTypePassword,
		IsDefault: false,
		Kind:      authenticator.KindPrimary,
		Password: &authenticator.PasswordSpec{
			PasswordHash: passwordHash,
		},
	}

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(info)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAPasswordInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	mfaObj, ok := record.MFA()
	if !ok {
		return
	}

	mfa := MFA(mfaObj)
	mfaPasswordObj, ok := mfa.Password()
	if !ok {
		return
	}

	passwordHash := Password(mfaPasswordObj).PasswordHash()

	spec := &authenticator.Spec{
		UserID:    userID,
		Type:      model.AuthenticatorTypePassword,
		IsDefault: false,
		Kind:      authenticator.KindSecondary,
		Password: &authenticator.PasswordSpec{
			PasswordHash: passwordHash,
		},
	}

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(info)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAOOBOTPEmailInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
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
		result.Warnings = append(result.Warnings, Warning{
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

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(info)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFAOOBOTPPhoneInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
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
		result.Warnings = append(result.Warnings, Warning{
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

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(info)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) insertMFATOTPInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
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

	info, err := s.Authenticators.New(spec)
	if err != nil {
		return
	}

	err = s.Authenticators.Create(info)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertRecordInTxn(ctx context.Context, result *importResult, options *Options, record Record, info *identity.Info) (err error) {
	err = s.upsertIdentitiesInTxn(ctx, result, options, record, info)
	if err != nil {
		return
	}

	err = s.upsertVerifiedClaimsInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertStandardAttributesInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertCustomAttributesInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertDisabledInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertRolesInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	err = s.upsertGroupsInTxn(ctx, result, record, info.UserID)
	if err != nil {
		return
	}

	// password update behavior is IGNORED.
	// mfa.password update behavior is IGNORED.
	// mfa.totp update behavior is IGNORED.

	result.Outcome = OutcomeUpdated
	return
}

func (s *UserImportService) upsertIdentitiesInTxnHelper(ctx context.Context, result *importResult, userID string, infos []*identity.Info, typ model.LoginIDKeyType, ptr *string) (err error) {
	if ptr == nil {
		err := s.removeIdentityInTxn(ctx, result, infos, typ)
		if err != nil {
			return err
		}
	} else {
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Type:  typ,
				Key:   string(typ),
				Value: *ptr,
			},
		}
		err := s.upsertIdentityInTxn(ctx, result, userID, infos, spec)
		if err != nil {
			return err
		}
	}
	return nil
}

// nolint: gocognit
func (s *UserImportService) upsertIdentitiesInTxn(ctx context.Context, result *importResult, options *Options, record Record, info *identity.Info) (err error) {
	userID := info.UserID
	infos, err := s.Identities.ListByUser(userID)
	if err != nil {
		return
	}

	switch options.Identifier {
	case IdentifierEmail:
		if phoneNumberPtr, phoneNumberOK := record.PhoneNumber(); phoneNumberOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypePhone, phoneNumberPtr)
			if err != nil {
				return
			}
		}

		if preferredUsernamePtr, preferredUsernameOK := record.PreferredUsername(); preferredUsernameOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypeUsername, preferredUsernamePtr)
			if err != nil {
				return
			}
		}
	case IdentifierPhoneNumber:
		if emailPtr, emailOK := record.Email(); emailOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypeEmail, emailPtr)
			if err != nil {
				return
			}
		}

		if preferredUsernamePtr, preferredUsernameOK := record.PreferredUsername(); preferredUsernameOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypeUsername, preferredUsernamePtr)
			if err != nil {
				return
			}
		}
	case IdentifierPreferredUsername:
		if emailPtr, emailOK := record.Email(); emailOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypeEmail, emailPtr)
			if err != nil {
				return
			}
		}

		if phoneNumberPtr, phoneNumberOK := record.PhoneNumber(); phoneNumberOK {
			err = s.upsertIdentitiesInTxnHelper(ctx, result, userID, infos, model.LoginIDKeyTypePhone, phoneNumberPtr)
			if err != nil {
				return
			}
		}
	default:
		err = fmt.Errorf("unknown identifier: %v", options.Identifier)
	}

	return
}

func (s *UserImportService) removeIdentityInTxn(ctx context.Context, result *importResult, infos []*identity.Info, typ model.LoginIDKeyType) error {
	var toBeRemoved []*identity.Info
	for _, info := range infos {
		info := info
		if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == typ && info.LoginID.LoginIDKey == string(typ) {
			toBeRemoved = append(toBeRemoved, info)
		}
	}

	for _, info := range toBeRemoved {
		err := s.Identities.Delete(info)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserImportService) upsertIdentityInTxn(ctx context.Context, result *importResult, userID string, infos []*identity.Info, spec *identity.Spec) error {
	var toBeUpdated []identityUpdate
	var toBeInserted []*identity.Info

	isUpdated := false
	for _, info := range infos {
		info := info

		if info.Type == model.IdentityTypeLoginID && info.LoginID.LoginIDType == spec.LoginID.Type && info.LoginID.LoginIDKey == spec.LoginID.Key {
			isUpdated = true
			updatedInfo, err := s.Identities.UpdateWithSpec(info, spec, identity.NewIdentityOptions{
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
		info, err := s.Identities.New(userID, spec, identity.NewIdentityOptions{
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

		err = s.Identities.Update(identityUpdate.OldInfo, identityUpdate.NewInfo)
		if err != nil {
			return err
		}
	}

	for _, info := range toBeInserted {
		err := s.checkIdentityDuplicate(ctx, info)
		if err != nil {
			return err
		}

		err = s.Identities.Create(info)
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

		verifiedClaim := s.VerifiedClaims.NewVerifiedClaim(userID, c.Name, c.Value)
		err := s.VerifiedClaims.MarkClaimVerified(verifiedClaim)
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

		err := s.VerifiedClaims.DeleteClaim(toBeDeleted)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserImportService) upsertEmailVerifiedInTxn(ctx context.Context, result *importResult, record Record, userID string, infos []*identity.Info, verifiedClaims []*verification.Claim) (err error) {
	if emailVerified, emailVerifiedOK := record.EmailVerified(); emailVerifiedOK {
		emailPtr, emailOK := record.Email()
		if !emailOK {
			result.Warnings = append(result.Warnings, Warning{
				Message: "email_verified has no effect when email is absent.",
			})
		} else if emailPtr == nil {
			result.Warnings = append(result.Warnings, Warning{
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

func (s *UserImportService) upsertPhoneNumberVerifiedInTxn(ctx context.Context, result *importResult, record Record, userID string, infos []*identity.Info, verifiedClaims []*verification.Claim) (err error) {
	if phoneNumberVerified, phoneNumberVerifiedOK := record.PhoneNumberVerified(); phoneNumberVerifiedOK {
		phoneNumberPtr, phoneNumberOK := record.PhoneNumber()
		if !phoneNumberOK {
			result.Warnings = append(result.Warnings, Warning{
				Message: "phone_number_verified has no effect when phone_number is absent.",
			})
		} else if phoneNumberPtr == nil {
			result.Warnings = append(result.Warnings, Warning{
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

func (s *UserImportService) upsertVerifiedClaimsInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	infos, err := s.Identities.ListByUser(userID)
	if err != nil {
		return
	}

	verifiedClaims, err := s.VerifiedClaims.GetClaims(userID)
	if err != nil {
		return
	}

	err = s.upsertEmailVerifiedInTxn(ctx, result, record, userID, infos, verifiedClaims)
	if err != nil {
		return
	}

	err = s.upsertPhoneNumberVerifiedInTxn(ctx, result, record, userID, infos, verifiedClaims)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertStandardAttributesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	u, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return
	}

	err = s.insertStandardAttributesInTxn(ctx, result, record, u)
	if err != nil {
		return
	}

	return
}

func (s *UserImportService) upsertCustomAttributesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	return s.insertCustomAttributesInTxn(ctx, result, record, userID)
}

func (s *UserImportService) upsertDisabledInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	disabled, ok := record.Disabled()
	if !ok {
		return
	}

	u, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return
	}

	if disabled {
		var accountStatus *user.AccountStatus
		accountStatus, err = u.AccountStatus().Disable(nil)
		if err != nil {
			return
		}
		err = s.UserCommands.UpdateAccountStatus(u.ID, *accountStatus)
		if err != nil {
			return
		}
	} else {
		var accountStatus *user.AccountStatus
		accountStatus, err = u.AccountStatus().Reenable()
		if err != nil {
			return
		}
		err = s.UserCommands.UpdateAccountStatus(u.ID, *accountStatus)
		if err != nil {
			return
		}
	}

	return
}

func (s *UserImportService) upsertRolesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	return s.insertRolesInTxn(ctx, result, record, userID)
}

func (s *UserImportService) upsertGroupsInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	return s.insertGroupsInTxn(ctx, result, record, userID)
}
