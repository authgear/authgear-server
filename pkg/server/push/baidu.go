package push

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/skygeario/skygear-server/pkg/server/push/baidu"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type BaiduAndroidPushMessage struct {
	Msg        baidu.AndroidNotificationMsg `json:"msg"`
	MsgExpires int                          `json:"msg_expires"`
}

type BaiduPusher struct {
	client *baidu.Client
}

func NewBaiduPusher(apiKey string, secretKey string) *BaiduPusher {
	return &BaiduPusher{
		client: baidu.NewClient(
			"https://api.tuisong.baidu.com/rest/3.0",
			apiKey,
			secretKey,
		),
	}
}

func (p *BaiduPusher) Send(m Mapper, device skydb.Device) error {
	if device.Type != "baidu-android" {
		return fmt.Errorf(`Want device.Type = "baidu-android", got %v`, device.Type)
	}

	message := BaiduAndroidPushMessage{}
	if err := mapBaiduMessage(m, &message); err != nil {
		log.Errorf("Failed to convert baidu message: %v", err)
		return err
	}

	req := baidu.NewPushSingleDeviceRequest(device.Token, message.Msg)
	req.MsgExpires = message.MsgExpires

	_, err := p.client.PushSingleDevice(req)

	return err
}

func mapBaiduMessage(mapper Mapper, msg *BaiduAndroidPushMessage) error {
	m := mapper.Map()
	if baiduMap, ok := m["baidu-android"].(map[string]interface{}); ok {
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

		return decoder.Decode(baiduMap)
	}

	return errors.New("push/baidu: push notification has no data")
}

var _ Sender = &BaiduPusher{}
