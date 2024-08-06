package ldap

import (
	"errors"
	"net/url"

	"github.com/go-ldap/ldap/v3"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/ldaputil"
)

type Client struct {
	Config       *config.LDAPServerConfig
	SecretConfig *config.LDAPServerUserCredentialsItem
}

func NewClient(config *config.LDAPServerConfig, secret *config.LDAPServerUserCredentialsItem) *Client {
	return &Client{
		Config:       config,
		SecretConfig: secret,
	}
}

func (c *Client) connect() (*ldap.Conn, error) {
	u, err := url.Parse(c.Config.URL)
	if err != nil {
		return nil, err
	}

	if u.Port() == "" {
		switch u.Scheme {
		case "ldap":
			u.Host = u.Host + ":389"
		case "ldaps":
			u.Host = u.Host + ":636"
		}
	}

	ldapURLString := u.String()

	conn, err := ldap.DialURL(ldapURLString)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "ldap" {
		_ = conn.StartTLS(nil)
	}

	return conn, nil
}

func (c *Client) bind(conn *ldap.Conn) error {
	username := c.SecretConfig.DN
	password := c.SecretConfig.Password
	err := conn.Bind(username, password)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AuthenticateUser(username string, password string) (*ldap.Entry, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	err = c.bind(conn)
	if err != nil {
		return nil, err
	}

	searchFilter, err := ldaputil.ParseFilter(c.Config.SearchFilterTemplate, username)
	if err != nil {
		return nil, err
	}

	// func NewSearchRequest(
	// 	 BaseDN string, Scope, DerefAliases,
	//   SizeLimit, (Set to 2 to check 0 or more than 1 entry returned)
	//   TimeLimit int, (10 seconds timeout)
	// 	 TypesOnly bool, (FALSE to return both attribute descriptions and values, TRUE to return attribute description only.)
	// 	 Filter string,
	// 	 Attributes []string, (nil means all attributes)
	// 	 Controls []Control,
	// ) *SearchRequest
	searchRequest := ldap.NewSearchRequest(
		c.Config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 2, 10, false,
		searchFilter,
		[]string{},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) != 1 {
		return nil, api.ErrInvalidCredentials
	}

	entry := sr.Entries[0]
	userDN := entry.DN
	err = conn.Bind(userDN, password)
	if err != nil {
		// Check if the error is due to invalid credentials
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, errors.Join(api.ErrInvalidCredentials, err)
		}
		return nil, err
	}

	uniqueIdentifierValue := entry.GetAttributeValue(c.Config.UserUniqueIdentifierAttribute)
	if uniqueIdentifierValue == "" {
		return nil, api.ErrInvalidCredentials
	}

	return sr.Entries[0], nil
}
