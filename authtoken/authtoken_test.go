package authtoken

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempDir() string {
	dir, err := ioutil.TempDir("", "oddb.auth.test")
	if err != nil {
		panic(err)
	}
	return dir
}

func TestNewToken(t *testing.T) {
	token := New("46709394", time.Time{})

	if token.UserInfoID != "46709394" {
		t.Fatalf("got token.UserInfoID = %v, want 46709394", token)
	}

	if token.AccessToken == "" {
		t.Fatalf("got token = %v, want non-empty AccessToken value", token)
	}

	if token.ExpiredAt.IsZero() {
		t.Fatalf("got token = %v, want non-zero ExpiredAt value", token)
	}
}

func TestNewTokenWithExpiry(t *testing.T) {
	expiredAt := time.Unix(0, 1)

	token := New("46709394", expiredAt)

	if !token.ExpiredAt.Equal(expiredAt) {
		t.Fatalf("got token.ExpiredAt = %v, want %v", token.ExpiredAt, expiredAt)
	}
}

func TestTokenIsExpired(t *testing.T) {
	now := time.Now()
	token := Token{}

	token.ExpiredAt = now.Add(1 * time.Second)
	if token.IsExpired() {
		t.Fatalf("got expired token = %v, now = %v, want it not expired", token, now)
	}

	token.ExpiredAt = now.Add(-1 * time.Second)
	if !token.IsExpired() {
		t.Fatalf("got non-expired token = %v, now = %v, want it expired", token, now)
	}
}

func TestEmptyTokenIsExpired(t *testing.T) {
	token := Token{}
	if !token.IsExpired() {
		t.Fatalf("got non-expired empty token = %v, want it expired", token)
	}
}

func TestFileStorePut(t *testing.T) {
	const savedFileContent = `{"accessToken":"sometoken","expiredAt":"1970-01-01T00:00:01Z","userInfoID":"someuserinfoid"}
`
	token := Token{
		AccessToken: "sometoken",
		ExpiredAt:   time.Unix(1, 0).UTC(),
		UserInfoID:  "someuserinfoid",
	}

	dir := tempDir()
	defer os.RemoveAll(dir)

	store := FileStore(dir)
	if err := store.Put(&token); err != nil {
		t.Fatalf("got err = %v, want nil", err)
	}

	filePath := filepath.Join(dir, "sometoken")
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(fileBytes, []byte(savedFileContent)) {
		t.Fatalf("got file content = %#v, want %#v", string(fileBytes), savedFileContent)
	}
}
