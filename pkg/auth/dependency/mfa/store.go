package mfa

import (
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
}
