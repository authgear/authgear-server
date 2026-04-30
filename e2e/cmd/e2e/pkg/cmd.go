package e2e

import (
	"fmt"
	"os"
	"os/exec"
)

func CreatePortalConfigSource(dbURL string, dbSchema string, resourceDir string, upsert bool) error {
	upsertFlag := ""
	if upsert {
		upsertFlag = "--upsert"
	}
	cmd := fmt.Sprintf(
		"../dist/authgear-portal internal configsource create %s --database-url=\"%s\" --database-schema=\"%s\" %s",
		resourceDir,
		dbURL,
		dbSchema,
		upsertFlag,
	)
	return ExecCmd(cmd)
}

func CreatePortalDefaultDomain(dbURL string, dbSchema string, defaultDomainSuffix string, appID string) error {
	cmd := fmt.Sprintf(
		"../dist/authgear-portal internal domain create-default %s --database-url=\"%s\" --database-schema=\"%s\" --default-domain-suffix=\"%s\"",
		appID,
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
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
