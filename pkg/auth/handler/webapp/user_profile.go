package webapp

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

type UserProfile map[string]interface{}

func makeIsPresentAttr(name string) string {
	return name + "_is_present"
}

func extractString(attrs map[string]interface{}, output UserProfile, key string) {
	if value, ok := attrs[key].(string); ok && value != "" {
		output[makeIsPresentAttr(key)] = true
		output[key] = value
	} else {
		output[makeIsPresentAttr(key)] = false
		output[key] = ""
	}
}

func extractBool(attrs map[string]interface{}, output UserProfile, key string) {
	if value, ok := attrs[key].(bool); ok {
		output[makeIsPresentAttr(key)] = true
		output[key] = value
	} else {
		output[makeIsPresentAttr(key)] = false
		output[key] = false
	}
}

func GetUserProfile(user *model.User) UserProfile {
	userProfile := UserProfile{}

	extractString(user.StandardAttributes, userProfile, stdattrs.Email)
	extractBool(user.StandardAttributes, userProfile, stdattrs.EmailVerified)
	extractString(user.StandardAttributes, userProfile, stdattrs.PhoneNumber)
	extractBool(user.StandardAttributes, userProfile, stdattrs.PhoneNumberVerified)
	extractString(user.StandardAttributes, userProfile, stdattrs.PreferredUsername)
	extractString(user.StandardAttributes, userProfile, stdattrs.FamilyName)
	extractString(user.StandardAttributes, userProfile, stdattrs.GivenName)
	extractString(user.StandardAttributes, userProfile, stdattrs.MiddleName)
	extractString(user.StandardAttributes, userProfile, stdattrs.Name)
	extractString(user.StandardAttributes, userProfile, stdattrs.Nickname)
	extractString(user.StandardAttributes, userProfile, stdattrs.Picture)
	extractString(user.StandardAttributes, userProfile, stdattrs.Profile)
	extractString(user.StandardAttributes, userProfile, stdattrs.Website)
	extractString(user.StandardAttributes, userProfile, stdattrs.Gender)
	extractString(user.StandardAttributes, userProfile, stdattrs.Birthdate)
	extractString(user.StandardAttributes, userProfile, stdattrs.Zoneinfo)
	extractString(user.StandardAttributes, userProfile, stdattrs.Locale)

	return userProfile
}
