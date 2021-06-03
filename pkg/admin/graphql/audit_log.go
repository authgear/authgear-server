package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var auditLogActivityType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AuditLogActivityType",
	Values: graphql.EnumValueConfigMap{
		"USER_CREATED": &graphql.EnumValueConfig{
			Value: "user.created",
		},
		"USER_AUTHENTICATED": &graphql.EnumValueConfig{
			Value: "user.authenticated",
		},
		"USER_SIGNED_OUT": &graphql.EnumValueConfig{
			Value: "user.signed_out",
		},
		"USER_ANONYMOUS_PROMOTED": &graphql.EnumValueConfig{
			Value: "user.anonymous.promoted",
		},
		"AUTHENTICATION_IDENTITY_LOGIN_ID_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.identity.login_id.failed",
		},
		"AUTHENTICATION_IDENTITY_ANONYMOUS_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.identity.anonymous.failed",
		},
		"AUTHENTICATION_IDENTITY_BIOMETRIC_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.identity.biometric.failed",
		},
		"AUTHENTICATION_PRIMARY_PASSWORD_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.primary.password.failed",
		},
		"AUTHENTICATION_PRIMARY_OOB_OTP_EMAIL_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.primary.oob_otp_email.failed",
		},
		"AUTHENTICATION_PRIMARY_OOB_OTP_SMS_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.primary.oob_otp_sms.failed",
		},
		"AUTHENTICATION_SECONDARY_PASSWORD_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.secondary.password.failed",
		},
		"AUTHENTICATION_SECONDARY_TOTP_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.secondary.totp.failed",
		},
		"AUTHENTICATION_SECONDARY_OOB_OTP_EMAIL_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.secondary.oob_otp_email.failed",
		},
		"AUTHENTICATION_SECONDARY_OOB_OTP_SMS_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.secondary.oob_otp_sms.failed",
		},
		"AUTHENTICATION_SECONDARY_RECOVERY_CODE_FAILED": &graphql.EnumValueConfig{
			Value: "authentication.secondary.recovery_code.failed",
		},
		"IDENTITY_EMAIL_ADDED": &graphql.EnumValueConfig{
			Value: "identity.email.added",
		},
		"IDENTITY_EMAIL_REMOVED": &graphql.EnumValueConfig{
			Value: "identity.email.removed",
		},
		"IDENTITY_EMAIL_UPDATED": &graphql.EnumValueConfig{
			Value: "identity.email.updated",
		},
		"IDENTITY_PHONE_ADDED": &graphql.EnumValueConfig{
			Value: "identity.phone.added",
		},
		"IDENTITY_PHONE_REMOVED": &graphql.EnumValueConfig{
			Value: "identity.phone.removed",
		},
		"IDENTITY_PHONE_UPDATED": &graphql.EnumValueConfig{
			Value: "identity.phone.updated",
		},
		"IDENTITY_USERNAME_ADDED": &graphql.EnumValueConfig{
			Value: "identity.username.added",
		},
		"IDENTITY_USERNAME_REMOVED": &graphql.EnumValueConfig{
			Value: "identity.username.removed",
		},
		"IDENTITY_USERNAME_UPDATED": &graphql.EnumValueConfig{
			Value: "identity.username.updated",
		},
		"IDENTITY_OAUTH_CONNECTED": &graphql.EnumValueConfig{
			Value: "identity.oauth.connected",
		},
		"IDENTITY_OAUTH_DISCONNECTED": &graphql.EnumValueConfig{
			Value: "identity.oauth.disconnected",
		},
	},
})

const typeAuditLog = "AuditLog"

var nodeAuditLog = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeAuditLog,
		Description: "Audit log",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeUser, nil),
			"createdAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.DateTime),
			},
			"activityType": &graphql.Field{
				Type: graphql.NewNonNull(auditLogActivityType),
			},
			"user": &graphql.Field{
				Type: nodeUser,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*audit.Log)
					gqlCtx := GQLContext(p.Context)
					return gqlCtx.Users.Load(source.UserID).Value, nil
				},
			},
			"ipAddress": &graphql.Field{
				Type: graphql.String,
			},
			"userAgent": &graphql.Field{
				Type: graphql.String,
			},
			"clientID": &graphql.Field{
				Type: graphql.String,
			},
			"data": &graphql.Field{
				Type: AuditLogData,
			},
		},
	}),
	&audit.Log{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.AuditLogs.Load(id).Value, nil
	},
)

var connAuditLog = graphqlutil.NewConnectionDef(nodeAuditLog)
