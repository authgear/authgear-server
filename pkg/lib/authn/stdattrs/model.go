package stdattrs

import (
	"fmt"
	"reflect"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
	"github.com/authgear/authgear-server/pkg/util/nameutil"
)

type T map[string]interface{}

func (t T) FormattedName() string {
	// Choose between name or given_name+middle_name+family_name
	var name string
	if s, ok := t[Name].(string); ok && s != "" {
		name = s
	} else {
		givenName, _ := t[GivenName].(string)
		familyName, _ := t[FamilyName].(string)
		middleName, _ := t[MiddleName].(string)
		s := nameutil.Format(givenName, middleName, familyName)
		if s != "" {
			name = s
		}
	}

	nickname, _ := t[Nickname].(string)

	switch {
	case name == "" && nickname == "":
		return ""
	case name != "" && nickname == "":
		return name
	case name == "" && nickname != "":
		return nickname
	default:
		return fmt.Sprintf("%s (%s)", name, nickname)
	}
}

func (t T) FormattedNames() string {
	name, _ := t[Name].(string)

	givenName, _ := t[GivenName].(string)
	familyName, _ := t[FamilyName].(string)
	middleName, _ := t[MiddleName].(string)
	gmf := nameutil.Format(givenName, middleName, familyName)

	nickname, _ := t[Nickname].(string)

	isNameEmpty := name == ""
	isGmfEmpty := gmf == ""
	isNicknameEmpty := nickname == ""

	switch {
	case isNameEmpty && isGmfEmpty && isNicknameEmpty:
		return ""
	case isNameEmpty && isGmfEmpty && !isNicknameEmpty:
		return nickname
	case isNameEmpty && !isGmfEmpty && isNicknameEmpty:
		return gmf
	case isNameEmpty && !isGmfEmpty && !isNicknameEmpty:
		return fmt.Sprintf("%s (%s)", gmf, nickname)
	case !isNameEmpty && isGmfEmpty && isNicknameEmpty:
		return name
	case !isNameEmpty && isGmfEmpty && !isNicknameEmpty:
		return fmt.Sprintf("%s (%s)", name, nickname)
	case !isNameEmpty && !isGmfEmpty && isNicknameEmpty:
		return fmt.Sprintf("%s\n%s", name, gmf)
	case !isNameEmpty && !isGmfEmpty && !isNicknameEmpty:
		return fmt.Sprintf("%s (%s)\n%s", name, nickname, gmf)
	}
	return ""
}

func (t T) ToClaims() map[string]interface{} {
	return map[string]interface{}(t)
}

func (t T) WithNameCopiedToGivenName() T {
	out := make(T)
	for k, v := range t {
		out[k] = v
	}

	if name, ok := t[Name].(string); ok && name != "" {
		if _, ok := t[GivenName].(string); !ok {
			out[GivenName] = name
		}
	}
	return out
}

// NonIdentityAware returns a copy of t with identity-aware attributes removed.
func (t T) NonIdentityAware() T {
	out := make(T)
	for k1, val := range t {
		for _, k2 := range NonIdentityAwareKeys {
			if k1 == k2 {
				out[k1] = val
			}
		}
	}
	return out
}

// MergedWith returns a T with that merged into t.
func (t T) MergedWith(that T) T {
	out := make(T)
	for k, v := range t {
		out[k] = v
	}

	for k, v := range that {
		out[k] = v
	}

	return out
}

func (t T) Clone() T {
	out := make(T)
	for k, v := range t {
		if k == Address {
			address := make(map[string]interface{})
			if m, ok := v.(map[string]interface{}); ok {
				for mk, mv := range m {
					address[mk] = mv
				}
			}
			out[k] = address
		} else {
			out[k] = v
		}
	}
	return out
}

func (t T) Tidy() T {
	out := t.Clone()
	if address, ok := out[Address].(map[string]interface{}); ok {
		if len(address) <= 0 {
			delete(out, Address)
		}
	}
	return out
}

