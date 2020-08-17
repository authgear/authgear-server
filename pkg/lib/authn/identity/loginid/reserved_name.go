package loginid

import (
	"io/ioutil"
	"os"
	"strings"
)

type ReservedNameChecker struct {
	reservedWords []string
}

func NewReservedNameChecker(sourceFile string) (*ReservedNameChecker, error) {
	f, err := os.Open(sourceFile)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	reservedWords := strings.Split(string(content), "\n")

	return &ReservedNameChecker{
		reservedWords: reservedWords,
	}, nil
}

func (c *ReservedNameChecker) IsReserved(name string) (bool, error) {
	for i := 0; i < len(c.reservedWords); i++ {
		if c.reservedWords[i] == name {
			return true, nil
		}
	}

	return false, nil
}
