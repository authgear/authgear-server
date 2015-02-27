package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/oursky/ourd/oddb"
)

func sampleUserInfo() oddb.UserInfo {
	info := oddb.UserInfo{
		ID:    "uniqueid",
		Email: "john.doe@example.com",
	}
	info.SetPassword("password")

	return info
}

func tempDir() string {
	dir, err := ioutil.TempDir("", "oddb.userdb.test")
	if err != nil {
		panic(err)
	}
	return dir
}

func TestCreate(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	info := sampleUserInfo()
	db := newUserDatabase(dir)

	err := db.Create(&info)
	if err != nil {
		t.Fatalf("got err = %v, want UserInfo created", err)
	}

	err = db.Create(&info)
	if err != oddb.ErrUserDuplicated {
		t.Fatalf("got err = %v, want oddb.ErrUserDuplicated", err)
	}
}

func TestGet(t *testing.T) {
	const userInfoString = `{"id":"alreadyexistid","email":"john.doe@example.com"}`
	expectedUserInfo := oddb.UserInfo{
		ID:    "alreadyexistid",
		Email: "john.doe@example.com",
	}

	dir := tempDir()
	defer os.RemoveAll(dir)

	err := ioutil.WriteFile(filepath.Join(dir, "alreadyexistid"), []byte(userInfoString), 0666)
	if err != nil {
		panic(err)
	}

	db := newUserDatabase(dir)
	info := oddb.UserInfo{}
	err = db.Get("alreadyexistid", &info)
	if err != nil {
		t.Fatalf("got err = %v, want nil", err)
	}

	if !reflect.DeepEqual(info, expectedUserInfo) {
		t.Fatalf("got info = %v, want %v", info, expectedUserInfo)
	}
}

func TestGetNotExist(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	db := newUserDatabase(dir)
	if err := db.Get("notexistid", &oddb.UserInfo{}); err != oddb.ErrUserNotFound {
		t.Fatalf("got err = %v, want oddb.ErrUserNotFound", err)
	}
}

func TestUpdate(t *testing.T) {
	const userInfoString = `{"id":"alreadyexistid","email":"john.doe@example.com"}`
	userInfoToUpdate := oddb.UserInfo{
		ID:             "alreadyexistid",
		Email:          "jane.doe@example.com",
		HashedPassword: []byte("password"),
	}
	// NOTE: JSONEncoder writes a newline at the end
	const updatedUserInfo = `{"id":"alreadyexistid","email":"jane.doe@example.com","password":"cGFzc3dvcmQ="}
`

	dir := tempDir()
	defer os.RemoveAll(dir)

	infoPath := filepath.Join(dir, "alreadyexistid")
	err := ioutil.WriteFile(infoPath, []byte(userInfoString), 0666)
	if err != nil {
		panic(err)
	}

	db := newUserDatabase(dir)
	info := userInfoToUpdate
	err = db.Update(&info)

	updatedBytes, err := ioutil.ReadFile(infoPath)
	if err != nil {
		panic(err)
	}

	if string(updatedBytes) != updatedUserInfo {
		t.Fatalf("got %#v, want %#v", string(updatedBytes), updatedUserInfo)
	}
}

func TestUpdateNotExist(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	info := oddb.UserInfo{ID: "notexistid"}

	db := newUserDatabase(dir)
	if err := db.Update(&info); err != oddb.ErrUserNotFound {
		t.Fatalf("got err = %v, want oddb.ErrUserNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	const userInfoID = "alreadyexistid"

	dir := tempDir()
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, userInfoID)
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	db := newUserDatabase(dir)
	err = db.Delete(userInfoID)
	if err != nil {
		t.Fatalf("got err = %v, want nil", err)
	}

	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		t.Fatalf("got err = %v, want ErrNotExist", err)
	}
}

func TestDeleteNotExist(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)

	db := newUserDatabase(dir)
	if err := db.Delete("notexistid"); err != oddb.ErrUserNotFound {
		t.Fatalf("got err = %v, want oddb.ErrUserNotFound", err)
	}
}
