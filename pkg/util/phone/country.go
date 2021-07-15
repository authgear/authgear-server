package phone

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/nyaruka/phonenumbers"
)

type Country struct {
	Alpha2             string
	CountryCallingCode string
}

var AllCountries []Country

var AllAlpha2 []string

var alpha2ToCountry map[string]Country

var JSONSchemaString string

// supportedAlpha2 is a list of supported alpha2 country codes.
// Ideally, we should call phonenumbers.GetSupportedRegions() to get this list.
// But the portal dependency i18n-iso-countries does not support "AC" and "TA"
// So we hard code the result of phonenumbers.GetSupportedRegions() with "AC" and "TA" excluded.
var supportedAlpha2 []string = []string{
	"AD",
	"AE",
	"AF",
	"AG",
	"AI",
	"AL",
	"AM",
	"AO",
	"AR",
	"AS",
	"AT",
	"AU",
	"AW",
	"AX",
	"AZ",
	"BA",
	"BB",
	"BD",
	"BE",
	"BF",
	"BG",
	"BH",
	"BI",
	"BJ",
	"BL",
	"BM",
	"BN",
	"BO",
	"BQ",
	"BR",
	"BS",
	"BT",
	"BW",
	"BY",
	"BZ",
	"CA",
	"CC",
	"CD",
	"CF",
	"CG",
	"CH",
	"CI",
	"CK",
	"CL",
	"CM",
	"CN",
	"CO",
	"CR",
	"CU",
	"CV",
	"CW",
	"CX",
	"CY",
	"CZ",
	"DE",
	"DJ",
	"DK",
	"DM",
	"DO",
	"DZ",
	"EC",
	"EE",
	"EG",
	"EH",
	"ER",
	"ES",
	"ET",
	"FI",
	"FJ",
	"FK",
	"FM",
	"FO",
	"FR",
	"GA",
	"GB",
	"GD",
	"GE",
	"GF",
	"GG",
	"GH",
	"GI",
	"GL",
	"GM",
	"GN",
	"GP",
	"GQ",
	"GR",
	"GT",
	"GU",
	"GW",
	"GY",
	"HK",
	"HN",
	"HR",
	"HT",
	"HU",
	"ID",
	"IE",
	"IL",
	"IM",
	"IN",
	"IO",
	"IQ",
	"IR",
	"IS",
	"IT",
	"JE",
	"JM",
	"JO",
	"JP",
	"KE",
	"KG",
	"KH",
	"KI",
	"KM",
	"KN",
	"KP",
	"KR",
	"KW",
	"KY",
	"KZ",
	"LA",
	"LB",
	"LC",
	"LI",
	"LK",
	"LR",
	"LS",
	"LT",
	"LU",
	"LV",
	"LY",
	"MA",
	"MC",
	"MD",
	"ME",
	"MF",
	"MG",
	"MH",
	"MK",
	"ML",
	"MM",
	"MN",
	"MO",
	"MP",
	"MQ",
	"MR",
	"MS",
	"MT",
	"MU",
	"MV",
	"MW",
	"MX",
	"MY",
	"MZ",
	"NA",
	"NC",
	"NE",
	"NF",
	"NG",
	"NI",
	"NL",
	"NO",
	"NP",
	"NR",
	"NU",
	"NZ",
	"OM",
	"PA",
	"PE",
	"PF",
	"PG",
	"PH",
	"PK",
	"PL",
	"PM",
	"PR",
	"PS",
	"PT",
	"PW",
	"PY",
	"QA",
	"RE",
	"RO",
	"RS",
	"RU",
	"RW",
	"SA",
	"SB",
	"SC",
	"SD",
	"SE",
	"SG",
	"SH",
	"SI",
	"SJ",
	"SK",
	"SL",
	"SM",
	"SN",
	"SO",
	"SR",
	"SS",
	"ST",
	"SV",
	"SX",
	"SY",
	"SZ",
	"TC",
	"TD",
	"TG",
	"TH",
	"TJ",
	"TK",
	"TL",
	"TM",
	"TN",
	"TO",
	"TR",
	"TT",
	"TV",
	"TW",
	"TZ",
	"UA",
	"UG",
	"US",
	"UY",
	"UZ",
	"VA",
	"VC",
	"VE",
	"VG",
	"VI",
	"VN",
	"VU",
	"WF",
	"WS",
	"XK",
	"YE",
	"YT",
	"ZA",
	"ZM",
	"ZW",
}

func init() {
	for _, alpha2 := range supportedAlpha2 {
		i := phonenumbers.GetCountryCodeForRegion(alpha2)
		ccc := strconv.Itoa(i)
		country := Country{
			Alpha2:             alpha2,
			CountryCallingCode: ccc,
		}
		AllCountries = append(AllCountries, country)
	}

	sort.Slice(AllCountries, func(i, j int) bool {
		c1 := AllCountries[i]
		c2 := AllCountries[j]
		return c1.Alpha2 < c2.Alpha2
	})

	AllAlpha2 = make([]string, len(AllCountries))
	for i, c := range AllCountries {
		AllAlpha2[i] = c.Alpha2
	}

	alpha2ToCountry = make(map[string]Country)
	for _, c := range AllCountries {
		alpha2ToCountry[c.Alpha2] = c
	}

	jsonSchema := map[string]interface{}{
		"type": "string",
		"enum": AllAlpha2,
	}

	b, err := json.Marshal(jsonSchema)
	if err != nil {
		panic(err)
	}
	JSONSchemaString = string(b)
}

func GetCountryByAlpha2(alpha2 string) (c Country, ok bool) {
	c, ok = alpha2ToCountry[alpha2]
	return c, ok
}
