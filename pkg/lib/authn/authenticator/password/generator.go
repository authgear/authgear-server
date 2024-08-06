package password

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"strings"
)

// Character list for each category.
const (
	CharListLowercase    = "abcdefghijklmnopqrstuvwxyz"
	CharListUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharListAlphabet     = CharListLowercase + CharListUppercase
	CharListDigit        = "0123456789"
	CharListAlphanumeric = CharListAlphabet + CharListDigit
	// TODO: Define real list of special characters
	CharListSymbol = "!@#$%^&*()-_=+,.?/:;{}[]`~"
)

const (
	DefaultCharList = CharListAlphanumeric
	// Max trials to generate a password that satisfies the checker.
	MaxTrials = 10
	// Default minimum length of a password, overrides min length in the policy if less than it.
	DefaultMinLength = 8
	// When min guessable level is > 0, the minimum length of a password.
	GuessableEnabledMinLength = 32
)

type RandSource interface {
	Pick(list string) (int, error)
	Shuffle(list string) (string, error)
}

type CryptoRandSource struct{}

// Uint32 is used to save memory since our character list is short enough.
func (r *CryptoRandSource) uint32() (uint32, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(b), nil
}

func (r *CryptoRandSource) Pick(list string) (int, error) {
	n := uint32(len(list))

	// Discard the numbers that are greater than the maximum multiple of n
	// to avoid modulo bias.
	discard := uint32(math.MaxUint32 - math.MaxUint32%n)
	v, err := r.uint32()
	if err != nil {
		return 0, err
	}

	for v >= discard {
		v, err = r.uint32()
		if err != nil {
			return 0, err
		}
	}

	return int(v % n), nil
}

func (r *CryptoRandSource) Shuffle(list string) (string, error) {
	var charList = []byte(list)
	for i := len(charList) - 1; i > 0; i-- {
		j, err := r.Pick(string(charList[:i+1]))
		if err != nil {
			return "", err
		}
		charList[i], charList[j] = charList[j], charList[i]
	}
	return string(charList), nil
}

type Generator struct {
	Checker    *Checker
	RandSource RandSource

	PwMinLength         int
	PwUppercaseRequired bool
	PwLowercaseRequired bool
	PwAlphabetRequired  bool
	PwDigitRequired     bool
	PwSymbolRequired    bool
	PwMinGuessableLevel int
	PwExcludedKeywords  []string
}

func (g *Generator) Generate() (string, error) {
	for i := 0; i < MaxTrials; i++ {
		password, err := g.generate()
		if err != nil {
			return "", err
		}

		if err := g.checkPassword(password); err == nil {
			return password, nil
		}
	}

	return "", ErrPasswordGenerateFailed
}

func (g *Generator) generate() (string, error) {
	var charList = g.prepareCharList()

	// Ensure the password is at least the minimum length.
	var minLength = g.PwMinLength
	if minLength < DefaultMinLength {
		minLength = DefaultMinLength
	}

	// Override min length if min guessable level is enabled.
	if g.PwMinGuessableLevel > 0 {
		minLength = GuessableEnabledMinLength
	}

	var password strings.Builder

	if g.PwUppercaseRequired {
		c, err := g.RandSource.Pick(CharListUppercase)
		if err != nil {
			return "", err
		}
		password.WriteByte(CharListUppercase[c])
	}
	if g.PwLowercaseRequired {
		c, err := g.RandSource.Pick(CharListLowercase)
		if err != nil {
			return "", err
		}
		password.WriteByte(CharListLowercase[c])
	}
	if g.PwAlphabetRequired && !g.PwUppercaseRequired && !g.PwLowercaseRequired {
		c, err := g.RandSource.Pick(CharListAlphabet)
		if err != nil {
			return "", err
		}
		password.WriteByte(CharListAlphabet[c])
	}
	if g.PwDigitRequired {
		c, err := g.RandSource.Pick(CharListDigit)
		if err != nil {
			return "", err
		}
		password.WriteByte(CharListDigit[c])
	}
	if g.PwSymbolRequired {
		c, err := g.RandSource.Pick(CharListSymbol)
		if err != nil {
			return "", err
		}
		password.WriteByte(CharListSymbol[c])
	}

	for j := 0; j < minLength; j++ {
		c, err := g.RandSource.Pick(charList)
		if err != nil {
			return "", err
		}
		password.WriteByte(charList[c])
	}

	shuffled, err := g.RandSource.Shuffle(password.String())
	if err != nil {
		return "", err
	}

	return shuffled, nil
}

func createSetFromList(list string) string {
	var set strings.Builder
	setMap := make(map[byte]bool)

	for i := 0; i < len(list); i++ {
		if _, ok := setMap[list[i]]; !ok {
			set.WriteByte(list[i])
			setMap[list[i]] = true
		}
	}

	return set.String()
}

func (g *Generator) prepareCharList() string {
	charList := DefaultCharList

	if g.PwLowercaseRequired {
		charList += CharListLowercase
	}
	if g.PwUppercaseRequired {
		charList += CharListUppercase
	}
	if g.PwAlphabetRequired {
		charList += CharListAlphabet
	}
	if g.PwDigitRequired {
		charList += CharListDigit
	}
	if g.PwSymbolRequired {
		charList += CharListSymbol
	}

	for _, keyword := range g.PwExcludedKeywords {
		charList = strings.ReplaceAll(charList, keyword, "")
	}

	// Deduplication
	charList = createSetFromList(charList)

	return charList
}

func (g *Generator) checkPassword(password string) error {
	return g.Checker.ValidateCurrentPassword(password)
}
