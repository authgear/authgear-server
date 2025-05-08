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
	"time"
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

type ErrorObject struct {
	Code string `json:"code"`
}

type LicenseObject struct {
	ExpireAt      *time.Time `json:"expire_at"`
	IsActivated   bool       `json:"is_activated"`
	IsExpired     bool       `json:"is_expired"`
	LicenseeEmail *string    `json:"licensee_email"`
}

type LicenseResponse struct {
	Data  *LicenseObject `json:"data"`
	Error *ErrorObject   `json:"error"`
}

func invokeLicenseEndpoint(ctx context.Context, client *http.Client, opts LicenseOptions, path string) (licenseObject *LicenseObject, err error) {
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
		return
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, &LicenseServerResponseError{DumpedResponse: dumpedResponse})
		}
	}()

	var licenseResponse LicenseResponse
	err = json.NewDecoder(resp.Body).Decode(&licenseResponse)
	if err != nil {
		return
	}

	if licenseResponse.Error != nil {
		switch licenseResponse.Error.Code {
		case "license_key_not_found":
			err = ErrLicenseServerLicenseKeyNotFound
			return
		case "license_key_already_activated":
			err = ErrLicenseServerLicenseKeyAlreadyActivated
			return
		default:
			err = ErrLicenseServerUnknownResponse
			return
		}
	}

	licenseObject = licenseResponse.Data
	return
}

func CheckLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (licenseObject *LicenseObject, err error) {
	return invokeLicenseEndpoint(ctx, client, opts, "/v1/license/check")
}

func ActivateLicense(ctx context.Context, client *http.Client, opts LicenseOptions) (licenseObject *LicenseObject, err error) {
	return invokeLicenseEndpoint(ctx, client, opts, "/v1/license/activate")
}
