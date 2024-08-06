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
	Config config.LDAPServerConfig
	conn   *ldap.Conn
}

func NewClient(config config.LDAPServerConfig) *Client {
	return &Client{
		Config: config,
	}
}

func (c *Client) Connect() error {
	u, err := url.Parse(c.Config.URL)
	if err != nil {
		return err
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
		return err
	}

	if u.Scheme == "ldap" {
		_ = conn.StartTLS(nil)
	}

	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Bind(username string, password string) error {
	if c.conn == nil {
		panic("ldap: connection is not established")
	}
	err := c.conn.Bind(username, password)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AuthenticateUser(username string, password string) (*ldap.Entry, error) {
	searchFilter, err := ldaputil.ParseFilter(c.Config.SearchFilterTemplate, username)
	if err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		c.Config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 2, 10, false,
		searchFilter,
		[]string{},
		nil,
	)

	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(sr.Entries) != 1 {
		return nil, api.ErrInvalidCredentials
	}

	userDN := sr.Entries[0].DN
	err = c.conn.Bind(userDN, password)
	if err != nil {
		// Check if the error is due to invalid credentials
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return nil, errors.Join(api.ErrInvalidCredentials, err)
		}
		return nil, err
	}

	return sr.Entries[0], nil
}
