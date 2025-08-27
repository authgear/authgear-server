package graphql

import (
	"fmt"
	"time"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

var createUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(identityDef),
			Description: "Definition of the identity of new user.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "If null, then no password is created. If empty string, generate a password. Otherwise, create the specified password.",
		},
		"sendPassword": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Indicate whether to send the new password to the user.",
		},
		"setPasswordExpired": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Indicate whether the user is required to change password on next login.",
		},
	},
})

var createUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateUserPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"createUser",
	&graphql.Field{
		Description: "Create new user",
		Type:        graphql.NewNonNull(createUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createUserInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			defData := input["definition"].(map[string]interface{})
			identityDef, err := model.ParseIdentityDef(defData)
			if err != nil {
				return nil, err
			}

			password_, passwordSpecified := input["password"].(string)
			var password *string = nil
			if passwordSpecified {
				password = &password_
			}

			sendPassword, _ := input["sendPassword"].(bool)
			setPasswordExpired, _ := input["setPasswordExpired"].(bool)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			userID, err := gqlCtx.UserFacade.Create(ctx, identityDef, facade.CreatePasswordOptions{
				Password:           password,
				SendPassword:       sendPassword,
				SetPasswordExpired: setPasswordExpired,
			})
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateUserExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var resetPasswordInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ResetPasswordInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "New password.",
		},
		"sendPassword": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Indicate whether to send the new password to the user.",
		},
		"setPasswordExpired": &graphql.InputObjectFieldConfig{
			Type:        graphql.Boolean,
			Description: "Indicate whether the user is required to change password on next login.",
		},
	},
})

var resetPasswordPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResetPasswordPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"resetPassword",
	&graphql.Field{
		Description: "Reset password of user",
		Type:        graphql.NewNonNull(resetPasswordPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(resetPasswordInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			password, _ := input["password"].(string)
			generatePassword := password == ""
			sendPassword, _ := input["sendPassword"].(bool)
			setPasswordExpired, _ := input["setPasswordExpired"].(bool)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.ResetPassword(ctx, userID, password, generatePassword, sendPassword, setPasswordExpired)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationResetPasswordExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var setPasswordExpiredInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetPasswordExpiredInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"expired": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Indicate whether the user's password is expired.",
		},
	},
})

var setPasswordExpiredPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetPasswordExpiredPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setPasswordExpired",
	&graphql.Field{
		Description: "Force user to change password on next login",
		Type:        graphql.NewNonNull(setPasswordExpiredPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setPasswordExpiredInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			isExpired, _ := input["expired"].(bool)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.SetPasswordExpired(ctx, userID, isExpired)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationSetPasswordExpiredExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				Expired: isExpired,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var setMFAGracePeriodInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetMFAGracePeriodInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID",
		},
		"endAt": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.DateTime),
			Description: "Indicate when will user's MFA grace period end",
		},
	},
})

var setMFAGracePeriodPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetMFAGracePeriodPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setMFAGracePeriod",
	&graphql.Field{
		Description: "Grant user grace period for MFA enrollment",
		Type:        graphql.NewNonNull(setMFAGracePeriodPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setMFAGracePeriodInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			endAt := input["endAt"].(time.Time)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.SetMFAGracePeriod(ctx, userID, &endAt)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var removeMFAGracePeriodInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "removeMFAGracePeriodInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID",
		},
	},
})

var removeMFAGracePeriodPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "removeMFAGracePeriodPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"removeMFAGracePeriod",
	&graphql.Field{
		Description: "Revoke user grace period for MFA enrollment",
		Type:        graphql.NewNonNull(removeMFAGracePeriodPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeMFAGracePeriodInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.SetMFAGracePeriod(ctx, userID, nil)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var sendResetPasswordMessageInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SendResetPasswordMessageInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"loginID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target login ID.",
		},
	},
})

var _ = registerMutationField(
	"sendResetPasswordMessage",
	&graphql.Field{
		Description: "Send a reset password message to user",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(sendResetPasswordMessageInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			loginID := input["loginID"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			flowReference := authenticationflow.FlowReference{
				Type: authenticationflow.FlowTypeAccountRecovery,
				Name: "default",
			}

			flowObj, err := declarative.GetFlowRootObject(gqlCtx.Config, flowReference)
			if err != nil {
				return nil, err
			}

			jsonPointer, ok := authenticationflow.FindStepByType(flowObj, config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode)
			if !ok {
				panic(fmt.Errorf("unexpected: cannot find verify_account_recovery_code step"))
			}

			err = gqlCtx.ForgotPassword.SendCode(ctx, loginID, &forgotpassword.CodeOptions{
				IsAdminAPIResetPassword:       true,
				AuthenticationFlowType:        string(flowReference.Type),
				AuthenticationFlowName:        flowReference.Name,
				AuthenticationFlowJSONPointer: jsonPointer,
			})
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationSendResetPasswordMessageExecutedEventPayload{
				LoginID: loginID,
			})
			if err != nil {
				return nil, err
			}

			return nil, nil
		},
	},
)

