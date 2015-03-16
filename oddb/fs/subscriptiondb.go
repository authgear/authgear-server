package fs

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/oursky/ourd/oddb"
)

type subscriptionDB struct {
	Dir string
}

func newSubscriptionDB(dir string) subscriptionDB {
	return subscriptionDB{dir}
}

func (db subscriptionDB) Get(key string, s *oddb.Subscription) error {
	log.Panicln("GetSubscription not implemented")
	return nil
}

func (db subscriptionDB) Save(s *oddb.Subscription) error {
	if err := os.MkdirAll(db.Dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(db.Dir, s.ID))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(s); err != nil {
		return err
	}

	return nil
}

func (db subscriptionDB) Delete(key string) error {
	log.Panicln("DeleteSubscription not implemented")
	return nil
}
