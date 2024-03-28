package phone

type Parser interface {
	ParseInputPhoneNumber(phone string) (e164 string, err error)
	ParseCountryCallingCodeAndNationalNumber(nationalNumber string, countryCallingCode string) (e164 string, err error)
	SplitE164(e164 string) (nationalNumber string, countryCallingCode string, err error)
	CheckE164(phone string) error
	IsNorthAmericaNumber(e164 string) (bool, error)
}

var LegalParser Parser = &legalParser{}
var LegalAndValidParser Parser = &legalAndValidParser{}
