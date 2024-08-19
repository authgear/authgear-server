package mockldap

import (
	"fmt"

	ldap "github.com/lor00x/goldap/message"
	ldapserver "github.com/vjeantet/ldapserver"
	"sigs.k8s.io/yaml"
)

// FIXME
// Use a real yaml instead
var userData = `
- credential:
    password: readonly
  dn: cn=readonly,ou=bot,ou=service,dc=authgear,dc=com
  uid: readonly
  attributes:
    givenName: Readonly
    sn: User
    mail: noreply@example.com

- credential:
    password: jdoepassword
  dn: cn=jdoe,ou=people,ou=HK,dc=authgear,dc=com
  uid: jdoe
  attributes:
    givenName: John
    sn: Doe
    mail: jdoe@example.com

- credential:
    password: bjanepassword
  uid: bjane
  dn: cn=bjane,ou=people,ou=HK,dc=authgear,dc=com
  attributes:
    givenName: Jane
    sn: Bloggs
    mail: bjane@example.com

- credential:
    password: msmithpassword
  uid: msmith
  dn: cn=msmith,ou=people,ou=UK,dc=authgear,dc=com
  attributes:
    givenName: Michael
    sn: Smith
    mail: msmith@example.com

- credential:
    password: duplicatepassword
  uid: duplicate
  dn: cn=duplicate,ou=people,ou=UK,dc=authgear,dc=com
  attributes:
    givenName: Duplicated
    sn: User
    mail: duser@example.com

- credential:
    password: mockpassword
  dn: cn=mock,ou=people,ou=HK,dc=authgear,dc=com
  uid: mock
  attributes:
    givenName: John
    sn: Doe
    mail: mock@example.com
`

type LDAPRouteHandler struct {
	Users []User
}

func NewLDAPRouteHandler() (*LDAPRouteHandler, error) {
	var users []User
	data := []byte(userData)
	err := yaml.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}
	return &LDAPRouteHandler{
		Users: users,
	}, nil
}

func (s *LDAPRouteHandler) HandleSearch(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	searchRequest := m.GetSearchRequest()
	baseDN := searchRequest.BaseObject()
	if baseDN != "dc=authgear,dc=com" {
		res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
		w.Write(res)
		return
	}
	// Mock server only support `(uid=[uid])` now
	filterString := searchRequest.FilterString()
	for _, u := range s.Users {
		if filterString == fmt.Sprintf("(uid=%s)", u.UID) {
			e := ldapserver.NewSearchResultEntry(u.DN)
			e.AddAttribute(ldap.AttributeDescription("uid"), ldap.AttributeValue(u.UID))
			for k, v := range u.Attributes {
				e.AddAttribute(ldap.AttributeDescription(k), ldap.AttributeValue(v))
			}
			w.Write(e)
			// Magic filter for returning duplicated users
			if filterString == "(uid=duplicate)" {
				w.Write(e)
			}
		}
	}
	res := ldapserver.NewSearchResultDoneResponse(ldapserver.LDAPResultSuccess)
	w.Write(res)
}

func (s *LDAPRouteHandler) HandleBind(w ldapserver.ResponseWriter, m *ldapserver.Message) {
	r := m.GetBindRequest()
	if r.AuthenticationChoice() != "simple" {
		res := ldapserver.NewBindResponse(ldapserver.LDAPResultUnwillingToPerform)
		res.SetDiagnosticMessage("Authentication choice not supported")
		w.Write(res)
		return
	}

	for _, u := range s.Users {
		if r.Name() == ldap.LDAPDN(u.DN) && r.AuthenticationSimple().String() == u.Credential.Password {
			res := ldapserver.NewBindResponse(ldapserver.LDAPResultSuccess)
			w.Write(res)
			return
		}
	}

	res := ldapserver.NewBindResponse(ldapserver.LDAPResultInvalidCredentials)
	res.SetDiagnosticMessage("invalid credentials")
	w.Write(res)
}
