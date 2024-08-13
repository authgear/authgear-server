package user

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func TestAccountStatus(t *testing.T) {
	Convey("AccountStatus", t, func() {
		deleteAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		anonymizeAt := time.Date(2006, 1, 2, 3, 4, 5, 6, time.UTC)
		Convey("normal", func() {
			var normal AccountStatus
			var err error

			_, err = normal.Reenable()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")

			disabled, err := normal.Disable(nil)
			So(err, ShouldBeNil)
			So(disabled.Type(), ShouldEqual, AccountStatusTypeDisabled)

			scheduledDeletion, err := normal.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = normal.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: normal -> normal")
		})

		Convey("disable", func() {
			disabled := AccountStatus{
				IsDisabled: true,
			}
			var err error

			_, err = disabled.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: disabled -> disabled")

			normal, err := disabled.Reenable()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)

			scheduledDeletion, err := disabled.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduledDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)

			_, err = disabled.UnscheduleDeletionByAdmin()
			So(err, ShouldBeError, "invalid account status transition: disabled -> normal")
		})

		Convey("scheduled deletion by admin", func() {
			scheduledDeletion := AccountStatus{
				IsDisabled: true,
				DeleteAt:   &deleteAt,
			}
			var err error

			_, err = scheduledDeletion.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> disabled")

			_, err = scheduledDeletion.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> normal")

			_, err = scheduledDeletion.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_deletion_disabled -> scheduled_deletion_disabled")

			normal, err := scheduledDeletion.UnscheduleDeletionByAdmin()
			So(err, ShouldBeNil)
			So(normal.Type(), ShouldEqual, AccountStatusTypeNormal)
		})

		Convey("anonymize", func() {
			anonymized := AccountStatus{
				IsDisabled:   true,
				IsAnonymized: true,
				AnonymizeAt:  &anonymizeAt,
			}
			var err error

			_, err = anonymized.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: anonymized -> disabled")

			_, err = anonymized.Reenable()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> normal")

			_, err = anonymized.Anonymize()
			So(err, ShouldBeError, "invalid account status transition: anonymized -> anonymized")
		})

		Convey("scheduled anonymization by admin", func() {
			scheduledAnonymization := AccountStatus{
				IsDisabled:  true,
				AnonymizeAt: &anonymizeAt,
			}
			var err error

			_, err = scheduledAnonymization.Disable(nil)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> disabled")

			_, err = scheduledAnonymization.Reenable()
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> normal")

			_, err = scheduledAnonymization.ScheduleAnonymizationByAdmin(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> scheduled_anonymization_disabled")

			_, err = scheduledAnonymization.ScheduleDeletionByEndUser(deleteAt)
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> scheduled_deletion_deactivated")

			_, err = scheduledAnonymization.Anonymize()
			So(err, ShouldBeError, "invalid account status transition: scheduled_anonymization_disabled -> anonymized")

			unscheduleAnonymization, err := scheduledAnonymization.UnscheduleAnonymizationByAdmin()
			So(err, ShouldBeNil)
			So(unscheduleAnonymization.Type(), ShouldEqual, AccountStatusTypeNormal)

			scheduleDeletion, err := scheduledAnonymization.ScheduleDeletionByAdmin(deleteAt)
			So(err, ShouldBeNil)
			So(scheduleDeletion.Type(), ShouldEqual, AccountStatusTypeScheduledDeletionDisabled)
		})
	})
}

func TestComputeUserEndUserActionID(t *testing.T) {

	// Convey("EndUserAccountID", t, func() {
	// So((&User{}).EndUserAccountID(), ShouldEqual, "")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "email": "user@example.com",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user@example.com")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "preferred_username": "user",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "phone_number": "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "+85298765432")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "preferred_username": "user",
	// "phone_number":       "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user")
	// So((&User{
	// StandardAttributes: map[string]interface{}{
	// "email":              "user@example.com",
	// "preferred_username": "user",
	// "phone_number":       "+85298765432",
	// },
	// }).EndUserAccountID(), ShouldEqual, "user@example.com")
	// })
	Convey("ComputeUserEndUserActionID", t, func() {
		So(computeEndUserAccountID(&User{}, nil, nil), ShouldEqual, "")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"email": "user@example.com",
			},
		}, nil, nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"preferred_username": "user",
			},
		}, nil, nil), ShouldEqual, "user")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"phone_number": "+85298765432",
			},
		}, nil, nil), ShouldEqual, "+85298765432")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
		}, nil, nil), ShouldEqual, "user")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
		}, nil, nil), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(&User{
			StandardAttributes: map[string]interface{}{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
		}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					RawEntryJSON: map[string]interface{}{
						"dn": "cn=user,dc=example,dc=org",
					},
				},
			},
		}, &model.UserWeb3Info{
			Accounts: []model.NFTOwnership{
				{
					AccountIdentifier: model.AccountIdentifier{
						Address: "0x0",
					},
					NetworkIdentifier: model.NetworkIdentifier{
						Blockchain: "ethereum",
						Network:    "10",
					},
				},
			},
		}), ShouldEqual, "user@example.com")

		So(computeEndUserAccountID(&User{}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					RawEntryJSON: map[string]interface{}{
						"dn": "cn=user,dc=example,dc=org",
					},
				},
			},
		}, &model.UserWeb3Info{
			Accounts: []model.NFTOwnership{
				{
					AccountIdentifier: model.AccountIdentifier{
						Address: "0x0",
					},
					NetworkIdentifier: model.NetworkIdentifier{
						Blockchain: "ethereum",
						Network:    "10",
					},
				},
			},
		}), ShouldEqual, "cn=user,dc=example,dc=org")

		So(computeEndUserAccountID(&User{}, []*identity.Info{
			{
				Type: model.IdentityTypeLDAP,
				LDAP: &identity.LDAP{
					UserIDAttributeValue: "example-user",
				},
			},
		}, nil), ShouldEqual, "example-user")

		So(computeEndUserAccountID(&User{}, nil, &model.UserWeb3Info{
			Accounts: []model.NFTOwnership{
				{
					AccountIdentifier: model.AccountIdentifier{
						Address: "0x0",
					},
					NetworkIdentifier: model.NetworkIdentifier{
						Blockchain: "ethereum",
						Network:    "10",
					},
				},
			},
		}), ShouldEqual, "ethereum:0x0@10")
	})
}
