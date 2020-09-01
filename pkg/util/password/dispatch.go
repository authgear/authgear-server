package password

var latestFormat passwordFormat

var defaultFormat passwordFormat
var supportedFormats map[string]passwordFormat

func init() {
	latestFormat = bcryptSHA512Password{}

	defaultFormat = bcryptPassword{}
	supportedFormats = map[string]passwordFormat{}
	for _, fmt := range []passwordFormat{
		bcryptSHA512Password{},
	} {
		supportedFormats[fmt.ID()] = fmt
	}
}

func resolveFormat(hash []byte) (passwordFormat, error) {
	id, _, err := parsePasswordFormat(hash)
	if err != nil {
		return nil, err
	}

	fmt, ok := supportedFormats[string(id)]
	if ok {
		return fmt, nil
	}
	return defaultFormat, nil
}

func Hash(password []byte) ([]byte, error) {
	if len(password) > MaxLength {
		password = password[:MaxLength]
	}

	return latestFormat.Hash(password)
}

func Compare(password, hash []byte) error {
	if len(password) > MaxLength {
		password = password[:MaxLength]
	}

	fmt, err := resolveFormat(hash)
	if err != nil {
		return err
	}
	return fmt.Compare(password, hash)
}

func TryMigrate(password []byte, hash *[]byte) (migrated bool, err error) {
	if len(password) > MaxLength {
		password = password[:MaxLength]
	}

	fmt, err := resolveFormat(*hash)
	if err != nil {
		return
	}
	if fmt.ID() == latestFormat.ID() {
		return
	}
	newHash, err := latestFormat.Hash(password)
	if err != nil {
		return
	}

	*hash = newHash
	migrated = true
	return
}
