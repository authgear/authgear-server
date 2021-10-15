package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
)

type SettingsProfileViewModel struct {
	Name                 string
	GivenName            string
	FamilyName           string
	MiddleName           string
	Nickname             string
	Picture              string
	Profile              string
	Website              string
	Email                string
	PhoneNumber          string
	PreferredUsername    string
	Gender               string
	Birthdate            string
	Zoneinfo             string
	Locale               string
	AddressStreetAddress string
	AddressLocality      string
	AddressRegion        string
	AddressPostalCode    string
	AddressCountry       string
}

type SettingsProfileUserService interface {
	Get(userID string) (*model.User, error)
}

type SettingsProfileViewModeler struct {
	Users SettingsProfileUserService
}

func (m *SettingsProfileViewModeler) ViewModel(userID string) (*SettingsProfileViewModel, error) {
	user, err := m.Users.Get(userID)
	if err != nil {
		return nil, err
	}

	stdAttrs := user.StandardAttributes

	str := func(key string) string {
		value, _ := stdAttrs[key].(string)
		return value
	}

	addressStr := func(key string) string {
		address, ok := stdAttrs[stdattrs.Address].(map[string]interface{})
		if !ok {
			return ""
		}

		value, _ := address[key].(string)
		return value
	}

	viewModel := &SettingsProfileViewModel{
		Name:                 str(stdattrs.Name),
		GivenName:            str(stdattrs.GivenName),
		FamilyName:           str(stdattrs.FamilyName),
		MiddleName:           str(stdattrs.MiddleName),
		Nickname:             str(stdattrs.Nickname),
		Picture:              str(stdattrs.Picture),
		Profile:              str(stdattrs.Profile),
		Website:              str(stdattrs.Website),
		Email:                str(stdattrs.Email),
		PhoneNumber:          str(stdattrs.PhoneNumber),
		PreferredUsername:    str(stdattrs.PreferredUsername),
		Gender:               str(stdattrs.Gender),
		Birthdate:            str(stdattrs.Birthdate),
		Zoneinfo:             str(stdattrs.Zoneinfo),
		Locale:               str(stdattrs.Locale),
		AddressStreetAddress: addressStr(stdattrs.StreetAddress),
		AddressLocality:      addressStr(stdattrs.Locality),
		AddressRegion:        addressStr(stdattrs.Region),
		AddressPostalCode:    addressStr(stdattrs.PostalCode),
		AddressCountry:       addressStr(stdattrs.Country),
	}

	return viewModel, nil
}
