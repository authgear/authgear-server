package validation

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"
	"github.com/iawaknahc/originmatcher"
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/rolesgroupsutil"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/territoryutil"

	web3util "github.com/authgear/authgear-server/pkg/util/web3"
)

func init() {
	jsonschemaformat.DefaultChecker["phone"] = FormatPhone{}
	jsonschemaformat.DefaultChecker["email-name-addr"] = FormatEmail{AllowName: true}
	jsonschemaformat.DefaultChecker["uri"] = FormatURI{}
	jsonschemaformat.DefaultChecker["http_origin"] = FormatHTTPOrigin{}
	jsonschemaformat.DefaultChecker["http_origin_spec"] = FormatHTTPOriginSpec{}
	jsonschemaformat.DefaultChecker["ldap_url"] = FormatLDAPURL{}
	jsonschemaformat.DefaultChecker["ldap_dn"] = FormatLDAPDN{}
	jsonschemaformat.DefaultChecker["ldap_search_filter_template"] = FormatLDAPSearchFilterTemplate{}
	jsonschemaformat.DefaultChecker["ldap_attribute"] = FormatLDAPAttribute{}
	jsonschemaformat.DefaultChecker["wechat_account_id"] = FormatWeChatAccountID{}
	jsonschemaformat.DefaultChecker["bcp47"] = FormatBCP47{}
	jsonschemaformat.DefaultChecker["timezone"] = FormatTimezone{}
	jsonschemaformat.DefaultChecker["date-time"] = FormatDateTime{}
	jsonschemaformat.DefaultChecker["birthdate"] = FormatBirthdate{}
	jsonschemaformat.DefaultChecker["iso3166-1-alpha-2"] = FormatAlpha2{}
	jsonschemaformat.DefaultChecker["x_totp_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_oob_otp_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_verification_code"] = secretcode.OOBOTPSecretCode
	jsonschemaformat.DefaultChecker["x_recovery_code"] = secretcode.RecoveryCode
	jsonschemaformat.DefaultChecker["x_custom_attribute_pointer"] = FormatCustomAttributePointer{}
	jsonschemaformat.DefaultChecker["x_picture"] = FormatPicture{}
	jsonschemaformat.DefaultChecker["x_hook_uri"] = FormatHookURI{}
	jsonschemaformat.DefaultChecker["google_tag_manager_container_id"] = FormatGoogleTagManagerContainerID{}
	jsonschemaformat.DefaultChecker["x_web3_contract_id"] = FormatContractID{}
	jsonschemaformat.DefaultChecker["x_web3_network_id"] = FormatNetworkID{}
	jsonschemaformat.DefaultChecker["x_duration_string"] = FormatDurationString{}
	jsonschemaformat.DefaultChecker["x_base64_url"] = FormatBase64URL{}
	jsonschemaformat.DefaultChecker["x_re2_regex"] = FormatRe2Regex{}
	jsonschemaformat.DefaultChecker["x_role_group_key"] = rolesgroupsutil.FormatKey{}
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

	err := phone.Require_IsPossibleNumber_IsValidNumber_UserInputInE164(str)
	if err != nil {
		return err
	}

	return nil
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

	return FormatAbsolutePath{}.CheckFormat(p)
}

type FormatPicture struct{}

func (FormatPicture) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http":
		fallthrough
	case "https":
		return FormatURI{}.CheckFormat(value)
	case "authgearimages":
		if u.Host != "" {
			return errors.New("authgearimages URI does not have host")
		}
		p := u.EscapedPath()

		return FormatAbsolutePath{}.CheckFormat(p)
	default:
		return fmt.Errorf("invalid scheme: %v", u.Scheme)
	}
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

type FormatHTTPOriginSpec struct{}

func (FormatHTTPOriginSpec) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	err := originmatcher.CheckValidSpecStrict(str)
	if err != nil {
		return err
	}

	return nil
}

type FormatLDAPURL struct{}

func (FormatLDAPURL) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	if u.Scheme != "ldap" && u.Scheme != "ldaps" {
		return errors.New("expect input URL with scheme ldap / ldaps")
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

type FormatLDAPDN struct{}

func (FormatLDAPDN) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if str == "" {
		return errors.New("expect non-empty base DN")
	}

	dn, err := ldap.ParseDN(str)

	if err != nil || len(dn.RDNs) == 0 {
		return errors.New("invalid DN")
	}

	return nil

}

type FormatLDAPSearchFilterTemplate struct{}

func (FormatLDAPSearchFilterTemplate) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	tmpl, err := template.New("search_filter").Parse(str)
	tmplError := errors.New("invalid template")
	if err != nil {
		return tmplError
	}
	// check if the template can be execute with valid input (phone no., username, email)
	testcases := []string{"+85298765432", "username", "test@test.com"}
	for _, testcase := range testcases {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, map[string]string{"Username": testcase})
		if err != nil {
			return err
		}
		filterString := buf.String()
		filterString = strings.TrimSpace(filterString)
		_, err = ldap.CompileFilter(filterString)
		if err != nil {
			return errors.New("invalid search filter")
		}
	}

	return nil
}

type FormatLDAPAttribute struct{}

