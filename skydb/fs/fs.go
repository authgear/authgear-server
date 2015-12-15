package fs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/skydb"
)

const userDBKey = "_user"

const publicDBKey = "_public"
const privateDBKey = "_private"

// the number returned by t.Unix() when t is an empty time
const zeroUnix = -62135596800

var dbHookFuncs []skydb.DBHookFunc

// fileConn implements skydb.Conn interface
type fileConn struct {
	Dir      string
	AppName  string
	userDB   userDatabase
	deviceDB *deviceDatabase
	publicDB skydb.Database
}

// Open returns a new connection to fs implementation
func Open(appName, dir string) (skydb.Conn, error) {
	if appName == "" {
		return nil, errors.New("fs: appName cannot be empty")
	}

	containerPath := filepath.Join(dir, appName)
	userDBPath := filepath.Join(containerPath, userDBKey)
	deviceDBPath := filepath.Join(containerPath, "_device")
	publicDBPath := filepath.Join(containerPath, publicDBKey)

	conn := &fileConn{
		Dir:      containerPath,
		AppName:  appName,
		userDB:   newUserDatabase(userDBPath),
		deviceDB: newDeviceDatabase(deviceDBPath),
	}
	conn.publicDB = newDatabase(conn, publicDBPath, publicDBKey, "")

	return conn, nil
}

func (conn *fileConn) Close() error {
	return nil
}

func (conn *fileConn) CreateUser(info *skydb.UserInfo) error {
	return conn.userDB.Create(info)
}

func (conn *fileConn) GetUser(id string, info *skydb.UserInfo) error {
	return conn.userDB.Get(id, info)
}

func (conn *fileConn) GetUserByUsernameEmail(username string, email string, userinfo *skydb.UserInfo) error {
	panic("not implemented")
}

func (conn *fileConn) GetUserByPrincipalID(principalID string, info *skydb.UserInfo) error {
	panic("not implemented")
}

func (conn *fileConn) UpdateUser(info *skydb.UserInfo) error {
	return conn.userDB.Update(info)
}

func (conn *fileConn) QueryUser(emails []string) ([]skydb.UserInfo, error) {
	return conn.userDB.Query(emails)
}

func (conn *fileConn) DeleteUser(id string) error {
	return conn.userDB.Delete(id)
}

func (conn *fileConn) GetAsset(name string, asset *skydb.Asset) error {
	panic("not implemented")
}

func (conn *fileConn) SaveAsset(assert *skydb.Asset) error {
	panic("not implemented")
}

func (conn *fileConn) QueryRelation(user string, name string, direction string) []skydb.UserInfo {
	panic("not implemented")
}

func (conn *fileConn) QueryRelationCount(user string, name string, direction string) (uint64, error) {
	panic("not implemented")
}

func (conn *fileConn) AddRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

func (conn *fileConn) RemoveRelation(user string, name string, targetUser string) error {
	panic("not implemented")
}

func (conn *fileConn) GetDevice(id string, device *skydb.Device) error {
	return conn.deviceDB.Get(id, device)
}

func (conn *fileConn) QueryDevicesByUser(user string) ([]skydb.Device, error) {
	return conn.deviceDB.Query(user)
}

func (conn *fileConn) SaveDevice(device *skydb.Device) error {
	return conn.deviceDB.Save(device)
}

func (conn *fileConn) DeleteDevice(id string) error {
	return conn.deviceDB.Delete(id)
}

func (conn *fileConn) DeleteDeviceByToken(token string, t time.Time) error {
	panic("not implemented")
}

func (conn *fileConn) DeleteEmptyDevicesByTime(t time.Time) error {
	panic("not implemented")
}

func (conn *fileConn) PublicDB() skydb.Database {
	return conn.publicDB
}

func (conn *fileConn) PrivateDB(userKey string) skydb.Database {
	dbPath := filepath.Join(conn.Dir, userKey)
	return newDatabase(conn, dbPath, privateDBKey, userKey)
}

