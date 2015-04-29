package pq

import (
	"github.com/oursky/ourd/oddb"
)

func (db *database) GetSubscription(key string, subscription *oddb.Subscription) error { return nil }

func (db *database) SaveSubscription(subscription *oddb.Subscription) error {
	return nil
}

func (db *database) DeleteSubscription(key string) error { return nil }

func (db *database) GetMatchingSubscription(record *oddb.Record) []oddb.Subscription { return nil }
