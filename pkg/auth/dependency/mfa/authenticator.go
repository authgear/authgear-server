package mfa

import (
	"time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
)

type TOTPAuthenticator struct {
	ID          string
	UserID      string
	Type        coreAuth.AuthenticatorType
	Activated   bool
	CreatedAt   time.Time
	ActivatedAt *time.Time
	Secret      string
	DisplayName string
}

type OOBAuthenticator struct {
	ID          string
	UserID      string
	Type        coreAuth.AuthenticatorType
	Activated   bool
	CreatedAt   time.Time
	ActivatedAt *time.Time
	Channel     coreAuth.AuthenticatorOOBChannel
	Phone       string
	Email       string
}

type RecoveryCodeAuthenticator struct {
	ID        string
	UserID    string
	Type      coreAuth.AuthenticatorType
	Code      string
	CreatedAt time.Time
	Consumed  bool
}

type BearerTokenAuthenticator struct {
	ID        string
	UserID    string
	Type      coreAuth.AuthenticatorType
	ParentID  string
	Token     string
	CreatedAt time.Time
	ExpireAt  time.Time
}

type OOBCode struct {
	ID              string
	UserID          string
	AuthenticatorID string
	Code            string
	CreatedAt       time.Time
	ExpireAt        time.Time
}
