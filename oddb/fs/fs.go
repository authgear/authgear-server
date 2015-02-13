package file

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/oursky/ourd/oddb"
)

const publicDBKey = "_public"
const privateDBKey = "_private"

// fileConn implements oddb.Conn interface
type fileConn struct {
	dir string
}

// Open returns a new connection to fs implementation
func Open(dir string) (oddb.Conn, error) {
	return &fileConn{dir}, nil
}

func (conn fileConn) Container(key string) (oddb.Container, error) {
	containerPath := filepath.Join(conn.dir, key)
	return newContainer(containerPath, key), nil
}
func (conn fileConn) Close() error {
	return nil
}

type fileContainer struct {
	Dir      string
	Key      string
	publicDB oddb.Database
}

func newContainer(dir string, key string) *fileContainer {
	publicDBPath := filepath.Join(dir, publicDBKey)
	return &fileContainer{
		Dir:      dir,
		Key:      key,
		publicDB: newDatabase(publicDBPath, publicDBKey),
	}
}

func (container fileContainer) PublicDB() oddb.Database {
	return container.publicDB
}

func (container fileContainer) PrivateDB(userKey string) oddb.Database {
	dbPath := filepath.Join(container.Dir, userKey)
	return newDatabase(dbPath, privateDBKey)
}

type fileDatabase struct {
	Dir string
	Key string
}

func newDatabase(dir string, key string) *fileDatabase {
	return &fileDatabase{
		Dir: dir,
		Key: key,
	}
}

func (db fileDatabase) ID() string {
	return db.Key
}

func (db fileDatabase) Get(key string, record *oddb.Record) error {
	file, err := os.Open(db.recordPath(record))
	if err != nil {
		return err
	}

	jsonDecoder := json.NewDecoder(file)
	return jsonDecoder.Decode(record)
}

func (db fileDatabase) Save(record *oddb.Record) error {
	filePath := db.recordPath(record)
	if err := os.MkdirAll(db.Dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	jsonEncoder := json.NewEncoder(file)
	return jsonEncoder.Encode(record)
}

func (db fileDatabase) Delete(key string) error {
	return os.Remove(filepath.Join(db.Dir, key))
}

// Query performs a query on the current Database.
//
// FIXME: Curent implementation is not complete. It assumes the first
// argument being the type of Record and always returns a Rows that
// iterates over all records of that type.
func (db fileDatabase) Query(query string, args ...interface{}) (oddb.Rows, error) {
	const grepFmt = "grep -he \"{\\\"_type\\\":\\\"%v\\\"\" %v"

	grep := fmt.Sprintf(grepFmt, args[0], filepath.Join(db.Dir, "*"))

	var outbuf bytes.Buffer
	var errbuf bytes.Buffer

	cmd := exec.Command("sh", "-c", grep)
	cmd.Stdout = &outbuf
	cmd.Stdin = &errbuf

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// NOTE: this cast is platform depedent and is only tested
			// on UNIX-like system
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					// grep has a exit status of 1 if it finds nothing
					// See: http://www.gnu.org/software/grep/manual/html_node/Exit-Status.html
					return &memoryRows{0, []oddb.Record{}}, nil
				}
			}
		}
		log.Fatalf("Failed to grep: %v\nStderr: %v", err.Error(), errbuf.String())
	}

	records := []oddb.Record{}
	scanner := bufio.NewScanner(&outbuf)
	for scanner.Scan() {
		record := oddb.Record{}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return &memoryRows{0, records}, nil
}

func (db fileDatabase) recordPath(record *oddb.Record) string {
	return filepath.Join(db.Dir, record.Key)
}

type memoryRows struct {
	currentRowIndex int
	records         []oddb.Record
}

func (rs *memoryRows) Close() error {
	return nil
}

func (rs *memoryRows) Next(record *oddb.Record) error {
	if rs.currentRowIndex >= len(rs.records) {
		return io.EOF
	}

	*record = rs.records[rs.currentRowIndex]
	rs.currentRowIndex = rs.currentRowIndex + 1
	return nil
}

func init() {
	oddb.Register("fs", oddb.DriverFunc(Open))
}
