package phone

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ErrNotInE164Format means the given phone number is not in E.164 format.
var ErrNotInE164Format = errors.New("not in E.164 format")

// ErrPhoneNumberInvalid means phone number doesn't pass validation
var ErrPhoneNumberInvalid = errors.New("invalid phone number")

func combineCallingCodeWithNumber(nationalNumber string, callingCode string) string {
	nationalNumber = strings.TrimSpace(nationalNumber)
	callingCode = strings.TrimSpace(callingCode)
	if !strings.HasPrefix(callingCode, "+") {
		callingCode = fmt.Sprintf("+%s", callingCode)
	}
	return fmt.Sprintf("%s%s", callingCode, nationalNumber)
}

// According to https://godoc.org/github.com/nyaruka/phonenumbers#Parse,
// no validation is performed during parsing, so we need to call IsValidNumber to check
func ParseCombinedPhoneNumber(phone string) (e164 string, err error) {
	// check if input contains non-numeric character(s)
	// The nationalNumber part of phone is parsed to uint64,
	// letters in input phone number will be parsed successfully.
	isNumericString, _ := regexp.Match(`^\+[0-9\ \-]*$`, []byte(phone))
	if !isNumericString {
		err = ErrNotInE164Format
		return
	}
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return
	}
	isPhoneValid := phonenumbers.IsValidNumber(num)
	if !isPhoneValid {
		err = ErrPhoneNumberInvalid
		return
	}
	e164 = phonenumbers.Format(num, phonenumbers.E164)
	return
}

// ParseE164ToCallingCodeAndNumber parse E.164 format phone number to national number and calling code.
func ParseE164ToCallingCodeAndNumber(e164 string) (nationalNumber string, callingCode string, err error) {
	err = EnsureE164(e164)
	if err != nil {
		return
	}

	num, err := phonenumbers.Parse(e164, "")
	if err != nil {
		return
	}
	isPhoneValid := phonenumbers.IsValidNumber(num)
	if !isPhoneValid {
		err = ErrPhoneNumberInvalid
		return
	}
	callingCode = strconv.Itoa(int(num.GetCountryCode()))
	nationalNumber = phonenumbers.GetNationalSignificantNumber(num)
	return
}

// Parse to E164 format
func Parse(nationalNumber string, callingCode string) (e164 string, err error) {
	rawInput := combineCallingCodeWithNumber(nationalNumber, callingCode)
	e164, err = ParseCombinedPhoneNumber(rawInput)
	return
}

// EnsureE164 ensures the given phone is in E.164 format.
func EnsureE164(phone string) error {
	formatted, err := ParseCombinedPhoneNumber(phone)
	if err != nil {
		return err
	}
	if formatted != phone {
		return ErrNotInE164Format
	}
	return nil
}

// Mask masks the give phone number.
func Mask(phone string) string {
	var buf strings.Builder
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return ""
	}

	countryCallingCode := int(num.GetCountryCode())
	nationalSignificantNumber := phonenumbers.GetNationalSignificantNumber(num)
	runes := []rune(nationalSignificantNumber)
	length := len(runes)
	halfLength := length / 2

	buf.WriteRune('+')
	buf.WriteString(strconv.Itoa(countryCallingCode))
	for i := 0; i < length; i++ {
		if i < halfLength {
			buf.WriteRune(runes[i])
		} else {
			buf.WriteRune('*')
		}
	}

	return buf.String()
}
