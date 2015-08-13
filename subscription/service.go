package subscription

import (
	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/push"
)

// Service is responsible to send push notification to device whenever
// a record has been modified in db.
type Service struct {
	ConnOpener         func() (oddb.Conn, error)
	NotificationSender push.Sender
	recordEventChan    chan oddb.RecordEvent
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
			s.handleRecordHook(db, event.Record)
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

func (s *Service) handleRecordHook(db oddb.Database, record *oddb.Record) {
	subscriptions := db.GetMatchingSubscriptions(record)

	device := oddb.Device{}
	for _, subscription := range subscriptions {
		log.Printf("Got a matching subscription:\n%#v\n", subscription)

		conn := db.Conn()
		if err := conn.GetDevice(subscription.DeviceID, &device); err != nil {
			log.Panicf("Failed to get device with id = %v: %v", subscription.DeviceID, err)
		}

		customMap := map[string]interface{}{
			"aps": map[string]interface{}{
				"content_available": 1,
			},
			"_ourd": map[string]interface{}{
				"subscription-id": subscription.ID,
			},
		}

		log.Infof("Sending notification to device token = %s", device.Token)
		err := s.NotificationSender.Send(
			push.MapMapper(customMap),
			device.Token,
		)
		if err != nil {
			log.Printf("Failed to send notification: %v\n", err)
		} else {
			log.Infof("Sent notification to device token = %s", device.Token)
		}
	}
}
