package userimport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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

type UserCommands interface {
	Create(userID string) (*user.User, error)
	UpdateAccountStatus(userID string, accountStatus user.AccountStatus) error
}

type IdentityService interface {
	New(userID string, spec *identity.Spec, options identity.NewIdentityOptions) (*identity.Info, error)
	Create(info *identity.Info) error
	ListByClaim(name string, value string) ([]*identity.Info, error)
}

type VerifiedClaimService interface {
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
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
	UserCommands        UserCommands
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
			// TODO(userimport): update
			err = fmt.Errorf("upsert is not implemented yet")
			return
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

	err = s.insertStandardAttributesInTxn(ctx, result, record, userID)
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

func (s *UserImportService) insertStandardAttributesInTxn(ctx context.Context, result *importResult, record Record, userID string) (err error) {
	stdAttrs, ok := record.StandardAttributes()
	if !ok {
		return
	}

	err = s.StandardAttributes.UpdateStandardAttributes(accesscontrol.RoleGreatest, userID, stdAttrs)
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
