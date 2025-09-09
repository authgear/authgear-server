package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type CloudAPIClient struct {
	HTTPClient  HTTPClient
	Credentials *config.WhatsappCloudAPICredentials
}

func NewWhatsappCloudAPIClient(
	credentials *config.WhatsappCloudAPICredentials,
	httpClient HTTPClient,
) *CloudAPIClient {
	if credentials == nil {
		return nil
	}
	return &CloudAPIClient{
		HTTPClient:  httpClient,
		Credentials: credentials,
	}
}

func (c *CloudAPIClient) SendAuthenticationOTP(ctx context.Context, opts *SendAuthenticationOTPOptions, lang string) (messageID string, err error) {
	// Whatsapp Cloud API is Meta Graph API.
	// So the endpoint starts with https://graph.facebook.com
	// See https://developers.facebook.com/docs/whatsapp/cloud-api/overview#http-protocol
	//
	// The latest version of Meta Graph API is v22.0
	// See https://developers.facebook.com/docs/graph-api/changelog/
	endpoint, err := url.Parse("https://graph.facebook.com/v22.0")
	if err != nil {
		return "", err
	}

	// The API reference of this endpoint
	// https://developers.facebook.com/docs/whatsapp/cloud-api/reference/messages
	//
	// The general guide on using this endpoint.
	// https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages#requests
	//
	// The phone number format
	// In short, our E.164 format is fine.
	// https://developers.facebook.com/docs/whatsapp/cloud-api/reference/phone-numbers#whatsapp-user-phone-number-formats
	//
	// The specific guide for sending authentication template
	// Note that they do not provide documentation on sending other authentication template.
	// Only Copy Code authentication template is documented.
	// https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-message-templates/auth-otp-template-messages
	endpoint = endpoint.JoinPath(c.Credentials.PhoneNumberID, "messages")

	// I prefer we just use map here, instead of introducing lots of one-time use structs.
	bodyJSON := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                opts.To,
		"type":              "template",
		"template": map[string]interface{}{
			"name": c.Credentials.AuthenticationTemplateConfig.CopyCodeButton.Name,
			"language": map[string]interface{}{
				"policy": "deterministic",
				"code":   lang,
			},
			"components": []interface{}{
				map[string]interface{}{
					"type": "body",
					"parameters": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": opts.OTP,
						},
					},
				},
				map[string]interface{}{
					"type":     "button",
					"sub_type": "url",
					"index":    "0",
					"parameters": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": opts.OTP,
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(bodyJSON)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.String(), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// The access token is supposed to be a pre-generated non-expiring access token of a system user.
	// See https://developers.facebook.com/docs/whatsapp/cloud-api/overview#access-tokens
	//
	// The minimum permission to send message is not well documented.
	// Here is my observation.
	//
	// 1. You need to assign the app to the system user. The minimum permission is "Develop app".
	// 2. You need to assign the Whatsapp account to the system user. The minimum permission are:
	//   - Message templates (view only)
	//   - Phone numbers (view and manage)
	// 3. The access token itself needs "WhatsappCloudAPICredentials". See https://developers.facebook.com/docs/whatsapp/cloud-api/overview#permissions
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.Credentials.AccessToken))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Cloud API is hosted on top of Meta Graph API.
	// HTTP status 2xx means success.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var msgResp WhatsappSendMessageResponse
		err = json.NewDecoder(resp.Body).Decode(&msgResp)
		if err != nil {
			return
		}
		if len(msgResp.Messages) == 1 {
			return msgResp.Messages[0].ID, nil
		}
	}

	whatsappAPIError := &WhatsappAPIError{
		APIType:        config.WhatsappAPITypeCloudAPI,
		HTTPStatusCode: resp.StatusCode,
	}

	if dumpedResponse, err := httputil.DumpResponse(resp, true); err == nil {
		whatsappAPIError.DumpedResponse = dumpedResponse
	}

	errResp, err := c.tryParseErrorResponse(resp)
	if err != nil {
		return "", errors.Join(err, whatsappAPIError)
	}
	whatsappAPIError.CloudAPIResponse = errResp

	if errResp != nil {
		// This code path is not actually reachable because Cloud API does not report
		// invalid Whatsapp number in this endpoint.
		if errResp.Error.Code == cloudAPIErrorCodeMaybeInvalidUser {
			return "", errors.Join(ErrInvalidWhatsappUser, whatsappAPIError)
		}
	}

	return "", whatsappAPIError
}

func (c *CloudAPIClient) GetLanguages() []string {
	var out []string
	for _, code := range c.Credentials.AuthenticationTemplateConfig.CopyCodeButton.Languages {
		out = append(out, code)
	}
	return out
}

func (c *CloudAPIClient) tryParseErrorResponse(resp *http.Response) (*WhatsappCloudAPIErrorResponse, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// If we failed to read the response body, it is an error.
		return nil, err
	}

	// Error handling
	// https://developers.facebook.com/docs/graph-api/guides/error-handling
	var errResp WhatsappCloudAPIErrorResponse
	err = json.Unmarshal(respBody, &errResp)
	if err != nil {
		return nil, err
	}

	return &errResp, nil
}