func (FormatLDAPAttribute) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if len(str) == 0 {
		return errors.New("expect non-empty attribute")
	}

	// An attribute description option is represented by the ABNF:
	// options = *( SEMI option )
	// option = 1*keychar
	// keychar = ALPHA / DIGIT / HYPHEN
	// According to https://datatracker.ietf.org/doc/html/rfc4512#section-2.5
	matched, err := regexp.MatchString(`^[a-zA-Z\d-]+$`, str)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid attribute")
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

type FormatTimezone struct{}

func (FormatTimezone) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	hasSlash := strings.Contains(str, "/")
	if !hasSlash {
		return fmt.Errorf("valid timezone name has at least 1 slash: %#v", str)
	}

	_, err := time.LoadLocation(str)
	if err != nil {
		return fmt.Errorf("invalid timezone name: %w", err)
	}

	return nil
}

type FormatDateTime struct{}

func (FormatDateTime) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("date-time must be in rfc3999 format")
	}

	return nil
}

type FormatBirthdate struct{}

func (FormatBirthdate) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if _, err := time.Parse("2006-01-02", str); err == nil {
		return nil
	}

	if _, err := time.Parse("0000-01-02", str); err == nil {
		return nil
	}

	if _, err := time.Parse("--01-02", str); err == nil {
		return nil
	}

	if _, err := time.Parse("2006", str); err == nil {
		return nil
	}

	return fmt.Errorf("invalid birthdate: %#v", str)
}

type FormatAlpha2 struct{}

func (FormatAlpha2) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	for _, allowed := range territoryutil.Alpha2 {
		if allowed == str {
			return nil
		}
	}

	return fmt.Errorf("invalid ISO 3166-1 alpha-2 code: %#v", str)
}

type FormatCustomAttributePointer struct{}

func (FormatCustomAttributePointer) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	p, err := jsonpointer.Parse(str)
	if err != nil {
		return err
	}

	if len(p) != 1 {
		return fmt.Errorf("custom attribute pointer must be one-level but found %v", len(p))
	}

	var runes []rune
	for _, r := range p[0] {
		runes = append(runes, r)
	}
	if len(runes) <= 0 {
		return fmt.Errorf("custom attribute pointer must not be empty")
	}

	checkStart := func(r rune) bool {
		return (r >= 'a' && r <= 'z')
	}

	checkEnd := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
	}

	check := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '_')
	}

	last := len(runes) - 1
	for i, r := range runes {
		var checkFunc func(rune) bool
		switch i {
		case 0:
			checkFunc = checkStart
		case last:
			checkFunc = checkEnd
		default:
			checkFunc = check
		}
		if !checkFunc(r) {
			return fmt.Errorf("invalid character at %v: %#v", i, string(r))
		}
	}

	return nil
}

type FormatGoogleTagManagerContainerID struct{}

func (FormatGoogleTagManagerContainerID) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	if !strings.HasPrefix(str, "GTM-") {
		return errors.New("expect google tag manager container ID to start with GTM-")
	}

	return nil
}

type FormatContractID struct{}

func (FormatContractID) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	contractID, err := web3util.ParseContractID(str)
	if err != nil {
		return fmt.Errorf("invalid contract ID: %#v", str)
	}

	if contractID.Blockchain == "ethereum" {
		if _, ok := model.ParseEthereumNetwork(contractID.Network); !ok {
			return fmt.Errorf("invalid ethereum chain ID: %#v", contractID.Network)
		}
	}

	return nil
}

type FormatNetworkID struct{}

func (FormatNetworkID) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	contractID, err := web3util.ParseContractID(str)
	if err != nil {
		return fmt.Errorf("invalid network ID: %#v", str)
	}

	if contractID.Address != "0x0" {
		return fmt.Errorf("invalid network ID: %#v", str)
	}

	return nil
}

type FormatAbsolutePath struct{}

func (FormatAbsolutePath) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if str == "" {
		str = "/"
	}

	hasTrailingSlash := strings.HasSuffix(str, "/")
	cleaned := path.Clean(str)
	if hasTrailingSlash && !strings.HasSuffix(cleaned, "/") {
		cleaned = cleaned + "/"
	}

	if !path.IsAbs(str) || cleaned != str {
		return fmt.Errorf("invalid path: %v", str)
	}

	return nil
}

type FormatHookURI struct{}

func (FormatHookURI) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http", "https":
		return FormatURI{}.CheckFormat(value)
	case "authgeardeno":
		if u.Host != "" {
			return fmt.Errorf("authgeardeno URI does not have host")
		}
		p := u.EscapedPath()

		return FormatAbsolutePath{}.CheckFormat(p)
	default:
		return fmt.Errorf("invalid scheme: %v", u.Scheme)
	}
}

type FormatDurationString struct{}

func (FormatDurationString) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	d, err := time.ParseDuration(str)
	if err != nil {
		return err
	} else if d <= 0 {
		return fmt.Errorf("non-positive duration %q", str)
	}

	return nil
}

type FormatBase64URL struct{}

func (FormatBase64URL) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	_, err := base64.RawURLEncoding.Strict().DecodeString(str)
	if err != nil {
		return err
	}
	return nil
}

type FormatRe2Regex struct{}

func (FormatRe2Regex) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	_, err := regexp.Compile(str)

	if err != nil {
		return err
	}
	return nil
}
