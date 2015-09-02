package subscription

import (
	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
)

// Service is responsible to send push notification to device whenever
// a record has been modified in db.
type Service struct {
	ConnOpener      func() (oddb.Conn, error)
	Notifier        Notifier
	recordEventChan chan oddb.RecordEvent
}

// Init initializes the record change detection at startup time.
func (s *Service) Init() *Service {
	conn, err := s.ConnOpener()
	if err != nil {
		log.Panicf("Failed to obtain connection: %v", err)
	}

	s.recordEventChan = make(chan oddb.RecordEvent)
	conn.Subscribe(s.recordEventChan)

	return s
}

// Listen listens for Conn record event
func (s *Service) Listen() {
	for {
		event := <-s.recordEventChan
		switch event.Event {
		case oddb.RecordCreated, oddb.RecordUpdated, oddb.RecordDeleted:
			conn, err := s.ConnOpener()
			if err != nil {
				log.WithFields(log.Fields{
					"event": event,
					"err":   err,
				}).Errorln("subscription/service: failed to open conn")
				continue
			}
			db := getDB(conn, event.Record)
			s.handleRecordHook(db, event.Event, event.Record)
		default:
			log.Panicf("Unrecgonized event: %v", event)
		}
	}
}

func getDB(conn oddb.Conn, record *oddb.Record) oddb.Database {
	if record.DatabaseID == "" {
		return conn.PublicDB()
	}

	return conn.PrivateDB(record.DatabaseID)
}

func (s *Service) handleRecordHook(db oddb.Database, event oddb.RecordHookEvent, record *oddb.Record) {
	subscriptions := db.GetMatchingSubscriptions(record)

	device := oddb.Device{}
	for _, subscription := range subscriptions {
		log.Printf("subscription: got a matching sub id = %s", subscription.ID)

		conn := db.Conn()
		if err := conn.GetDevice(subscription.DeviceID, &device); err != nil {
			log.Panicf("subscription: failed to get device with id = %v: %v", subscription.DeviceID, err)
		}

		notice := Notice{subscription.ID, event, record}
		if err := s.Notifier.Notify(device, notice); err != nil {
			log.Errorf("subscription: failed to send notice to device id = %s", device.ID)
		}
	}
}
