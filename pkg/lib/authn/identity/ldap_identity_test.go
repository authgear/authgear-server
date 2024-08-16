package identity

import (
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLDAPIdentity(t *testing.T) {
	Convey("Test LDAP Identity", t, func() {
		Convey("Test UserIDAttributeValueDisplayValue", func() {
			Convey("It should return correct value if attribute is known", func() {
				ldap := &LDAP{
					UserIDAttributeName:  "uid",
					UserIDAttributeValue: []byte("example-user"),
				}
				So(ldap.UserIDAttributeValueDisplayValue(), ShouldEqual, "example-user")
			})
			Convey("It should return correct value if unknown attribute is a valid utf8 string", func() {
				ldap := &LDAP{
					UserIDAttributeName:  "unkown-attribute",
					UserIDAttributeValue: []byte("example-user"),
				}
				So(ldap.UserIDAttributeValueDisplayValue(), ShouldEqual, "example-user")
			})
			Convey("It should return a base64 encoded string if unknown attribute is not a valid utf8 string", func() {
				UUID := uuid.MustParse("8f4a9ad1-7325-3245-baaf-3d636a13d506")
				uuidBytes, err := UUID.MarshalBinary()
				So(err, ShouldBeNil)
				ldap := &LDAP{
					UserIDAttributeName:  "unkown-attribute",
					UserIDAttributeValue: uuidBytes,
				}
				So(ldap.UserIDAttributeValueDisplayValue(), ShouldEqual, "j0qa0XMlMkW6rz1jahPVBg==")
			})
		})

		Convey("Test EntryJSON", func() {
			Convey("It should only return known attribute", func() {
				ldap := &LDAP{
					RawEntryJSON: map[string]interface{}{
						"dn": "dn",
						"objectGUID": []interface{}{
							"j0qa0XMlMkW6rz1jahPVBg==",
						},
						"unknown-attr-0": []interface{}{
							"MTIzNA==",
						},
						"employeeID": []interface{}{
							"MTIzNA==",
						},
						"unknown-attr-1": []interface{}{
							"MTIzNA==",
						},
						"unknown-attr-2": []interface{}{
							"MTIzNA==",
						},
					},
				}
				So(ldap.EntryJSON(), ShouldResemble, map[string]interface{}{
					"dn": "dn",
					"objectGUID": []string{
						"8f4a9ad1-7325-3245-baaf-3d636a13d506",
					},
					"employeeID": []string{
						"1234",
					},
				})
			})
		})

		Convey("Test DisplayID", func() {
			Convey("It should DN if exists", func() {
				ldap := &LDAP{
					RawEntryJSON: map[string]interface{}{
						"dn": "dn",
						"objectGUID": []interface{}{
							"j0qa0XMlMkW6rz1jahPVBg==",
						},
						"employeeID": []interface{}{
							"MTIzNA==",
						},
					},
				}
				So(ldap.DisplayID(), ShouldEqual, "dn")
			})
			Convey("It should user id attribute if dn not exists", func() {
				ldap := &LDAP{
					UserIDAttributeName:  "uid",
					UserIDAttributeValue: []byte("example-user"),
					RawEntryJSON: map[string]interface{}{
						"objectGUID": []interface{}{
							"j0qa0XMlMkW6rz1jahPVBg==",
						},
						"employeeID": []interface{}{
							"MTIzNA==",
						},
					},
				}
				So(ldap.DisplayID(), ShouldEqual, "uid=example-user")
			})
			Convey("It should encode user id attribute correctly if dn not exists", func() {
				ldap := &LDAP{
					UserIDAttributeName:  "Some=Attribute",
					UserIDAttributeValue: []byte("ExampleUser"),
					RawEntryJSON: map[string]interface{}{
						"objectGUID": []interface{}{
							"j0qa0XMlMkW6rz1jahPVBg==",
						},
						"employeeID": []interface{}{
							"MTIzNA==",
						},
					},
				}
				So(ldap.DisplayID(), ShouldEqual, "Some\\=Attribute=ExampleUser")
			})
		})
	})
}
