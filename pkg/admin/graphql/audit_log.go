package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
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
		"USER_PROFILE_UPDATED": &graphql.EnumValueConfig{
			Value: "user.profile.updated",
		},
		"USER_DISABLED": &graphql.EnumValueConfig{
			Value: "user.disabled",
		},
		"USER_REENABLED": &graphql.EnumValueConfig{
			Value: "user.reenabled",
		},
		"USER_SIGNED_OUT": &graphql.EnumValueConfig{
			Value: "user.signed_out",
		},
		"USER_SESSION_TERMINATED": &graphql.EnumValueConfig{
			Value: "user.session.terminated",
		},
		"USER_ANONYMOUS_PROMOTED": &graphql.EnumValueConfig{
			Value: "user.anonymous.promoted",
		},
		"USER_DELETION_SCHEDULED": &graphql.EnumValueConfig{
			Value: "user.deletion_scheduled",
		},
		"USER_DELETION_UNSCHEDULED": &graphql.EnumValueConfig{
			Value: "user.deletion_unscheduled",
		},
		"USER_DELETED": &graphql.EnumValueConfig{
			Value: "user.deleted",
		},
		"USER_ANONYMIZATION_SCHEDULED": &graphql.EnumValueConfig{
			Value: "user.anonymization_scheduled",
		},
		"USER_ANONYMIZATION_UNSCHEDULED": &graphql.EnumValueConfig{
			Value: "user.anonymization_unscheduled",
		},
		"USER_ANONYMIZED": &graphql.EnumValueConfig{
			Value: "user.anonymized",
		},
		"BOT_PROTECTION_VERIFICATION_FAILED": &graphql.EnumValueConfig{
			Value: "bot_protection.verification.failed",
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
		"IDENTITY_EMAIL_VERIFIED": &graphql.EnumValueConfig{
			Value: "identity.email.verified",
		},
		"IDENTITY_EMAIL_UNVERIFIED": &graphql.EnumValueConfig{
			Value: "identity.email.unverified",
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
		"IDENTITY_PHONE_VERIFIED": &graphql.EnumValueConfig{
			Value: "identity.phone.verified",
		},
		"IDENTITY_PHONE_UNVERIFIED": &graphql.EnumValueConfig{
			Value: "identity.phone.unverified",
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
		"IDENTITY_BIOMETRIC_ENABLED": &graphql.EnumValueConfig{
			Value: "identity.biometric.enabled",
		},
		"IDENTITY_BIOMETRIC_DISABLED": &graphql.EnumValueConfig{
			Value: "identity.biometric.disabled",
		},
		"EMAIL_SENT": &graphql.EnumValueConfig{
			Value: "email.sent",
		},
		"SMS_SENT": &graphql.EnumValueConfig{
			Value: "sms.sent",
		},
		"WHATSAPP_SENT": &graphql.EnumValueConfig{
			Value: "whatsapp.sent",
		},
		"WHATSAPP_OTP_VERIFIED": &graphql.EnumValueConfig{
			Value: "whatsapp.otp.verified",
		},
		"ADMIN_API_MUTATION_ANONYMIZE_USER_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.anonymize_user.executed",
		},
		"ADMIN_API_MUTATION_CREATE_IDENTITY_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_identity.executed",
		},
		"ADMIN_API_MUTATION_CREATE_SESSION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_session.executed",
		},
		"ADMIN_API_MUTATION_CREATE_USER_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_user.executed",
		},
		"ADMIN_API_MUTATION_CREATE_AUTHENTICATOR_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_authenticator.executed",
		},
		"ADMIN_API_MUTATION_DELETE_AUTHENTICATOR_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_authenticator.executed",
		},
		"ADMIN_API_MUTATION_DELETE_AUTHORIZATION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_authorization.executed",
		},
		"ADMIN_API_MUTATION_DELETE_IDENTITY_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_identity.executed",
		},
		"ADMIN_API_MUTATION_DELETE_USER_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_user.executed",
		},
		"ADMIN_API_MUTATION_GENERATE_OOB_OTP_CODE_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.generate_oob_otp_code.executed",
		},
		"ADMIN_API_MUTATION_RESET_PASSWORD_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.reset_password.executed",
		},
		"ADMIN_API_MUTATION_REVOKE_ALL_SESSIONS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.revoke_all_sessions.executed",
		},
		"ADMIN_API_MUTATION_REVOKE_SESSION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.revoke_session.executed",
		},
		"ADMIN_API_MUTATION_SCHEDULE_ACCOUNT_ANONYMIZATION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.schedule_account_anonymization.executed",
		},
		"ADMIN_API_MUTATION_SCHEDULE_ACCOUNT_DELETION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.schedule_account_deletion.executed",
		},
		"ADMIN_API_MUTATION_SEND_RESET_PASSWORD_MESSAGE_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.send_reset_password_message.executed",
		},
		"ADMIN_API_MUTATION_SET_DISABLED_STATUS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.set_disabled_status.executed",
		},
		"ADMIN_API_MUTATION_SET_VERIFIED_STATUS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.set_verified_status.executed",
		},
		"ADMIN_API_MUTATION_UNSCHEDULE_ACCOUNT_ANONYMIZATION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.unschedule_account_anonymization.executed",
		},
		"ADMIN_API_MUTATION_UNSCHEDULE_ACCOUNT_DELETION_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.unschedule_account_deletion.executed",
		},
		"ADMIN_API_MUTATION_UPDATE_IDENTITY_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.update_identity.executed",
		},
		"ADMIN_API_MUTATION_UPDATE_USER_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.update_user.executed",
		},
		"ADMIN_API_MUTATION_UPDATE_ROLE_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.update_role.executed",
		},
		"ADMIN_API_MUTATION_UPDATE_GROUP_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.update_group.executed",
		},
		"ADMIN_API_MUTATION_ADD_GROUP_TO_ROLES_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_group_to_roles.executed",
		},
		"ADMIN_API_MUTATION_ADD_ROLE_TO_GROUPS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_role_to_groups.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_GROUP_FROM_ROLES_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_group_from_roles.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_ROLE_FROM_GROUPS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_role_from_groups.executed",
		},
		"ADMIN_API_MUTATION_ADD_GROUP_TO_USERS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_group_to_users.executed",
		},
		"ADMIN_API_MUTATION_ADD_USER_TO_GROUPS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_user_to_groups.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_GROUP_FROM_USERS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_group_from_users.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_USER_FROM_GROUPS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_user_from_groups.executed",
		},
		"ADMIN_API_MUTATION_ADD_ROLE_TO_USERS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_role_to_users.executed",
		},
		"ADMIN_API_MUTATION_ADD_USER_TO_ROLES_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.add_user_to_roles.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_ROLE_FROM_USERS_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_role_from_users.executed",
		},
		"ADMIN_API_MUTATION_REMOVE_USER_FROM_ROLES_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.remove_user_from_roles.executed",
		},
		"ADMIN_API_MUTATION_DELETE_GROUP_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_group.executed",
		},
		"ADMIN_API_MUTATION_DELETE_ROLE_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.delete_role.executed",
		},
		"ADMIN_API_MUTATION_CREATE_GROUP_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_group.executed",
		},
		"ADMIN_API_MUTATION_CREATE_ROLE_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.create_role.executed",
		},
		"ADMIN_API_MUTATION_SET_PASSWORD_EXPIRED_EXECUTED": &graphql.EnumValueConfig{
			Value: "admin_api.mutation.set_password_expired.executed",
		},
		"PROJECT_APP_UPDATED": &graphql.EnumValueConfig{
			Value: "project.app.updated",
		},
		"PROJECT_APP_SECRET_VIEWED": &graphql.EnumValueConfig{
			Value: "project.app.secret.viewed",
		},
		"PROJECT_BILLING_CHECKOUT_CREATED": &graphql.EnumValueConfig{
			Value: "project.billing.checkout.created",
		},
		"PROJECT_BILLING_SUBSCRIPTION_CANCELLED": &graphql.EnumValueConfig{
			Value: "project.billing.subscription.cancelled",
		},
		"PROJECT_BILLING_SUBSCRIPTION_STATUS_UPDATED": &graphql.EnumValueConfig{
			Value: "project.billing.subscription.status.updated",
		},
		"PROJECT_BILLING_SUBSCRIPTION_UPDATED": &graphql.EnumValueConfig{
			Value: "project.billing.subscription.updated",
		},
		"PROJECT_COLLABORATOR_DELETED": &graphql.EnumValueConfig{
			Value: "project.collaborator.deleted",
		},
		"PROJECT_COLLABORATOR_INVITATION_ACCEPTED": &graphql.EnumValueConfig{
			Value: "project.collaborator.invitation.accepted",
		},
		"PROJECT_COLLABORATOR_INVITATION_CREATED": &graphql.EnumValueConfig{
			Value: "project.collaborator.invitation.created",
		},
		"PROJECT_COLLABORATOR_INVITATION_DELETED": &graphql.EnumValueConfig{
			Value: "project.collaborator.invitation.deleted",
		},
		"PROJECT_DOMAIN_CREATED": &graphql.EnumValueConfig{
			Value: "project.domain.created",
		},
		"PROJECT_DOMAIN_DELETED": &graphql.EnumValueConfig{
			Value: "project.domain.deleted",
		},
		"PROJECT_DOMAIN_VERIFIED": &graphql.EnumValueConfig{
			Value: "project.domain.verified",
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
			"id": relay.GlobalIDField(typeAuditLog, nil),
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
