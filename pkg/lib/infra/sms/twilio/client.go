package twilio

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

type TwilioClient struct {
	Client            *http.Client
	TwilioCredentials *config.TwilioCredentials
}

func NewTwilioClient(c *config.TwilioCredentials) *TwilioClient {
	if c == nil {
		return nil
	}

	return &TwilioClient{
		Client:            utilhttputil.NewExternalClient(5 * time.Second),
		TwilioCredentials: c,
	}
}

func (t *TwilioClient) send0(ctx context.Context, opts smsapi.SendOptions) ([]byte, []byte, error) {
	// Written against
	// https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource

	u, err := url.Parse("https://api.twilio.com/2010-04-01/Accounts")
	if err != nil {
		return nil, nil, err
	}
	u = u.JoinPath(t.TwilioCredentials.AccountSID, "Messages.json")

	values := url.Values{}
	values.Set("Body", opts.Body)
	values.Set("To", string(opts.To))

	if t.TwilioCredentials.MessagingServiceSID != "" {
		values.Set("MessagingServiceSid", t.TwilioCredentials.MessagingServiceSID)
	} else {
		values.Set("From", opts.Sender)
	}

	requestBody := values.Encode()
	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(requestBody))

	// https://www.twilio.com/docs/usage/api#authenticate-with-http
	if t.TwilioCredentials.AuthToken != "" {
		// When Auth Token is used, username is Account SID, and password is Auth Token.
		req.SetBasicAuth(t.TwilioCredentials.AccountSID, t.TwilioCredentials.AuthToken)
	} else {
		// Normally we should not reach here.
		// But in case we do, we do not provide the auth header.
		// And Twilio should returns an error response to us in this case.
	}

	resp, err := t.Client.Do(req)
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
		})
	}

	return bodyBytes, dumpedResponse, nil
}

func (t *TwilioClient) Send(ctx context.Context, options smsapi.SendOptions) error {
	bodyBytes, dumpedResponse, err := t.send0(ctx, options)
	if err != nil {
		return err
	}

	sendResponse, err := ParseSendResponse(bodyBytes)
	if err != nil {
		return errors.Join(err, &smsapi.SendError{
			DumpedResponse: dumpedResponse,
		})
	}

	// Success case.
	if sendResponse.ErrorCode == nil {
		return nil
	}

	// Failed case.
	return &smsapi.SendError{
		DumpedResponse: dumpedResponse,
	}
}

var _ smsapi.Client = &TwilioClient{}
