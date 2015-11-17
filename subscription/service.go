package subscription

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skydb"
)

var timeNow = time.Now

// Service is responsible to send push notification to device whenever
// a record has been modified in db.
type Service struct {
	ConnOpener func() (skydb.Conn, error)
	Notifier   Notifier
	stop       chan struct{}
}

// Run listens for Conn record event
func (s *Service) Run() {
	// maximum number of events per second
	const EventCountBits = 28
	const EventCountMask = 1<<EventCountBits - 1

	var (
		// number of events processed, reset per second
		eventCount    uint
		prevUnix      = timeNow().Unix()
		recordEventCh = s.subscribe()
	)
	s.stop = make(chan struct{})
	defer func() { s.stop = nil }()

	for {
		select {
		case event := <-recordEventCh:
			switch event.Event {
			case skydb.RecordCreated, skydb.RecordUpdated, skydb.RecordDeleted:
				conn, err := s.ConnOpener()
				if err != nil {
					log.WithFields(log.Fields{
						"event": event,
						"err":   err,
					}).Errorln("subscription: failed to open skydb.Conn")
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
				log.Panicf("subscription: unrecgonized event: %v", event)
			}
		case <-s.stop:
			log.Infoln("subscription: stopping the service")
			break
		}
	}
}

// Stop stops the running subscription service
func (s *Service) Stop() {
	s.stop <- struct{}{}
}

func (s *Service) subscribe() chan skydb.RecordEvent {
	conn, err := s.ConnOpener()
	if err != nil {
		log.Panicf("subscription: failed to obtain connection: %v", err)
	}

	ch := make(chan skydb.RecordEvent)
	conn.Subscribe(ch)

	return ch
}

func (s *Service) handleRecordHook(db skydb.Database, e skydb.RecordEvent, seqNum uint64) {
	subscriptions := db.GetMatchingSubscriptions(e.Record)
	device := skydb.Device{}
	for _, subscription := range subscriptions {
		log.Printf("subscription: got a matching sub id = %s", subscription.ID)

		conn := db.Conn()
		if err := conn.GetDevice(subscription.DeviceID, &device); err != nil {
			log.Panicf("subscription: failed to get device with id = %v: %v", subscription.DeviceID, err)
		}

		notice := Notice{seqNum, subscription.ID, e.Event, e.Record}
		if err := s.Notifier.Notify(&device, notice); err != nil {
			log.Errorf("subscription: failed to send notice to device id = %s", device.ID)
		}
	}
}

func getDB(conn skydb.Conn, record *skydb.Record) skydb.Database {
	if record.DatabaseID == "" {
		return conn.PublicDB()
	}

	return conn.PrivateDB(record.DatabaseID)
}
