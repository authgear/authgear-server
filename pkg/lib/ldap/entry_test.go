package ldap

import (
	"encoding/base64"
	"testing"

	"github.com/go-ldap/ldap/v3"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEntry(t *testing.T) {
	Convey("Entry.ToJSON", t, func() {
		Convey("ByteValues should be converted to base64 strings", func() {
			originalBytes := []byte{1, 2, 3, 4, 5}

			attrName := "binaryAttribute"
			ldapAttribute := &ldap.EntryAttribute{
				Name:       attrName,
				ByteValues: [][]byte{originalBytes},
			}

			ldapEntry := &ldap.Entry{
				DN:         "cn=test,dc=example,dc=com",
				Attributes: []*ldap.EntryAttribute{ldapAttribute},
			}

			entry := &Entry{ldapEntry}

			jsonMap := entry.ToJSON()

			expectedBase64 := base64.StdEncoding.EncodeToString(originalBytes)

			So(jsonMap, ShouldContainKey, attrName)
			So(jsonMap[attrName], ShouldHaveLength, 1)
			So(jsonMap[attrName].([]interface{})[0], ShouldEqual, expectedBase64)
		})
	})
}
