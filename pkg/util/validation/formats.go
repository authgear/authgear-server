package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

func init() {
	jsonschemaformat.DefaultChecker["phone"] = FormatPhone{}
	jsonschemaformat.DefaultChecker["email"] = FormatEmail{AllowName: false}
	jsonschemaformat.DefaultChecker["email-name-addr"] = FormatEmail{AllowName: true}
	jsonschemaformat.DefaultChecker["uri"] = FormatURI{}
	jsonschemaformat.DefaultChecker["http_origin"] = FormatHTTPOrigin{}
	jsonschemaformat.DefaultChecker["wechat_account_id"] = FormatWeChatAccountID{}
	jsonschemaformat.DefaultChecker["bcp47"] = FormatBCP47{}
	jsonschemaformat.DefaultChecker["x_totp_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_oob_otp_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_verification_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_recovery_code"] = secretcode.RecoveryCode
}

// FormatPhone checks if input is a phone number in E.164 format.
// If the input is not a string, it is not an error.
// To enforce string, use other JSON schema constructs.
// This design allows this format to validate optional phone number.
type FormatPhone struct{}

func (f FormatPhone) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	return phone.EnsureE164(str)
}

// FormatEmail checks if input is an email address.
// If the input is not a string, it is not an error.
// To enforce string, use other JSON schema constructs.
// This design allows this format to validate optional email.
type FormatEmail struct {
	AllowName bool
}

func (f FormatEmail) CheckFormat(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return nil
	}

	addr, err := mail.ParseAddress(s)
	if err != nil {
		return err
	}

	if !f.AllowName && addr.Name != "" {
		return errors.New("input email must not have name")
	}

	return nil
}

// FormatURI checks if input is an absolute URI.
type FormatURI struct {
}

func (f FormatURI) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	if u.Scheme == "" || u.Host == "" {
		return errors.New("input URL must be absolute")
	}
	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}

	hasTrailingSlash := strings.HasSuffix(p, "/")
	cleaned := filepath.Clean(p)
	if hasTrailingSlash && !strings.HasSuffix(cleaned, "/") {
		cleaned = cleaned + "/"
	}

	if !filepath.IsAbs(p) || cleaned != p {
		return errors.New("input URL must be normalized")
	}

	return nil
}

// FormatHTTPOrigin checks if input is a valid origin with http/https scheme,
// host and optional port only.
type FormatHTTPOrigin struct {
}

func (f FormatHTTPOrigin) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("expect input URL with scheme http / https")
	}

	if u.Host == "" {
		return errors.New("expect input URL with non-empty host")
	}

	err = errors.New("expect input URL without user info, path, query and fragment")
	if u.User != nil {
		return err
	}
	if u.Path != "" || u.RawPath != "" {
		return err
	}
	if u.RawQuery != "" {
		return err
	}
	if u.Fragment != "" || u.RawFragment != "" {
		return err
	}

	return nil
}

// FormatWeChatAccountID checks if input start with gh_.
type FormatWeChatAccountID struct {
}

func (f FormatWeChatAccountID) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if !strings.HasPrefix(str, "gh_") {
		return errors.New("expect WeChat account id start with gh_")
	}

	return nil
}

type FormatBCP47 struct{}

func (f FormatBCP47) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	tag, err := language.Parse(str)
	if err != nil {
		return fmt.Errorf("invalid BCP 47 tag: %w", err)
	}

	canonical := tag.String()
	if str != canonical {
		return fmt.Errorf("non-canonical BCP 47 tag: %v != %v", str, canonical)
	}

	return nil
}
