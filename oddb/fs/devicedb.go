package fs

import (
	"encoding/json"
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
