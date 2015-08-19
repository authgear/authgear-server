package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/oursky/ourd/oddb"
)

type deviceDatabase struct {
	Dir string
}

func newDeviceDatabase(dir string) *deviceDatabase {
	return &deviceDatabase{dir}
}

func (db *deviceDatabase) Get(id string, device *oddb.Device) error {
	file, err := os.Open(filepath.Join(db.Dir, id))
	if err != nil {
		if os.IsNotExist(err) {
			return oddb.ErrDeviceNotFound
		}
		return err
	}

	jsonDecoder := json.NewDecoder(file)
	return jsonDecoder.Decode(device)

}

func (db *deviceDatabase) Save(device *oddb.Device) error {
	if err := os.MkdirAll(db.Dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(db.Dir, device.ID))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(device); err != nil {
		return err
	}

	return nil
}

func (db *deviceDatabase) Delete(id string) error {
	if err := os.Remove(filepath.Join(db.Dir, id)); err != nil {
		if os.IsNotExist(err) {
			return oddb.ErrDeviceNotFound
		}
		return err
	}

	return nil
}

type deviceDatabaseWalkFunc func(deviceinfo *oddb.Device)

func (db deviceDatabase) walk(walkerfunc deviceDatabaseWalkFunc) error {
	fileinfos, err := ioutil.ReadDir(db.Dir)
	if err != nil {
		return err
	}

	deviceinfo := oddb.Device{}
	for _, fileinfo := range fileinfos {
		if !fileinfo.IsDir() && fileinfo.Name()[0] != '.' {
			if err := db.Get(fileinfo.Name(), &deviceinfo); err != nil {
				panic(err)
			}

			walkerfunc(&deviceinfo)
		}
	}

	return nil
}

func (db deviceDatabase) Query(user string) ([]oddb.Device, error) {
	deviceinfos := []oddb.Device{}

	err := db.walk(func(deviceinfo *oddb.Device) {
		if user == deviceinfo.UserInfoID && user != "" {
			deviceinfos = append(deviceinfos, *deviceinfo)
		}
	})

	if err != nil {
		panic(err)
	}

	return deviceinfos, nil
}
