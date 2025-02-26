package phone

import (
	"regexp"
	"strconv"

	"github.com/nyaruka/phonenumbers"
)

type ParsedPhoneNumber struct {
	UserInput                         string
	E164                              string
	Alpha2                            []string
	IsPossibleNumber                  bool
	IsValidNumber                     bool
	CountryCallingCodeWithoutPlusSign string
	NationalNumberWithoutFormatting   string
}

func ParsePhoneNumberWithUserInput(userInput string) (*ParsedPhoneNumber, error) {
	isNumericString, _ := regexp.Match(`^ *\+[0-9\ \-]*$`, []byte(userInput))
	if !isNumericString {
		return nil, ErrNotInE164Format
	}

	num, err := phonenumbers.Parse(userInput, "")
	if err != nil {
		return nil, ErrNotInE164Format
	}

	e164 := phonenumbers.Format(num, phonenumbers.E164)
	isPossibleNumber := phonenumbers.IsPossibleNumber(num)
	isValidNumber := phonenumbers.IsValidNumber(num)
	countryCallingCode := int(num.GetCountryCode())
	nationalNumber := phonenumbers.GetNationalSignificantNumber(num)
	var alpha2 []string
	regionCode := phonenumbers.GetRegionCodeForNumber(num)
	if regionCode == "" {
		// It is possible that `GetRegionCodeForNumber` returns empty string if the number is not a valid number
		// In this case we determine the region from country code
		alpha2 = phonenumbers.GetRegionCodesForCountryCode(countryCallingCode)
	} else {
		alpha2 = []string{regionCode}
	}

	return &ParsedPhoneNumber{
		UserInput:                         userInput,
		E164:                              e164,
		Alpha2:                            alpha2,
		IsPossibleNumber:                  isPossibleNumber,
		IsValidNumber:                     isValidNumber,
		CountryCallingCodeWithoutPlusSign: strconv.Itoa(countryCallingCode),
		NationalNumberWithoutFormatting:   nationalNumber,
	}, nil
}

func (n ParsedPhoneNumber) RequireIsPossibleNumber() error {
	if !n.IsPossibleNumber {
		return ErrPhoneNumberInvalid
	}
	return nil
}

func (n ParsedPhoneNumber) RequireIsValidNumber() error {
	if !n.IsValidNumber {
		return ErrPhoneNumberInvalid
	}
	return nil
}

func (n ParsedPhoneNumber) RequireUserInputInE164() error {
	if n.UserInput != n.E164 {
		return ErrNotInE164Format
	}
	return nil
}

func (n ParsedPhoneNumber) IsNorthAmericaNumber() bool {
	return n.CountryCallingCodeWithoutPlusSign == "1"
}

func (n ParsedPhoneNumber) Require_IsPossibleNumber_IsValidNumber_UserInputInE164() error {
	if n.RequireIsPossibleNumber() != nil {
		return n.RequireIsPossibleNumber()
	}
	if n.RequireIsValidNumber() != nil {
		return n.RequireIsValidNumber()
	}
	if n.RequireUserInputInE164() != nil {
		return n.RequireUserInputInE164()
	}
	return nil
}

func Require_IsPossibleNumber_IsValidNumber_UserInputInE164(userInput string) error {
	parsed, err := ParsePhoneNumberWithUserInput(userInput)
	if err != nil {
		return err
	}
	return parsed.Require_IsPossibleNumber_IsValidNumber_UserInputInE164()
}

func Parse_IsPossibleNumber_ReturnE164(userInput string) (e164 string, err error) {
	parsed, err := ParsePhoneNumberWithUserInput(userInput)
	if err != nil {
		return "", err
	}
	if parsed.RequireIsPossibleNumber() != nil {
		return "", parsed.RequireIsPossibleNumber()
	}
	return parsed.E164, nil
}
