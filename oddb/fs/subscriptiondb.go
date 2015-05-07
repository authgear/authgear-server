package fs

import (
	"encoding/json"
	"io/ioutil"
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
	file, err := os.Open(filepath.Join(db.Dir, key))
	if err != nil {
		if os.IsNotExist(err) {
			return oddb.ErrSubscriptionNotFound
		}
		return err
	}

	jsonDecoder := json.NewDecoder(file)
	return jsonDecoder.Decode(s)
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

func (db subscriptionDB) GetMatchingSubscriptions(record *oddb.Record) []oddb.Subscription {
	subscriptions := []oddb.Subscription{}

	err := db.walk(func(subscription *oddb.Subscription) {
		if matchSubscriptionWithRecord(subscription, record) {
			subscriptions = append(subscriptions, *subscription)
		}
	})

	if err != nil {
		panic(err)
	}

	return subscriptions
}

type walkFunc func(subscription *oddb.Subscription)

func (db subscriptionDB) walk(walkerfunc walkFunc) error {
	fileinfos, err := ioutil.ReadDir(db.Dir)
	if err != nil {
		return err
	}

	subscription := oddb.Subscription{}
	for _, fileinfo := range fileinfos {
		if !fileinfo.IsDir() && fileinfo.Name()[0] != '.' {
			if err := db.Get(fileinfo.Name(), &subscription); err != nil {
				panic(err)
			}

			walkerfunc(&subscription)
		}
	}

	return nil
}

func matchSubscriptionWithRecord(subscription *oddb.Subscription, record *oddb.Record) bool {
	return (*queryMatcher)(&subscription.Query).match(record)
}
