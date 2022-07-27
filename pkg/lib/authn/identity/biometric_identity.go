package identity

import (
	"time"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/deviceinfo"
)

type Biometric struct {
	ID         string                 `json:"id"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	UserID     string                 `json:"user_id"`
	KeyID      string                 `json:"key_id"`
	Key        []byte                 `json:"key"`
	DeviceInfo map[string]interface{} `json:"device_info"`
}

func (i *Biometric) ToJWK() (jwk.Key, error) {
	return jwk.ParseKey(i.Key)
}

func (i *Biometric) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeBiometric,

		Biometric: i,
	}
}

func (i *Biometric) FormattedDeviceInfo() string {
	return deviceinfo.DeviceModel(i.DeviceInfo)
}
