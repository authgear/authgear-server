package authenticator

import (
	"github.com/authgear/authgear-server/pkg/api/model"
)

type Spec struct {
	UserID    string                  `json:"user_id,omitempty"`
	Type      model.AuthenticatorType `json:"type,omitempty"`
	IsDefault bool                    `json:"is_default,omitempty"`
	Kind      Kind                    `json:"kind,omitempty"`

	Password        *PasswordSpec        `json:"password,omitempty"`
	Passkey         *PasskeySpec         `json:"passkey,omitempty"`
	TOTP            *TOTPSpec            `json:"totp,omitempty"`
	OOBOTP          *OOBOTPSpec          `json:"oobotp,omitempty"`
	FaceRecognition *FaceRecognitionSpec `json:"face_recognition,omitempty"`
}
