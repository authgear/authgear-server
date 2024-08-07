package ldap

import (
	"crypto/tls"
	"errors"
	"net/url"

	"github.com/go-ldap/ldap/v3"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/ldaputil"
)

const (
	sizeLimit        = 2     // Set to 2 to check 0 or more than 1 entry returned
	timeoutInSeconds = 10    // 10 seconds timeout, may want to make it configurable in the future
	typesOnly        = false // FALSE to return both attribute descriptions and values, TRUE to return attribute description only.
)

// We do not need pass controls.
var controls []ldap.Control = nil

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
		err = conn.StartTLS(&tls.Config{
			// According to https://pkg.go.dev/crypto/tls#Client
			// tls.Config must either InsecureSkipVerify=true, or set ServerName.
			//
			// According to https://pkg.go.dev/net/url#URL.Hostname
			// Hostname() is without port.
			//
			// According to https://cs.opensource.google/go/go/+/refs/tags/go1.22.6:src/net/http/transport.go;l=1658
			// tls.Config.ServerName expects host without port.
			ServerName: u.Hostname(),
		})
		if err != nil {
			// Reconnect to the server without TLS
			conn, err = ldap.DialURL(ldapURLString)
			if err != nil {
				return nil, err
			}
		}
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

func (c *Client) search(conn *ldap.Conn, searchFilter string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		c.Config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, sizeLimit, timeoutInSeconds,
		typesOnly,
		searchFilter,
		[]string{"*"}, // return all attributes
		controls,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return sr, nil
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

	sr, err := c.search(conn, searchFilter)
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

	uniqueIdentifierValue := entry.GetAttributeValue(c.Config.UserIDAttributeName)
	if uniqueIdentifierValue == "" {
		return nil, api.ErrInvalidCredentials
	}

	return sr.Entries[0], nil
}

func (c *Client) TestConnection(username string) error {
	conn, err := c.connect()
	if err != nil {
		return api.ErrLDAPCannotConnect
	}

	defer conn.Close()

	err = c.bind(conn)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return api.ErrLDAPFailedToBindSearchUser
		}
		return err
	}

	if username != "" {
		searchFilter, err := ldaputil.ParseFilter(c.Config.SearchFilterTemplate, username)
		if err != nil {
			return err
		}
		sr, err := c.search(conn, searchFilter)
		if err != nil {
			return err
		}
		if len(sr.Entries) == 0 {
			return api.ErrLDAPEndUserSearchNotFound
		}
		if len(sr.Entries) > 1 {
			return api.ErrLDAPEndUserSearchMultipleResult
		}

		uniqueIdentifierValue := sr.Entries[0].GetAttributeValue(c.Config.UserIDAttributeName)
		if uniqueIdentifierValue == "" {
			return api.ErrLDAPMissingUniqueAttribute
		}
	}

	return nil

}
