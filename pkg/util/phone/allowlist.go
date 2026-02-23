package phone

// IsPhoneNumberCountryAllowed checks if any of the phone number's possible country codes
// are in the allowed list. Returns true if the number is allowed, false otherwise.
// If allowlist is empty, the number is considered allowed.
func IsPhoneNumberCountryAllowed(parsed *ParsedPhoneNumber, allowlist []string) bool {
	if len(allowlist) == 0 {
		return true
	}

	allowlistMap := make(map[string]bool)
	for _, allow := range allowlist {
		allowlistMap[allow] = true
	}

	// Allow the phone number if any of the possible region codes is in allowlist
	for _, alpha2 := range parsed.Alpha2 {
		if allowlistMap[alpha2] {
			return true
		}
	}

	return false
}
