package e2e

import (
	"fmt"
	"os"
	"os/exec"
)

func CreatePortalConfigSource(dbURL string, dbSchema string, resourceDir string) error {
	cmd := fmt.Sprintf(
		"../dist/authgear-portal internal configsource create %s --database-url=\"%s\" --database-schema=\"%s\"",
		resourceDir,
		dbURL,
		dbSchema,
	)
	return ExecCmd(cmd)
}

func CreatePortalDefaultDomain(dbURL string, dbSchema string, defaultDomainSuffix string) error {
	cmd := fmt.Sprintf(
		"../dist/authgear-portal internal domain create-default --database-url=\"%s\" --database-schema=\"%s\" --default-domain-suffix=\"%s\"",
		dbURL,
		dbSchema,
		defaultDomainSuffix,
	)
	return ExecCmd(cmd)
}

func ExecCmd(cmd string) error {
	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Dir = "."
	execCmd.Stdout = os.Stdout
	return execCmd.Run()
}
