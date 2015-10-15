package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/oursky/skygear/skydb"
)

// userDatabase is a delegate of fileConn to handle UserInfo's
// storage operations (namely CreateUser, GetUser, UpdateUser and
// DeleteUser).
//
// userDatabase reads and writes UserInfo on disk's directory specified
// by userDatabase.Dir.
type userDatabase struct {
	Dir string
}

func newUserDatabase(dir string) userDatabase {
	// hopefully it would get inlined
	return userDatabase{
		Dir: dir,
	}
}

func (db userDatabase) Create(info *skydb.UserInfo) error {
	// write the file iff the file does not exist
	err := writeUserInfo(db.Dir, info, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	return duplicateErrFromPathError(err)
}

func (db userDatabase) Get(id string, info *skydb.UserInfo) error {
	file, err := os.Open(filepath.Join(db.Dir, id))
	err = notfoundErrFromPathError(err)
	if err != nil {
		return err
	}

	jsonDecoder := json.NewDecoder(file)
	return jsonDecoder.Decode(info)
}

func (db userDatabase) Update(info *skydb.UserInfo) error {
	// write the file iff the file existed already
	err := writeUserInfo(db.Dir, info, os.O_WRONLY|os.O_TRUNC)
	return notfoundErrFromPathError(err)
}

func (db userDatabase) Query(emails []string) ([]skydb.UserInfo, error) {
	userinfos := []skydb.UserInfo{}

	err := db.walk(func(userinfo *skydb.UserInfo) {
		for _, needle := range emails {
			if needle == userinfo.Email && needle != "" {
				userinfos = append(userinfos, *userinfo)
			}
		}
	})

	if err != nil {
		panic(err)
	}

	return userinfos, nil
}

func (db userDatabase) Delete(id string) error {
	err := os.Remove(filepath.Join(db.Dir, id))
	return notfoundErrFromPathError(err)
}

type userDatabaseWalkFunc func(userinfo *skydb.UserInfo)

func (db userDatabase) walk(walkerfunc userDatabaseWalkFunc) error {
	fileinfos, err := ioutil.ReadDir(db.Dir)
	if err != nil {
		return err
	}

	userinfo := skydb.UserInfo{}
	for _, fileinfo := range fileinfos {
		if !fileinfo.IsDir() && fileinfo.Name()[0] != '.' {
			if err := db.Get(fileinfo.Name(), &userinfo); err != nil {
				panic(err)
			}

			walkerfunc(&userinfo)
		}
	}

	return nil
}

func writeUserInfo(dir string, info *skydb.UserInfo, flag int) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// NOTE: 0666 is the default permission used in os.Create
	file, err := os.OpenFile(filepath.Join(dir, info.ID), flag, 0666)
	if err != nil {
		return err
	}

	jsonEncoder := json.NewEncoder(file)
	return jsonEncoder.Encode(info)
}

func duplicateErrFromPathError(err error) error {
	if os.IsExist(err) {
		return skydb.ErrUserDuplicated
	}

	return err
}

func notfoundErrFromPathError(err error) error {
	if os.IsNotExist(err) {
		return skydb.ErrUserNotFound
	}

	return err
}
