package oddb

import (
	"bytes"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewUserInfo(t *testing.T) {
	info := NewUserInfo("john.doe@example.com", "secret")
	if info.ID == "" {
		t.Fatalf("got info.ID = %v, want \"\"", info.ID)
	}

	if info.Email != "john.doe@example.com" {
		t.Fatalf("got info.Email = %v, want john.doe@example.com", info.Email)
	}

	if bytes.Equal(info.HashedPassword, nil) {
		t.Fatalf("got info.HashPassword = %v, want non-empty value", info.HashedPassword)
	}
}

func TestSetPassword(t *testing.T) {
	info := UserInfo{}
	info.SetPassword("secret")
	err := bcrypt.CompareHashAndPassword(info.HashedPassword, []byte("secret"))
	if err != nil {
		t.Fatalf("got err = %v, want nil", err)
	}
}

func TestIsSamePassword(t *testing.T) {
	info := UserInfo{}
	info.SetPassword("secret")
	if !info.IsSamePassword("secret") {
		t.Fatalf("got UserInfo.HashedPassword = %v, want a hashed \"secret\"", info.HashedPassword)
	}
}
