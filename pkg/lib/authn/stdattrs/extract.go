package stdattrs

func extractString(input map[string]interface{}, output T, key string) {
	if value, ok := input[key].(string); ok && value != "" {
		output[key] = value
	}
}

func extractBool(input map[string]interface{}, output T, key string) {
	if value, ok := input[key].(bool); ok {
		output[key] = value
	}
}

func extractAddress(input map[string]interface{}, output T) {
	if inAddr, ok := input[Address].(map[string]interface{}); ok {
		outAddr := make(map[string]interface{})
		extractString(inAddr, T(outAddr), Formatted)
		extractString(inAddr, T(outAddr), StreetAddress)
		extractString(inAddr, T(outAddr), Locality)
		extractString(inAddr, T(outAddr), Region)
		extractString(inAddr, T(outAddr), PostalCode)
		extractString(inAddr, T(outAddr), Country)
		if len(outAddr) > 0 {
			output[Address] = outAddr
		}
	}
}

// Extract extracts OIDC standard claims.
// The output is NOT normalized.
func Extract(claims map[string]interface{}) T {
	out := T{}

	extractString(claims, out, Name)
	extractString(claims, out, GivenName)
	extractString(claims, out, FamilyName)
	extractString(claims, out, MiddleName)
	extractString(claims, out, Nickname)
	extractString(claims, out, PreferredUsername)
	extractString(claims, out, Profile)
	extractString(claims, out, Picture)
	extractString(claims, out, Website)
	extractString(claims, out, Email)
	extractBool(claims, out, EmailVerified)
	extractString(claims, out, Gender)
	extractString(claims, out, Birthdate)
	extractString(claims, out, Zoneinfo)
	extractString(claims, out, Locale)
	extractString(claims, out, PhoneNumber)
	extractBool(claims, out, PhoneNumberVerified)
	extractAddress(claims, out)

	return out
}
