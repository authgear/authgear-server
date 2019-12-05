package password

import (
	"strings"

	"io/ioutil"
	"os"
)

type ReservedNameChecker struct {
	sourceFile string
}

func (c *ReservedNameChecker) isReserved(name string) (bool, error) {
	f, err := os.Open(c.sourceFile)
	if err != nil {
		return false, err
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		return false, err
	}

	lists := strings.Split(string(content), "\n")
	for i := 0; i < len(lists); i++ {
		if lists[i] == name {
			return true, nil
		}
	}

	return false, nil
}
