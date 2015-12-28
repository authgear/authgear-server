package push

import (
	log "github.com/Sirupsen/logrus"
	"github.com/google/go-gcm"
	"github.com/mitchellh/mapstructure"
	"github.com/oursky/skygear/skydb"
)

var gcmSendHTTP = gcm.SendHttp

// GCMPusher sends push notifications via GCM.
type GCMPusher struct {
	APIKey string
}

// Send sends the dictionary represented by m to device.
func (p *GCMPusher) Send(m Mapper, device skydb.Device) error {
	message := gcm.HttpMessage{}

	if err := mapGCMMessage(m, &message); err != nil {
		log.Errorf("Failed to convert gcm message: %v", err)
		return err
	}

	message.To = device.Token
	message.RegistrationIds = nil

	// NOTE(limouren): might need to check repsonse for deleted / invalid
	// device here
	if _, err := gcmSendHTTP(p.APIKey, message); err != nil {
		log.Errorf("Failed to send GCM Notification: %v", err)
		return err
	}

	return nil
}

func mapGCMMessage(mapper Mapper, msg *gcm.HttpMessage) error {
	m := mapper.Map()
	if gcmMap, ok := m["gcm"].(map[string]interface{}); ok {
		config := mapstructure.DecoderConfig{
			TagName: "json",
			Result:  msg,
		}
		// NewDecoder only returns error when DecoderConfig.Result
		// is not a pointer.
		decoder, err := mapstructure.NewDecoder(&config)
		if err != nil {
			panic(err)
		}

		if err := decoder.Decode(gcmMap); err != nil {
			return err
		}
	}

	return nil
}