func (conn *fileConn) Subscribe(recordEventChan chan skydb.RecordEvent) error {
	return nil
}

type fileDatabase struct {
	conn      *fileConn
	Dir       string
	Key       string
	UserID    string
	subscriDB subscriptionDB
}

func newDatabase(conn *fileConn, dir string, key string, userID string) *fileDatabase {
	return &fileDatabase{
		conn:      conn,
		Dir:       dir,
		Key:       key,
		UserID:    userID,
		subscriDB: newSubscriptionDB(filepath.Join(dir, "_subscription")),
	}
}

// convenient method to execute hooks if err is nil
func (db fileDatabase) executeHook(record *skydb.Record, event skydb.RecordHookEvent, err error) error {
	if err != nil {
		return err
	}

	for _, hookFunc := range dbHookFuncs {
		go hookFunc(db, record, event)
	}

	return nil
}

func (db fileDatabase) Conn() skydb.Conn {
	return db.conn
}

func (db fileDatabase) ID() string {
	return db.Key
}

func (db fileDatabase) Get(id skydb.RecordID, record *skydb.Record) error {
	file, err := os.Open(db.recordPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return skydb.ErrRecordNotFound
		}
		return err
	}

	if err := json.NewDecoder(file).Decode(record); err != nil {
		return err
	}

	if record.CreatedAt.Unix() == zeroUnix {
		record.CreatedAt = time.Time{}
	}
	if record.UpdatedAt.Unix() == zeroUnix {
		record.UpdatedAt = time.Time{}
	}
	record.DatabaseID = db.UserID
	return nil
}

func (db fileDatabase) GetByIDs(ids []skydb.RecordID) (*skydb.Rows, error) {
	panic("not implemented")
}

func (db fileDatabase) Save(record *skydb.Record) error {
	filePath := db.recordPath(record.ID)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	event := recordEventByPath(filePath)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(file).Encode(record); err != nil {
		return err
	}

	record.DatabaseID = db.UserID

	return db.executeHook(record, event, err)
}

func (db fileDatabase) Delete(id skydb.RecordID) error {
	record := skydb.Record{}
	err := os.Remove(db.recordPath(id))
	if os.IsNotExist(err) {
		err = skydb.ErrRecordNotFound
	}

	return db.executeHook(&record, skydb.RecordDeleted, err)
}

type recordSorter struct {
	records []skydb.Record
	by      func(r1, r2 *skydb.Record) bool
}

func (s *recordSorter) Len() int {
	return len(s.records)
}

func (s *recordSorter) Swap(i, j int) {
	s.records[i], s.records[j] = s.records[j], s.records[i]
}

func (s *recordSorter) Less(i, j int) bool {
	less := s.by(&s.records[i], &s.records[j])
	return less
}

func (s *recordSorter) Sort() {
	sort.Sort(s)
}

func newRecordSorter(records []skydb.Record, sortinfo skydb.Sort) *recordSorter {
	var by func(r1, r2 *skydb.Record) bool

	field := sortinfo.KeyPath

	switch sortinfo.Order {
	default:
		by = func(r1, r2 *skydb.Record) bool {
			return reflectLess(r1.Get(field), r2.Get(field))
		}
	case skydb.Desc:
		by = func(r1, r2 *skydb.Record) bool {
			return !reflectLess(r1.Get(field), r2.Get(field))
		}
	}

	return &recordSorter{
		records: records,
		by:      by,
	}
}

