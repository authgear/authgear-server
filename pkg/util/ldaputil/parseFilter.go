package ldaputil

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/go-ldap/ldap/v3"
)

func ParseFilter(filter string, username string) (string, error) {
	trimedFilter := strings.TrimSpace(filter)

	tmpl, err := template.New("search_filter").Parse(trimedFilter)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// make a string array of phone, username, email
	err = tmpl.Execute(&buf, map[string]string{"Username": ldap.EscapeFilter(username)})

	if err != nil {
		return "", err
	}
	// check if the filter is correct
	result := buf.String()
	result = strings.TrimSpace(result)

	_, err = ldap.CompileFilter(result)
	if err != nil {
		return "", err
	}

	return result, nil
}
