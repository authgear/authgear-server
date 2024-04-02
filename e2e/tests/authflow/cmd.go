package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

type End2EndCmd struct {
	AppID    string
	TestCase TestCase
}

func (e *End2EndCmd) CreateConfigSource() error {
	cmd := fmt.Sprintf(
		"./dist/authgear-e2e create-configsource --app-id %s --config-source %s --config-override \"%s\"",
		e.AppID,
		e.resolvePath(e.TestCase.AuthgearYAMLSource.Extend),
		e.TestCase.AuthgearYAMLSource.Override,
	)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) ImportUsers(jsonPath string) error {
	cmd := fmt.Sprintf(
		"./dist/authgear-e2e import-users %s --app-id %s",
		e.resolvePath(jsonPath),
		e.AppID,
	)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) ExecuteCustomSQL(sqlPath string) error {
	cmd := fmt.Sprintf(
		"./dist/authgear-e2e exec-sql --app-id %s --custom-sql \"%s\"",
		e.AppID,
		e.resolvePath(sqlPath),
	)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) resolvePath(p string) string {
	return path.Join("./tests/authflow/", path.Dir(e.TestCase.Path), p)
}

func (e *End2EndCmd) execCmd(cmd string) error {
	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Dir = "../../"
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
