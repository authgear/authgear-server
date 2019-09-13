package mfa

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrAuthenticatorNotFound is authenticator not found.
var ErrAuthenticatorNotFound = skyerr.NewError(skyerr.ResourceNotFound, "authenticator not found")

// Store manipulates authenticators
type Store interface {
	// GetRecoveryCode gets recovery codes.
	GetRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// GenerateRecoveryCode deletes the existing codes and generate new ones.
	GenerateRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error)
	// DeleteRecoveryCode deletes recovery codes.
	DeleteRecoveryCode(userID string) error
	// UpdateRecoveryCode updates recovery code authenticator.
	UpdateRecoveryCode(a *RecoveryCodeAuthenticator) error

	// DeleteBearerToken deletes bearer token of the given parent authenticator.
	DeleteBearerTokenByParentID(userID string, parentID string) error
	// DeleteAllBearerToken deletes all bearer token of the given user.
	DeleteAllBearerToken(userID string) error
	// CreateBearerToken creates Bearer Token authenticator.
	CreateBearerToken(a *BearerTokenAuthenticator) error
	// GetBearerTokenByToken gets bearer token authenticator by token.
	GetBearerTokenByToken(userID string, token string) (*BearerTokenAuthenticator, error)

	// ListAuthenticators returns a list of authenticators.
	// Either TOTPAuthenticator or OOBAuthenticator.
	ListAuthenticators(userID string) ([]interface{}, error)

	// CreateTOTP creates TOTP authenticator.
	CreateTOTP(a *TOTPAuthenticator) error
	// GetTOTP gets TOTP authenticator.
	GetTOTP(userID string, id string) (*TOTPAuthenticator, error)
	// UpdateTOTP updates TOTP authenticator.
	UpdateTOTP(a *TOTPAuthenticator) error
	// DeleteTOTP deletes TOTP authenticator.
	DeleteTOTP(a *TOTPAuthenticator) error

	// CreateOOB creates OOB authenticator.
	CreateOOB(a *OOBAuthenticator) error
	// GetOOB gets OOB authenticator.
	GetOOB(userID string, id string) (*OOBAuthenticator, error)
	// UpdateOOB updates OOB authenticator.
	UpdateOOB(a *OOBAuthenticator) error
	// DeleteOOB deletes OOB authenticator.
	DeleteOOB(a *OOBAuthenticator) error

	// GetValidOOBCode gets all valid OOB codes.
	GetValidOOBCode(userID string, t time.Time) ([]OOBCode, error)
	// CreateOOBCode creates OOB code.
	CreateOOBCode(c *OOBCode) error
	// DeleteOOBCode deletes OOB code.
	DeleteOOBCode(c *OOBCode) error
	// DeleteOOBCodeByAuthenticator deletes all OOB codes of the given authenticator.
	DeleteOOBCodeByAuthenticator(a *OOBAuthenticator) error
}