var otpPurpose = graphql.NewEnum(graphql.EnumConfig{
	Name: "OTPPurpose",
	Values: graphql.EnumValueConfigMap{
		"LOGIN": &graphql.EnumValueConfig{
			Value: "login",
		},
		"VERIFICATION": &graphql.EnumValueConfig{
			Value: "verification",
		},
	},
})

type OTPPurpose string

const (
	OTPPurposeLogin        OTPPurpose = "login"
	OTPPurposeVerification OTPPurpose = "verification"
)

var generateOOBOTPCodeInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "GenerateOOBOTPCodeInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"purpose": &graphql.InputObjectFieldConfig{
			Type:        otpPurpose,
			Description: "Purpose of the generated OTP code.",
		},
		"target": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Target user's email or phone number.",
		},
	},
})

var generateOOBOTPCodePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "GenerateOOBOTPCodePayload",
	Fields: graphql.Fields{
		"code": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

var _ = registerMutationField(
	"generateOOBOTPCode",
	&graphql.Field{
		Description: "Generate OOB OTP code for user",
		Type:        graphql.NewNonNull(generateOOBOTPCodePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(generateOOBOTPCodeInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			target := input["target"].(string)

			purpose := OTPPurposeLogin
			if p, ok := input["purpose"].(string); ok {
				purpose = OTPPurpose(p)
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			var channel apimodel.AuthenticatorOOBChannel
			if err := phone.Require_IsPossibleNumber_IsValidNumber_UserInputInE164(target); err == nil {
				channel = apimodel.AuthenticatorOOBChannelSMS
			} else {
				channel = apimodel.AuthenticatorOOBChannelEmail
			}

			var kind otp.Kind
			switch purpose {
			case OTPPurposeLogin:
				kind = otp.KindOOBOTPCode(gqlCtx.Config, channel)
			case OTPPurposeVerification:
				kind = otp.KindVerification(gqlCtx.Config, channel)
			default:
				panic("admin: unknown purpose: " + purpose)
			}

			code, err := gqlCtx.OTPCode.GenerateOTP(
				ctx,
				kind,
				target,
				otp.FormCode,
				&otp.GenerateOptions{SkipRateLimits: true},
			)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload{
				Target:  target,
				Purpose: string(purpose),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"code": code,
			}).Value, nil
		},
	},
)

var setVerifiedStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetVerifiedStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"claimName": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Name of the claim to set verified status.",
		},
		"claimValue": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Value of the claim.",
		},
		"isVerified": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Indicate whether the target claim is verified.",
		},
	},
})

var setVerifiedStatusPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetVerifiedStatusPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setVerifiedStatus",
	&graphql.Field{
		Description: "Set verified status of a claim of user",
		Type:        graphql.NewNonNull(setVerifiedStatusPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setVerifiedStatusInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			claimName, _ := input["claimName"].(string)
			claimValue, _ := input["claimValue"].(string)
			isVerified, _ := input["isVerified"].(bool)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.VerificationFacade.SetVerified(ctx, userID, claimName, claimValue, isVerified)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationSetVerifiedStatusExecutedEventPayload{
				ClaimName:  claimName,
				ClaimValue: claimValue,
				IsVerified: isVerified,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var setDisabledStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetDisabledStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"isDisabled": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Indicate whether the target user is disabled.",
		},
		"reason": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Indicate the disable reason; If not provided, the user will be disabled with no reason.",
		},
	},
})

var setDisabledStatusPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetDisabledStatusPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setDisabledStatus",
	&graphql.Field{
		Description: "Set disabled status of user",
		Type:        graphql.NewNonNull(setDisabledStatusPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setDisabledStatusInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			isDisabled := input["isDisabled"].(bool)
			var reason *string
			if r, ok := input["reason"].(string); ok && isDisabled {
				reason = &r
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.SetDisabled(ctx, userID, isDisabled, reason)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationSetDisabledStatusExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
				IsDisabled: isDisabled,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var scheduleAccountDeletionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ScheduleAccountDeletionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var scheduleAccountDeletionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ScheduleAccountDeletionPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"scheduleAccountDeletion",
	&graphql.Field{
		Description: "Schedule account deletion",
		Type:        graphql.NewNonNull(scheduleAccountDeletionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(scheduleAccountDeletionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.ScheduleDeletion(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationScheduleAccountDeletionExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var unscheduleAccountDeletionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UnscheduleAccountDeletionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var unscheduleAccountDeletionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UnscheduleAccountDeletionPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"unscheduleAccountDeletion",
	&graphql.Field{
		Description: "Unschedule account deletion",
		Type:        graphql.NewNonNull(unscheduleAccountDeletionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(unscheduleAccountDeletionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.UnscheduleDeletion(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var updateUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"standardAttributes": &graphql.InputObjectFieldConfig{
			Type:        UserStandardAttributes,
			Description: "Whole standard attributes to be set on the user.",
		},
		"customAttributes": &graphql.InputObjectFieldConfig{
			Type:        UserCustomAttributes,
			Description: "Whole custom attributes to be set on the user.",
		},
	},
})

var updateUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateUserPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"updateUser",
	&graphql.Field{
		Description: "Update user",
		Type:        graphql.NewNonNull(updateUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateUserInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			stdAttrs, _ := input["standardAttributes"].(map[string]interface{})
			customAttrs, _ := input["customAttributes"].(map[string]interface{})

			err := gqlCtx.UserProfileFacade.UpdateUserProfile(
				ctx,
				accesscontrol.RoleGreatest,
				userID,
				stdAttrs,
				customAttrs,
			)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationUpdateUserExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var deleteUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var deleteUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteUserPayload",
	Fields: graphql.Fields{
		"deletedUserID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.ID),
		},
	},
})

var _ = registerMutationField(
	"deleteUser",
	&graphql.Field{
		Description: "Delete specified user",
		Type:        graphql.NewNonNull(deleteUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteUserInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			userModelVal, err := gqlCtx.Users.Load(ctx, userID).Value()
			// This is a footgun.
			// https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/
			if userModelVal == (*apimodel.User)(nil) {
				return nil, api.ErrUserNotFound
			}
			userModel := userModelVal.(*apimodel.User)

			if err != nil {
				return nil, err
			}

			err = gqlCtx.UserFacade.Delete(ctx, userModel.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteUserExecutedEventPayload{
				UserModel: *userModel,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"deletedUserID": userNodeID,
			}, nil
		},
	},
)

var scheduleAccountAnonymizationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ScheduleAccountAnonymizationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var scheduleAccountAnonymizationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ScheduleAccountAnonymizationPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"scheduleAccountAnonymization",
	&graphql.Field{
		Description: "Schedule account anonymization",
		Type:        graphql.NewNonNull(scheduleAccountAnonymizationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(scheduleAccountAnonymizationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.ScheduleAnonymization(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(
				ctx,
				&nonblocking.AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload{
					UserRef: apimodel.UserRef{
						Meta: apimodel.Meta{
							ID: userID,
						},
					},
				})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var unscheduleAccountAnonymizationInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UnscheduleAccountAnonymizationInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var unscheduleAccountAnonymizationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UnscheduleAccountAnonymizationPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"unscheduleAccountAnonymization",
	&graphql.Field{
		Description: "Unschedule account anonymization",
		Type:        graphql.NewNonNull(unscheduleAccountAnonymizationPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(unscheduleAccountAnonymizationInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.UnscheduleAnonymization(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(
				ctx,
				&nonblocking.AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload{
					UserRef: apimodel.UserRef{
						Meta: apimodel.Meta{
							ID: userID,
						},
					},
				})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, userID),
			}).Value, nil
		},
	},
)

var anonymizeUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AnonymizeUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var anonymizeUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AnonymizeUserPayload",
	Fields: graphql.Fields{
		"anonymizedUserID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.ID),
		},
	},
})

var _ = registerMutationField(
	"anonymizeUser",
	&graphql.Field{
		Description: "Anonymize specified user",
		Type:        graphql.NewNonNull(anonymizeUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(anonymizeUserInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			err := gqlCtx.UserFacade.Anonymize(ctx, userID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationAnonymizeUserExecutedEventPayload{
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: userID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"anonymizedUserID": userNodeID,
			}, nil
		},
	},
)
