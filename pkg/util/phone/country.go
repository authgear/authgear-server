package phone

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/nyaruka/phonenumbers"

	"github.com/authgear/authgear-server/pkg/util/territoryutil"
)

type Country struct {
	Alpha2             string
	CountryCallingCode string
}

var AllCountries []Country

var AllAlpha2 []string

var alpha2ToCountry map[string]Country

var JSONSchemaString string

func init() {
	for _, alpha2 := range territoryutil.Alpha2 {
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
