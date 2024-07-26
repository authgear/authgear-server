package phone

type Parser interface {
	ParseInputPhoneNumber(phone string) (e164 string, err error)
	SplitE164(e164 string) (nationalNumber string, countryCallingCode string, err error)
	CheckE164(phone string) error
}

var LegalParser Parser = &legalParser{}
var LegalAndValidParser Parser = &legalAndValidParser{}

// IsNorthAmericaNumber reports whether e164 is a possible +1 number. It does not check validity.
func IsNorthAmericaNumber(e164 string) (bool, error) {
	_, callingCode, err := LegalParser.SplitE164(e164)
	if err != nil {
		return false, err
	}
	return callingCode == "1", nil
}
