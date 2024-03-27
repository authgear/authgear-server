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
	var configSourcePath = path.Join(path.Dir(e.TestCase.Path), e.TestCase.AuthgearYAMLSource.Extend)
	cmd := fmt.Sprintf("../dist/authgear-portal internal e2e create-configsource --app-id %s --config-source ./tests/authflow/%s --config-override \"%s\"", e.AppID, configSourcePath, e.TestCase.AuthgearYAMLSource.Override)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) ImportUsers(jsonPath string) error {
	var userImportPath = path.Join(path.Dir(e.TestCase.Path), jsonPath)
	cmd := fmt.Sprintf("../dist/authgear internal e2e import-users ./tests/authflow/%s --app-id %s", userImportPath, e.AppID)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) CreateDefaultDomain() error {
	cmd := fmt.Sprintf("../dist/authgear-portal internal e2e create-default-domain --app-id %s", e.AppID)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) execCmd(cmd string) error {
	execCmd := exec.Command("sh", "-c", cmd)
	fmt.Println(execCmd.String())
	execCmd.Dir = "../../"
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
