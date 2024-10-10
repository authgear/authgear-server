package authenticator

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type FaceRecognition struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Kind             string    `json:"kind"`
	IsDefault        bool      `json:"is_default"`
	OpenCVFRPersonID string    `json:"opencv_fr_user_id"` // Person ID from https://sg.opencv.fr/
}

func (a *FaceRecognition) ToInfo() *Info {
	return &Info{
		ID:        a.ID,
		UserID:    a.UserID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Type:      model.AuthenticatorTypeFaceRecognition,
		Kind:      Kind(a.Kind),
		IsDefault: a.IsDefault,

		FaceRecognition: a,
	}
}
