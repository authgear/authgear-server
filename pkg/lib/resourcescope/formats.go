package resourcescope

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"

	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	jsonschemaformat.DefaultChecker["x_resource_uri"] = FormatResourceURI{}
	jsonschemaformat.DefaultChecker["x_scope_token"] = FormatScopeToken{}
}

type FormatResourceURI struct{}

func (FormatResourceURI) CheckFormat(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	// Opaque is set if the uri is not start with scheme://
	if u.Opaque != "" {
		return fmt.Errorf("resource URI must start with https://")
	}

	if u.User != nil {
		return fmt.Errorf("resource URI must not have user info")
	}

	if u.Host == "" {
		return fmt.Errorf("resource URI must have non-empty host")
	}

	switch u.Scheme {
	case "https":
		if u.RawQuery != "" {
			return fmt.Errorf("resource URI must not have query")
		}
		// url.Parse set Fragment, but we also check RawFragment here to ensure nothing is missed
		if u.Fragment != "" || u.RawFragment != "" {
			return fmt.Errorf("resource URI must not have fragment")
		}
		return validation.FormatURI{}.CheckFormat(ctx, value)
	default:
		return fmt.Errorf("invalid scheme: %v", u.Scheme)
	}
}

type FormatScopeToken struct{}

func (FormatScopeToken) CheckFormat(ctx context.Context, value interface{}) error {
	scope, ok := value.(string)
	if !ok {
		return nil
	}
	// See https://datatracker.ietf.org/doc/html/rfc6749#section-3.3
	tokenRe := regexp.MustCompile(`^[\x21\x23-\x5B\x5D-\x7E]+$`)
	if !tokenRe.MatchString(scope) {
		return fmt.Errorf("invalid scope-token: forbidden character")
	}

	return nil
}
