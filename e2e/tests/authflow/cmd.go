package tests

import (
	"fmt"
	"os"
	"os/exec"
)

type End2EndCmd struct {
	AppID    string
	TestCase TestCase
}

func (e *End2EndCmd) CreateConfigSource() error {
	cmd := fmt.Sprintf("../dist/authgear-portal internal e2e create-configsource --app-id %s --config-source-dir ./fixtures/%s", e.AppID, e.TestCase.Project)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) ImportUsers(appID string, jsonPath string) error {
	cmd := fmt.Sprintf("../dist/authgear internal e2e import-users ./tests/%s --app-id %s", jsonPath, e.AppID)
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
