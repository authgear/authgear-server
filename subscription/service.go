package subscription

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
)

var timeNow = time.Now

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
	// maximum number of events per second
	const EventCountBits = 16
	const EventCountMask = 1<<EventCountBits - 1

	var (
		// number of events processed, reset per second
		eventCount uint
		prevUnix   = timeNow().Unix()
	)

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

			currUnix := timeNow().Unix()
			if currUnix != prevUnix {
				eventCount = 0
				prevUnix = currUnix
			}
			seqNum := uint64(currUnix)<<EventCountBits | uint64(eventCount)&EventCountMask
			eventCount++

			db := getDB(conn, event.Record)
			s.handleRecordHook(db, event, seqNum)
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

func (s *Service) handleRecordHook(db oddb.Database, e oddb.RecordEvent, seqNum uint64) {
	subscriptions := db.GetMatchingSubscriptions(e.Record)
	device := oddb.Device{}
	for _, subscription := range subscriptions {
		log.Printf("subscription: got a matching sub id = %s", subscription.ID)

		conn := db.Conn()
		if err := conn.GetDevice(subscription.DeviceID, &device); err != nil {
			log.Panicf("subscription: failed to get device with id = %v: %v", subscription.DeviceID, err)
		}

		notice := Notice{seqNum, subscription.ID, e.Event, e.Record}
		if err := s.Notifier.Notify(device, notice); err != nil {
			log.Errorf("subscription: failed to send notice to device id = %s", device.ID)
		}
	}
}