func (t T) MergedWithForm(ptrs map[string]string) (T, error) {
	out := t.Clone().ToClaims()
	for ptrStr, val := range ptrs {
		ptr, err := jsonpointer.Parse(ptrStr)
		if err != nil {
			return nil, err
		}
		if val == "" {
			err = jsonpointerutil.RemoveFromJSONObject(ptr, out)
			if err != nil {
				return nil, err
			}
		} else {
			err = jsonpointerutil.AssignToJSONObject(ptr, out, val)
			if err != nil {
				return nil, err
			}
		}
	}
	// All address fields may have been removed, so we should call Tidy here.
	return T(out).Tidy(), nil
}

func (t T) MergedWithList(l attrs.List) (T, error) {
	out := t.Clone().ToClaims()
	for _, attr := range l {
		ptr, err := jsonpointer.Parse(attr.Pointer)
		if err != nil {
			return nil, err
		}

		if attr.Value == nil {
			err = jsonpointerutil.RemoveFromJSONObject(ptr, out)
			if err != nil {
				return nil, err
			}
		} else {
			err = jsonpointerutil.AssignToJSONObject(ptr, out, attr.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	// All address fields may have been removed, so we should call Tidy here.
	return T(out).Tidy(), nil
}

func (t T) ReadWithAccessControl(accessControl accesscontrol.T, role accesscontrol.Role) T {
	out := T{}
	for key, val := range t {
		subject := accesscontrol.Subject(jsonpointer.T{key}.String())
		level := accessControl.GetLevel(subject, role, config.AccessControlLevelReadwrite)
		if level >= config.AccessControlLevelReadonly {
			out[key] = val
		}
	}
	return out
}

func (t T) CheckWrite(accessControl accesscontrol.T, role accesscontrol.Role, that T) error {
	check := func(subject accesscontrol.Subject) error {
		level := accessControl.GetLevel(subject, role, config.AccessControlLevelReadwrite)
		if level < config.AccessControlLevelReadwrite {
			return AccessControlViolated.NewWithDetails(
				fmt.Sprintf("%v being written by %v with level %v", subject, role, level),
				apierrors.Details{
					"subject": subject,
					"role":    role,
					"level":   level,
				},
			)
		}
		return nil
	}

	for key, val := range t {
		subject := accesscontrol.Subject(jsonpointer.T{key}.String())

		changed := false
		newVal, ok := that[key]
		if !ok {
			// Deletion
			changed = true
		} else if !reflect.DeepEqual(val, newVal) {
			// Edition
			changed = true
		}
		if changed {
			err := check(subject)
			if err != nil {
				return err
			}
		}
	}

	for key := range that {
		subject := accesscontrol.Subject(jsonpointer.T{key}.String())

		changed := false
		_, ok := t[key]
		if !ok {
			// Addition
			changed = true
		}
		if changed {
			err := check(subject)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t T) WithDerivedAttributesRemoved() T {
	out := t.Clone()
	delete(out, EmailVerified)
	delete(out, PhoneNumberVerified)
	delete(out, UpdatedAt)
	return out
}

const (
	// Sub is not used because we do not always use sub as the unique identifier for
	// an user from the identity provider.
	// Sub = "sub"
	Email               = "email"
	EmailVerified       = "email_verified"
	PhoneNumber         = "phone_number"
	PhoneNumberVerified = "phone_number_verified"
	PreferredUsername   = "preferred_username"
	FamilyName          = "family_name"
	GivenName           = "given_name"
	MiddleName          = "middle_name"
	Name                = "name"
	Nickname            = "nickname"
	Picture             = "picture"
	Profile             = "profile"
	Website             = "website"
	Gender              = "gender"
	Birthdate           = "birthdate"
	Zoneinfo            = "zoneinfo"
	Locale              = "locale"
	Address             = "address"
	Formatted           = "formatted"
	StreetAddress       = "street_address"
	Locality            = "locality"
	Region              = "region"
	PostalCode          = "postal_code"
	Country             = "country"
	UpdatedAt           = "updated_at"
)

var NonIdentityAwareKeys []string = []string{
	FamilyName,
	GivenName,
	MiddleName,
	Name,
	Nickname,
	Picture,
	Profile,
	Website,
	Gender,
	Birthdate,
	Zoneinfo,
	Locale,
	Address,
}
