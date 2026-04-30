package testrunner

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	texttemplate "text/template"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type End2EndCmd struct {
	AppID    string
	Client   *e2eclient.Client
	TestCase TestCase
	Test     testing.TB
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
	AppID    string
	TestCase *TestCase
	Test     testing.TB
}

func NewEnd2EndCmd(options NewEnd2EndCmdOptions) (*End2EndCmd, error) {
	appID := options.AppID
	if appID == "" {
		appID = generateAppID()
	}
	e := &End2EndCmd{
		AppID:    appID,
		TestCase: *options.TestCase,
		Test:     options.Test,
	}

	extraFilesDirectory := ""
	if e.TestCase.ExtraFilesDirectory != "" {
		extraFilesDirectory = e.resolvePath(e.TestCase.ExtraFilesDirectory)
	}

	configOverride, err := renderConfigOverrideTemplate(e, e.TestCase.AuthgearYAMLSource.Override)
	if err != nil {
		return nil, err
	}
	featuresOverride, err := renderConfigOverrideTemplate(e, e.TestCase.AuthgearFeaturesYAMLSource.Override)
	if err != nil {
		return nil, err
	}

	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"create-configsource",
		"--app-id", e.AppID,
		"--config-source", e.resolvePath(e.TestCase.AuthgearYAMLSource.Extend),
		"--config-override", configOverride,
		"--features-override", featuresOverride,
		"--config-source-extra-files-directory", extraFilesDirectory,
	); err != nil {
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
	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"import-users",
		e.resolvePath(jsonPath),
		"--app-id", e.AppID,
	); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) ExecuteSQLInsertUpdateFile(sqlPath string) error {
	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"exec-sql-insert-update",
		"--app-id", e.AppID,
		"--custom-sql", e.resolvePath(sqlPath),
	); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) ExecuteCreateSession(hook *BeforeHookCreateSession) error {
	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"create-session",
		"--app-id", e.AppID,
		"--session-type", hook.SessionType,
		"--session-id", hook.SessionID,
		"--token", hook.Token,
		"--select-user-id-sql", hook.SelectUserIDSQL,
	); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) ExecuteCreateChallenge(hook *BeforeHookCreateChallenge) error {
	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"create-challenge",
		"--app-id", e.AppID,
		"--purpose", string(hook.Purpose),
		"--token", hook.Token,
	); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) QuerySQLSelectRaw(rawSQL string) (jsonArrString string, err error) {
	return e.execCmdArgs(
		"./dist/e2e",
		"query-sql-select",
		"--app-id", e.AppID,
		"--raw-sql", rawSQL,
	)
}

func (e *End2EndCmd) ExecuteSQLInsertUpdateAuditFile(sqlPath string) error {
	if _, err := e.execCmdArgs(
		"./dist/e2e",
		"exec-sql-insert-update-audit",
		"--app-id", e.AppID,
		"--custom-sql", e.resolvePath(sqlPath),
	); err != nil {
		return err
	}
	return nil
}

func (e *End2EndCmd) QuerySQLSelectAuditRaw(rawSQL string) (jsonArrString string, err error) {
	return e.execCmdArgs(
		"./dist/e2e",
		"query-sql-select-audit",
		"--app-id", e.AppID,
		"--raw-sql", rawSQL,
	)
}

func (e *End2EndCmd) GetLinkOTPCodeByClaim(claim string, value string) (string, error) {
	return e.execCmdArgs(
		"./dist/e2e",
		"link-otp-code",
		claim,
		value,
		"--app-id", e.AppID,
	)
}

func (e *End2EndCmd) GenerateIDToken(userID string) (string, error) {
	return e.execCmdArgs(
		"./dist/e2e",
		"generate-id-token",
		userID,
		"--app-id", e.AppID,
	)
}

func (e *End2EndCmd) resolvePath(p string) string {
	if path.IsAbs(p) {
		return p
	}
	return path.Join("./tests", path.Dir(e.TestCase.Path), p)
}

func (e *End2EndCmd) QuerySMTPLog(subject string, recipient string) ([]interface{}, error) {
	file, err := os.Open(filepath.Join("../../logs", "e2e-smtp.log"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rows []interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Subject:["+subject) && strings.Contains(line, "To:["+recipient) {
			rows = append(rows, map[string]interface{}{
				"subject":   subject,
				"recipient": recipient,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

func (e *End2EndCmd) QueryHookServer(path string) ([]interface{}, error) {
	resp, err := http.Get("http://127.0.0.1:2626/" + strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Requests []interface{} `json:"requests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Requests, nil
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

func (e *End2EndCmd) execCmdArgs(args ...string) (string, error) {
	var errb bytes.Buffer
	execCmd := exec.Command(args[0], args[1:]...)
	execCmd.Stderr = &errb
	execCmd.Dir = "../../"
	output, err := execCmd.Output()
	if err != nil {
		e.Test.Errorf("failed to execute command %s: %v\n%s", strings.Join(quoteArgs(args), " "), err, errb.String())
		return "", err
	}

	return string(output), nil
}

func quoteArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		out = append(out, strconv.Quote(arg))
	}
	return out
}

func renderConfigOverrideTemplate(cmd *End2EndCmd, content string) (string, error) {
	tmpl := texttemplate.New("").Delims("[[", "]]")
	tmpl.Funcs(makeTemplateFuncMap(cmd))
	if _, err := tmpl.Parse(content); err != nil {
		return "", err
	}

	data := map[string]any{
		"AppID": cmd.AppID,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
