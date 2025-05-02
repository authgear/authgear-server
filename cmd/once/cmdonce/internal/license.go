package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type LicenseServerResponseError struct {
	DumpedResponse []byte
}

func (e *LicenseServerResponseError) Error() string {
	return fmt.Sprintf("license server response: %v", base64.RawURLEncoding.EncodeToString(e.DumpedResponse))
}

type LicenseOptions struct {
	Endpoint    string
	LicenseKey  string
	Fingerprint string
}

func invokeLicenseEndpoint(ctx context.Context, client *http.Client, opts LicenseOptions, path string) (err error) {
	u, err := url.JoinPath(opts.Endpoint, path)
	if err != nil {
		return
	}

	data := make(url.Values)
	data.Set("license_key", opts.LicenseKey)
	data.Set("fingerprint", opts.Fingerprint)

	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(data.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, &LicenseServerResponseError{DumpedResponse: dumpedResponse})
		}
	}()

	var jsonBody map[string]any
	err = json.NewDecoder(resp.Body).Decode(&jsonBody)
	if err != nil {
		return err
	}

	if errorObj, ok := jsonBody["error"].(map[string]any); ok {
		code := errorObj["code"].(string)
		switch code {
		case "license_key_not_found":
			err = ErrLicenseServerLicenseKeyNotFound
			return
		case "license_key_already_activated":
			err = ErrLicenseServerLicenseKeyAlreadyActivated
			return
		case "license_key_expired":
			err = ErrLicenseServerLicenseKeyExpired
			return
		default:
			err = ErrLicenseServerUnknownResponse
			return
		}
	}

	return
}

func CheckLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (err error) {
	return invokeLicenseEndpoint(ctx, client, opts, "/v1/license/check")
}

func ActivateLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (err error) {
	return invokeLicenseEndpoint(ctx, client, opts, "/v1/license/activate")
}
