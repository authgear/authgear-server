// Implement Crockford's base32 encoding
// ref: https://www.crockford.com/base32.html
package base32

// Alphabet of base32 encoding
const Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// Separator permitted in base32-encoded string, it will be removed in normalization
const Separator = '-'

// Input string contains invalid base32 character
type InvalidBase32Character rune

func (err InvalidBase32Character) Error() string {
	return "invalid base32 character: " + string(err)
}

var normalizationMap map[rune]rune

func init() {
	const (
		sourceAlphabet     = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
		normalizedAlphabet = "0123456789ABCDEFGH1JK1MN0PQRSTUVWXYZABCDEFGH1JK1MN0PQRSTUVWXYZ"
	)
	sourceRunes := []rune(sourceAlphabet)
	normalizedRunes := []rune(normalizedAlphabet)
	normalizationMap = map[rune]rune{}
	for i := range sourceRunes {
		normalizationMap[sourceRunes[i]] = normalizedRunes[i]
	}
}

// Normalize base32-encoded string to alphabet
func Normalize(value string) (normalizedValue string, err error) {
	normalizedDigits := make([]rune, len(value))
	i := 0
	for _, digit := range value {
		if digit == Separator {
			continue
		}
		normalizedDigit, isValid := normalizationMap[digit]
		if !isValid {
			err = InvalidBase32Character(digit)
			return
		}
		normalizedDigits[i] = normalizedDigit
		i++
	}
	normalizedValue = string(normalizedDigits[:i])
	return
}
