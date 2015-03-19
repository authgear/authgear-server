package subscription

import (
	"log"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/push"
)

// Service is responsible to send push notification to device whenever
// a record has been modified in db.
type Service struct {
	ConnOpener         func() (oddb.Conn, error)
	NotificationSender push.Sender
}

// Init initializes the record change detection at startup time.
func (s *Service) Init() *Service {
	conn, err := s.ConnOpener()
	if err != nil {
		log.Panicf("Failed to obtain connection: %v", err)
	}
	conn.AddDBRecordHook(s.HandleRecordHook)
	return s
}

// HandleRecordHook provides a hook as the entry point for record change
// detection for oddb implementation that has no native support of
// record change listener.
func (s *Service) HandleRecordHook(db oddb.Database, record *oddb.Record, event oddb.RecordHookEvent) {
	switch event {
	case oddb.RecordCreated, oddb.RecordUpdated, oddb.RecordDeleted:
		s.handleRecordHook(db, record)
	default:
		log.Panicf("Unrecgonized event: %v", event)
	}
}

func (s *Service) handleRecordHook(db oddb.Database, record *oddb.Record) {
	subscriptions := db.GetMatchingSubscription(record)
	for _, subscription := range subscriptions {
		log.Printf("Got a matching subscription:\n%#v\n", subscription)
	}
}
