package event_test

import (
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

// payloadRegistry lists every payload type defined in pkg/api/event/blocking
// and pkg/api/event/nonblocking. Each entry is a zero-value pointer to the
// payload struct. A companion test asserts its length against the number of
// payload-defining source files in each package, so a new payload type
// cannot be silently omitted from this list.
var payloadRegistry = []any{
	&blocking.AuthenticationPostIdentifiedBlockingEventPayload{},
	&blocking.AuthenticationPreAuthenticatedBlockingEventPayload{},
	&blocking.AuthenticationPreInitializeBlockingEventPayload{},
	&blocking.OIDCIDTokenPreCreateBlockingEventPayload{},
	&blocking.OIDCJWTPreCreateBlockingEventPayload{},
	&blocking.UserPreCreateBlockingEventPayload{},
	&blocking.UserPreScheduleAnonymizationBlockingEventPayload{},
	&blocking.UserPreScheduleDeletionBlockingEventPayload{},
	&blocking.UserProfilePreUpdateBlockingEventPayload{},
	&nonblocking.AdminAPIMutationAddGroupToRolesExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddGroupToUsersExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddResourceToClientIDExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddRoleToGroupsExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddRoleToUsersExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddScopesToClientIDExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddUserToGroupsExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAddUserToRolesExecutedEventPayload{},
	&nonblocking.AdminAPIMutationAnonymizeUserExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateAuthenticatorExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateGroupExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateIdentityExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateResourceExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateRoleExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateScopeExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateSessionExecutedEventPayload{},
	&nonblocking.AdminAPIMutationCreateUserExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteAuthenticatorExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteAuthorizationExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteGroupExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteIdentityExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteResourceExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteRoleExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteScopeExecutedEventPayload{},
	&nonblocking.AdminAPIMutationDeleteUserExecutedEventPayload{},
	&nonblocking.AdminAPIMutationGenerateOOBOTPCodeExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetPasswordExpiredExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveGroupFromUsersExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveResourceFromClientIDExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveRoleFromUsersExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveScopesFromClientIDExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveUserFromGroupsExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRemoveUserFromRolesExecutedEventPayload{},
	&nonblocking.AdminAPIMutationReplaceScopesOfClientIDExecutedEventPayload{},
	&nonblocking.AdminAPIMutationResetAccountLockoutExecutedEventPayload{},
	&nonblocking.AdminAPIMutationResetPasswordExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRevokeAllSessionsExecutedEventPayload{},
	&nonblocking.AdminAPIMutationRevokeSessionExecutedEventPayload{},
	&nonblocking.AdminAPIMutationScheduleAccountAnonymizationExecutedEventPayload{},
	&nonblocking.AdminAPIMutationScheduleAccountDeletionExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSendResetPasswordMessageExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetAccountValidFromExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetAccountValidPeriodExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetAccountValidUntilExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetDisabledStatusExecutedEventPayload{},
	&nonblocking.AdminAPIMutationSetVerifiedStatusExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUnscheduleAccountAnonymizationExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUnscheduleAccountDeletionExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateGroupExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateIdentityExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateResourceExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateRoleExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateScopeExecutedEventPayload{},
	&nonblocking.AdminAPIMutationUpdateUserExecutedEventPayload{},
	&nonblocking.AuthenticationBlockedEventPayload{},
	&nonblocking.AuthenticationFailedEventPayload{},
	&nonblocking.AuthenticationFailedIdentityEventPayload{},
	&nonblocking.AuthenticationFailedLoginIDEventPayload{},
	&nonblocking.BotProtectionVerificationFailedEventPayload{},
	&nonblocking.EmailErrorEventPayload{},
	&nonblocking.EmailSentEventPayload{},
	&nonblocking.EmailSuppressedEventPayload{},
	&nonblocking.FraudProtectionDecisionRecordedEventPayload{},
	&nonblocking.IdentityBiometricDisabledEventPayload{},
	&nonblocking.IdentityBiometricEnabledEventPayload{},
	&nonblocking.IdentityLoginIDAddedEventPayload{},
	&nonblocking.IdentityLoginIDRemovedEventPayload{},
	&nonblocking.IdentityLoginIDUpdatedEventPayload{},
	&nonblocking.IdentityOAuthConnectedEventPayload{},
	&nonblocking.IdentityOAuthDisconnectedEventPayload{},
	&nonblocking.IdentityUnverifiedEventPayload{},
	&nonblocking.IdentityVerifiedEventPayload{},
	&nonblocking.M2MTokenCreatedEventPayload{},
	&nonblocking.ProjectAppCreatedEventPayload{},
	&nonblocking.ProjectAppSecretViewedEventPayload{},
	&nonblocking.ProjectAppUpdatedEventPayload{},
	&nonblocking.ProjectBillingCheckoutCreatedEventPayload{},
	&nonblocking.ProjectBillingSubscriptionCancelledEventPayload{},
	&nonblocking.ProjectBillingSubscriptionStatusUpdatedEventPayload{},
	&nonblocking.ProjectBillingSubscriptionUpdatedEventPayload{},
	&nonblocking.ProjectCollaboratorDeletedEventPayload{},
	&nonblocking.ProjectCollaboratorInvitationAcceptedEventPayload{},
	&nonblocking.ProjectCollaboratorInvitationCreatedEventPayload{},
	&nonblocking.ProjectCollaboratorInvitationDeletedEventPayload{},
	&nonblocking.ProjectDomainCreatedEventPayload{},
	&nonblocking.ProjectDomainDeletedEventPayload{},
	&nonblocking.ProjectDomainVerifiedEventPayload{},
	&nonblocking.RateLimitBlockedEventPayload{},
	&nonblocking.SMSErrorEventPayload{},
	&nonblocking.SMSSentEventPayload{},
	&nonblocking.SMSSuppressedEventPayload{},
	&nonblocking.UsageAlertTriggeredEventPayload{},
	&nonblocking.UserAnonymizationScheduledEventPayload{},
	&nonblocking.UserAnonymizationUnscheduledEventPayload{},
	&nonblocking.UserAnonymizedEventPayload{},
	&nonblocking.UserAnonymousPromotedEventPayload{},
	&nonblocking.UserAuthenticatedEventPayload{},
	&nonblocking.UserCreatedEventPayload{},
	&nonblocking.UserDeletedEventPayload{},
	&nonblocking.UserDeletionScheduledEventPayload{},
	&nonblocking.UserDeletionUnscheduledEventPayload{},
	&nonblocking.UserDisabledEventPayload{},
	&nonblocking.UserProfileUpdatedEventPayload{},
	&nonblocking.UserReauthenticatedEventPayload{},
	&nonblocking.UserReenabledEventPayload{},
	&nonblocking.UserSessionTerminatedEventPayload{},
	&nonblocking.UserSignedOutEventPayload{},
	&nonblocking.WhatsappErrorEventPayload{},
	&nonblocking.WhatsappOTPVerifiedEventPayload{},
	&nonblocking.WhatsappSentEventPayload{},
	&nonblocking.WhatsappSuppressedEventPayload{},
}

// countPayloadSourceFiles counts the *.go files in dir that define payload
// types, i.e. every file except test files and the shared helper file that
// carries no payload type of its own (util.go in blocking, project.go in
// nonblocking).
func countPayloadSourceFiles(t *testing.T, dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", dir, err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if name == "util.go" || name == "project.go" {
			continue
		}
		if !hasSuffix(name, ".go") || hasSuffix(name, "_test.go") {
			continue
		}
		count++
	}
	return count
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestPayloadRegistryIsComplete(t *testing.T) {
	Convey("payloadRegistry has one entry per payload-defining source file", t, func() {
		blockingCount := 0
		nonBlockingCount := 0
		for _, p := range payloadRegistry {
			switch p.(type) {
			case interface{ BlockingEventType() event.Type }:
				blockingCount++
			default:
				nonBlockingCount++
			}
		}

		So(blockingCount, ShouldEqual, countPayloadSourceFiles(t, "blocking"))
		So(nonBlockingCount, ShouldEqual, countPayloadSourceFiles(t, "nonblocking"))
	})
}

// TestResolveTagInvariant enforces that a payload with a resolve: tag never
// also declares its user deleted. A payload can be resolved from the
// database (resolve:), or it can carry a user that no longer exists in the
// database (DeletedUserIDs), but not both: resolve.go silently keeps a
// zero-valued user when the row is gone, and a webhook would observe that
// zero value. See event.Service.DispatchEventOnCommit for the code that
// depends on this invariant.
func TestResolveTagInvariant(t *testing.T) {
	Convey("no resolve:-tagged payload also declares a deleted user", t, func() {
		for _, p := range payloadRegistry {
			if !hasResolveTag(p) {
				continue
			}

			nonBlocking, ok := p.(event.NonBlockingPayload)
			if !ok {
				continue
			}

			Convey(reflect.TypeOf(p).Elem().String()+" has a resolve tag and must not declare a deleted user", func() {
				So(nonBlocking.DeletedUserIDs(), ShouldBeEmpty)
			})
		}
	})
}

func hasResolveTag(p any) bool {
	typ := reflect.TypeOf(p).Elem()
	for _, field := range reflect.VisibleFields(typ) {
		if _, ok := field.Tag.Lookup("resolve"); ok {
			return true
		}
	}
	return false
}
