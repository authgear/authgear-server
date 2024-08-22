package testrunner

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path"
	"testing"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type End2EndCmd struct {
	AppID    string
	Client   *e2eclient.Client
	TestCase TestCase
	Test     *testing.T
}

func generateAppID() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}

type NewEnd2EndCmdOptions struct {
	TestCase *TestCase
	Test     *testing.T
}

func NewEnd2EndCmd(options NewEnd2EndCmdOptions) (*End2EndCmd, error) {
	e := &End2EndCmd{
		AppID:    generateAppID(),
		TestCase: *options.TestCase,
		Test:     options.Test,
	}

	cmd := fmt.Sprintf(
		"./dist/e2e create-configsource --app-id %s --config-source %s --config-override \"%s\"",
		e.AppID,
		e.resolvePath(e.TestCase.AuthgearYAMLSource.Extend),
		e.TestCase.AuthgearYAMLSource.Override,
	)
	if _, err := e.execCmd(cmd); err != nil {
		return nil, err
	}

	e.Client = e2eclient.NewClient(
		context.Background(),
		"127.0.0.1:4000",
		"127.0.0.1:4002",
		httputil.HTTPHost(fmt.Sprintf("%s.authgeare2e.localhost:4000", e.AppID)),
	)

	return e, nil
}

func (e *End2EndCmd) ImportUsers(jsonPath string) error {
	cmd := fmt.Sprintf(
		"./dist/e2e import-users %s --app-id %s",
		e.resolvePath(jsonPath),
		e.AppID,
	)
	if _, err := e.execCmd(cmd); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) ExecuteSQLInsertUpdateFile(sqlPath string) error {
	cmd := fmt.Sprintf(
		"./dist/e2e exec-sql-insert-update --app-id %s --custom-sql \"%s\"",
		e.AppID,
		e.resolvePath(sqlPath),
	)
	if _, err := e.execCmd(cmd); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) QuerySQLSelectRaw(rawSQL string) (jsonArrString string, err error) {
	cmd := fmt.Sprintf(
		"./dist/e2e query-sql-select --app-id %s --raw-sql \"%s\"",
		e.AppID,
		rawSQL,
	)

	return e.execCmd(cmd)
}

func (e *End2EndCmd) GetLinkOTPCodeByClaim(claim string, value string) (string, error) {
	cmd := fmt.Sprintf(
		"./dist/e2e link-otp-code %s %s --app-id %s",
		claim,
		value,
		e.AppID,
	)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) GenerateIDToken(userID string) (string, error) {
	cmd := fmt.Sprintf(
		"./dist/e2e generate-id-token %s --app-id %s",
		userID,
		e.AppID,
	)
	return e.execCmd(cmd)
}

func (e *End2EndCmd) resolvePath(p string) string {
	return path.Join("./tests", path.Dir(e.TestCase.Path), p)
}

func (e *End2EndCmd) execCmd(cmd string) (string, error) {
	var errb bytes.Buffer
	execCmd := exec.Command("sh", "-c", cmd)
	execCmd.Stderr = &errb
	execCmd.Dir = "../../"
	output, err := execCmd.Output()
	if err != nil {
		e.Test.Errorf("failed to execute command %s: %v\n%s", cmd, err, errb.String())
		return "", err
	}

	return string(output), nil
}
