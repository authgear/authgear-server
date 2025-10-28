package nexmo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	gsmcharset "github.com/go-gsm/charset"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

type NexmoClient struct {
	Client           *http.Client
	NexmoCredentials *config.NexmoCredentials
}

func NewNexmoClient(c *config.NexmoCredentials) *NexmoClient {
	if c == nil {
		return nil
	}

	return &NexmoClient{
		Client:           utilhttputil.NewExternalClient(5 * time.Second),
		NexmoCredentials: c,
	}
}

func (c *NexmoClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://developer.vonage.com/en/api/sms
	u, err := url.Parse("https://rest.nexmo.com/sms/json")
	if err != nil {
		return nil, nil, err
	}

	to := makeTo(opts.To)
	typ := inferType(opts.Body)

	values := url.Values{}
	values.Set("api_key", c.NexmoCredentials.APIKey)
	values.Set("api_secret", c.NexmoCredentials.APISecret)
	values.Set("from", opts.Sender)
	values.Set("to", to)
	values.Set("text", opts.Body)
	values.Set("type", typ)

	resp, err := utilhttputil.PostFormWithContext(ctx, c.Client, u.String(), values)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Join(err, &smsapi.SendError{
			DumpedResponse: dumpedResponse,
			ProviderName:   string(config.SMSProviderNexmo),
			ProviderType:   string(config.SMSProviderNexmo),
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (c *NexmoClient) Send(ctx context.Context, opts smsapi.SendOptions) error {
	bodyBytes, dumpedResponse, err := c.send0(ctx, opts)
	if err != nil {
		return err
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		return errors.Join(err, &smsapi.SendError{
			DumpedResponse: dumpedResponse,
			ProviderName:   string(config.SMSProviderNexmo),
			ProviderType:   string(config.SMSProviderNexmo),
		})
	}

	// To consider this is a success, we need all messages to have status==0
	// and messages MUST NOT be empty.
	if len(sendResponse.Messages) <= 0 {
		// Failed case.
		return &smsapi.SendError{
			DumpedResponse: dumpedResponse,
			ProviderName:   string(config.SMSProviderNexmo),
			ProviderType:   string(config.SMSProviderNexmo),
		}
	}
	for _, msg := range sendResponse.Messages {
		if msg.Status != "0" {
			return &smsapi.SendError{
				DumpedResponse: dumpedResponse,
				ProviderName:   string(config.SMSProviderNexmo),
				ProviderType:   string(config.SMSProviderNexmo),
			}
		}
	}

	return nil
}

func makeTo(e164 string) string {
	return strings.TrimPrefix(e164, "+")
}

func inferType(body string) string {
	// The doc says type needs to be unicode if the body contains
	// characters that cannot be encoded in the GSM Standard and Extended tables.
	if gsmcharset.IsGsmAlpha(body) {
		return "text"
	}
	return "unicode"
}

var _ smsapi.Client = (*NexmoClient)(nil)