// reflectLess determines whether i1 should have order less than i2.
// This func doesn't deal with pointers
func reflectLess(i1, i2 interface{}) bool {
	if i1 == nil && i2 == nil {
		return true
	}
	if i1 == nil {
		return true
	}
	if i2 == nil {
		return false
	}

	v1 := reflect.ValueOf(i1)
	v2 := reflect.ValueOf(i2)

	if v1.Kind() != v2.Kind() {
		return fmt.Sprint(i1) < fmt.Sprint(i2)
	}

	switch v1.Kind() {
	case reflect.Bool:
		b1, b2 := i1.(bool), i2.(bool)
		if b1 && !b2 { // treating bool as number, then only [1, 0] returns false
			return false
		}
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() < v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v1.Uint() < v2.Uint()
	case reflect.Float32, reflect.Float64:
		return v1.Float() < v2.Float()
	case reflect.String:
		return v1.String() < v2.String()
	default:
		return fmt.Sprint(i1) < fmt.Sprint(i2)
	}
}

// Query performs a query on the current Database.
//
// FIXME: Curent implementation is not complete. It assumes the first
// argument being the type of Record and always returns a Rows that
// iterates over all records of that type.
func (db fileDatabase) Query(query *skydb.Query) (*skydb.Rows, error) {
	var outbuf bytes.Buffer
	var errbuf bytes.Buffer

	cmd := exec.Command("sh", "-c", "cat "+filepath.Join(db.Dir, query.Type, "*"))
	cmd.Stdout = &outbuf
	cmd.Stdin = &errbuf

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// NOTE: this cast is platform depedent and is only tested
			// on UNIX-like system
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				// cat has a exit status of 1 if directory not found
				if status.ExitStatus() != 1 {
					log.WithFields(log.Fields{
						"ExitStatus": status.ExitStatus(),
					}).Panicln("unexpected exit status")
				}
			}
		}

		log.WithFields(log.Fields{
			"err":    err.Error(),
			"stderr": errbuf.String(),
			"path":   db.Dir,
		}).Infoln("Failed to execute cat")

		return skydb.EmptyRows, nil
	}

	records := []skydb.Record{}
	scanner := bufio.NewScanner(&outbuf)
	for scanner.Scan() {
		record := skydb.Record{}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		record.DatabaseID = db.UserID
		fmt.Printf("record.CreatedAt.Unix() = %v\n", record.CreatedAt.Unix())
		if record.CreatedAt.Unix() == zeroUnix {
			record.CreatedAt = time.Time{}
		}
		if record.UpdatedAt.Unix() == zeroUnix {
			record.UpdatedAt = time.Time{}
		}
		records = append(records, record)
	}

	if len(query.Sorts) > 0 {
		if len(query.Sorts) > 1 {
			return nil, errors.New("multiple sort orders not supported")
		}

		newRecordSorter(records, query.Sorts[0]).Sort()
	}

	return skydb.NewRows(skydb.NewMemoryRows(records)), nil
}

func (db fileDatabase) QueryCount(query *skydb.Query) (uint64, error) {
	panic("not implemented")
}

func (db fileDatabase) Extend(recordType string, schema skydb.RecordSchema) error {
	// do nothing
	return nil
}

func (db fileDatabase) GetSubscription(key string, deviceID string, subscription *skydb.Subscription) error {
	return db.subscriDB.Get(key, subscription)
}

func (db fileDatabase) SaveSubscription(subscription *skydb.Subscription) error {
	return db.subscriDB.Save(subscription)
}

func (db fileDatabase) DeleteSubscription(key string, deviceID string) error {
	return db.subscriDB.Delete(key)
}

func (db fileDatabase) GetMatchingSubscriptions(record *skydb.Record) []skydb.Subscription {
	return db.subscriDB.GetMatchingSubscriptions(record)
}

func (db fileDatabase) GetSubscriptionsByDeviceID(deviceID string) []skydb.Subscription {
	return db.subscriDB.GetSubscriptionsByDeviceID(deviceID)
}

func (db fileDatabase) recordPath(id skydb.RecordID) string {
	return filepath.Join(db.Dir, id.Type, id.Key)
}

func init() {
	skydb.Register("fs", skydb.DriverFunc(Open))
}
