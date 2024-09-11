package copyutil

import "github.com/mitchellh/copystructure"

// Clone is a wrapper on copystructure.Copy for our customization and testing
func Clone(v interface{}) (interface{}, error) {
	return copystructure.Copy(v)
}
