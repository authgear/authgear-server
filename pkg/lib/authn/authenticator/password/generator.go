package password

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
}

type RandSource interface {
}

type CryptoRandSource struct{}

func (g *Generator) Generate() (string, error) {
	return "Generated", nil
}
