package command

import ()

type MigrationFlags []string

func (m *MigrationFlags) String() string {
	return ""
}

func (m *MigrationFlags) Set(value string) error {
	*m = append(*m, value)
	return nil
}
