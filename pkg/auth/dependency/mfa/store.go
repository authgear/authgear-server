package mfa

import (
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrAuthenticatorNotFound is authenticator not found.
var ErrAuthenticatorNotFound = skyerr.NewError(skyerr.ResourceNotFound, "authenticator not found")

// Store manipulates authenticators
type Store interface {
	// GetRecoveryCode gets recovery codes sorted alphabetically.
	GetRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// GenerateRecoveryCode deletes the existing codes and generate new ones.
	GenerateRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// DeleteRecoveryCode deletes recovery codes.
	DeleteRecoveryCode(userID string) error
	// UpdateRecoveryCode updates recovery code authenticator.
	UpdateRecoveryCode(a *RecoveryCodeAuthenticator) error

	// DeleteAllBearerToken deletes all bearer token of the given user.
	DeleteAllBearerToken(userID string) error
	// CreateBearerToken creates Bearer Token authenticator.
	CreateBearerToken(a *BearerTokenAuthenticator) error
	// GetBearerTokenByToken gets bearer token authenticator by token.
	GetBearerTokenByToken(userID string, token string) (*BearerTokenAuthenticator, error)

	// ListAuthenticators returns a list of authenticators ordered by activated at desc.
	// Either TOTPAuthenticator or OOBAuthenticator.
	ListAuthenticators(userID string) ([]Authenticator, error)

	// CreateTOTP creates TOTP authenticator.
	CreateTOTP(a *TOTPAuthenticator) error
	// GetTOTP gets TOTP authenticator.
	GetTOTP(userID string, id string) (*TOTPAuthenticator, error)
	// UpdateTOTP updates activated and activated_at of TOTP authenticator.
	UpdateTOTP(a *TOTPAuthenticator) error
	// DeleteTOTP deletes TOTP authenticator.
	DeleteTOTP(a *TOTPAuthenticator) error
	// DeleteInactiveTOTP deletes inactive TOTP authenticator.
	DeleteInactiveTOTP(userID string) error
	// GetOnlyInactiveTOTP gets the only TOTP authenticator.
	GetOnlyInactiveTOTP(userID string) (*TOTPAuthenticator, error)

	// CreateOOB creates OOB authenticator.
	CreateOOB(a *OOBAuthenticator) error
	// GetOOB gets OOB authenticator.
	GetOOB(userID string, id string) (*OOBAuthenticator, error)
	// UpdateOOB updates activated and activated_at of OOB authenticator.
	UpdateOOB(a *OOBAuthenticator) error
	// DeleteOOB deletes OOB authenticator.
	DeleteOOB(a *OOBAuthenticator) error
	// DeleteInactiveOOB deletes inactive OOB authenticator.
	DeleteInactiveOOB(userID string, exceptID string) error
	// GetOOBByChannel gets OOB authenticator by channel.
	GetOOBByChannel(userID string, channel coreAuth.AuthenticatorOOBChannel, phone string, email string) (*OOBAuthenticator, error)
	// GetOnlyInactiveOOB gets the only OOB authenticator.
	GetOnlyInactiveOOB(userID string) (*OOBAuthenticator, error)

	// GetValidOOBCode gets all valid OOB codes.
	GetValidOOBCode(userID string, t time.Time) ([]OOBCode, error)
	// CreateOOBCode creates OOB code.
	CreateOOBCode(c *OOBCode) error
	// DeleteOOBCode deletes OOB code.
	DeleteOOBCode(c *OOBCode) error
}
