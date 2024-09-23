package user_import

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func generateUsers(t testing.TB, name string, n int, m int) string {
	var records []any
	for range n {
		u := rand.N(m)

		hash, err := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("password%d", u)), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}

		records = append(records, map[string]any{
			"preferred_username": fmt.Sprintf("user_%d", u),
			"email":              fmt.Sprintf("user_%d@example.com", u),
			"password": map[string]any{
				"type":          "bcrypt",
				"password_hash": string(hash),
			},
		})
	}

	data := map[string]any{
		"identifier": "preferred_username",
		"records":    records,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(t.TempDir(), name)
	err = os.WriteFile(file, jsonData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	return file
}

func BenchmarkUserImport(b *testing.B) {
	var userFiles []string
	for i := range b.N {
		userFiles = append(userFiles, generateUsers(b, fmt.Sprintf("users%d.json", i), 100, 1000))
	}

	cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
		TestCase: &testrunner.TestCase{
			Path: "user_import/user_import_test.go",
		},
		Test: b,
	})
	cmd.ExtraEnv = append(cmd.ExtraEnv, "DATABASE_CONFIG_USE_PREPARED_STATEMENTS=1")

	b.ResetTimer()

	for i := range b.N {
		if err != nil {
			b.Fatal(err)
		}

		if err := cmd.ImportUsers(userFiles[i]); err != nil {
			b.Fatal(err)
		}
	}
}
