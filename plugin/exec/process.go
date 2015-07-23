package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	osexec "os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/oddb/oddbconv"
	odplugin "github.com/oursky/ourd/plugin"
)

var startCommand = func(cmd *osexec.Cmd, in []byte) (out []byte, err error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	_, err = stdin.Write(in)
	if err != nil {
		return
	}

	err = stdin.Close()
	if err != nil {
		return
	}

	s := bufio.NewScanner(stdout)
	if !s.Scan() {
		if err = s.Err(); err == nil {
			// reached EOF
			out = []byte{}
		} else {
			return
		}
	} else {
		out = s.Bytes()
	}

	err = stdout.Close()
	if err != nil {
		return
	}

	err = cmd.Wait()
	return
}

// execError is error resulted from application logic of plugin (e.g.
// an exception thrown within a lambda function)
type execError struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
}

func (err *execError) Error() string {
	return err.Name + "\n" + err.Description
}

type execTransport struct {
	Path string
	Args []string
}

func (p *execTransport) run(args []string, in []byte) (out []byte, err error) {
	finalArgs := make([]string, len(p.Args)+len(args))
	for i, arg := range p.Args {
		finalArgs[i] = arg
	}
	for i, arg := range args {
		finalArgs[i+len(p.Args)] = arg
	}

	cmd := osexec.Command(p.Path, finalArgs...)

	log.Debugf("Calling %s %s with     : %s", cmd.Path, cmd.Args, in)
	out, err = startCommand(cmd, in)
	log.Debugf("Called  %s %s returning: %s", cmd.Path, cmd.Args, out)

	return
}

// runProc unwrap inner error returned from run
func (p *execTransport) runProc(args []string, in []byte) (out []byte, err error) {
	var data []byte
	data, err = p.run(args, in)
	if err != nil {
		return
	}

	var resp struct {
		Result json.RawMessage `json:"result"`
		Err    *execError      `json:"error"`
	}

	jsonErr := json.Unmarshal(data, &resp)
	if jsonErr != nil {
		err = fmt.Errorf("failed to parse response: %v", jsonErr)
		return
	}

	if resp.Err != nil {
		err = resp.Err
		return
	}

	out = resp.Result
	return
}

func (p execTransport) RunInit() (out []byte, err error) {
	out, err = p.run([]string{"init"}, []byte{})
	return
}

func (p execTransport) RunLambda(name string, in []byte) (out []byte, err error) {
	out, err = p.runProc([]string{"op", name}, in)
	return
}

func (p execTransport) RunHandler(name string, in []byte) (out []byte, err error) {
	out, err = p.runProc([]string{"handler", name}, in)
	return
}

func (p execTransport) RunHook(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error) {
	in, err := json.Marshal((*jsonRecord)(record))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %v", err)
	}

	hookName := fmt.Sprintf("%v:%v", recordType, trigger)
	out, err := p.runProc([]string{"hook", hookName}, in)
	if err != nil {
		return nil, fmt.Errorf("run %s: %v", hookName, err)
	}

	var recordout oddb.Record
	if err := json.Unmarshal(out, (*jsonRecord)(&recordout)); err != nil {
		log.WithField("data", string(out)).Error("failed to unmarshal record")
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
	recordout.OwnerID = record.OwnerID
	return &recordout, nil
}

func (p execTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	out, err = p.runProc([]string{"timer", name}, in)
	return
}

func (p execTransport) RunProvider(providerID string, action string, in []byte) (out []byte, err error) {
	out, err = p.run([]string{"provider", providerID, action}, in)
	return
}

type execTransportFactory struct {
}

func (f execTransportFactory) Open(path string, args []string) (transport odplugin.Transport) {
	transport = execTransport{
		Path: path,
		Args: args,
	}
	return
}

func init() {
	odplugin.RegisterTransport("exec", execTransportFactory{})
}

type jsonRecord oddb.Record

func (record *jsonRecord) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	for key, value := range record.Data {
		switch v := value.(type) {
		case time.Time:
			data[key] = (oddbconv.MapTime)(v)
		case oddb.Asset:
			data[key] = (oddbconv.MapAsset)(v)
		case oddb.Reference:
			data[key] = (oddbconv.MapReference)(v)
		default:
			data[key] = value
		}
	}

	m := map[string]interface{}{}
	oddbconv.MapData(data).ToMap(m)

	m["_id"] = record.ID
	m["_ownerID"] = record.OwnerID
	m["_access"] = record.ACL

	return json.Marshal(m)
}

func (record *jsonRecord) UnmarshalJSON(data []byte) (err error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	var (
		id      oddb.RecordID
		acl     oddb.RecordACL
		dataMap map[string]interface{}
	)

	extractor := newMapExtractor(m)
	extractor.DoString("_id", func(s string) error {
		return id.UnmarshalText([]byte(s))
	})
	extractor.DoSlice("_access", func(slice []interface{}) error {
		return acl.InitFromJSON(slice)
	})
	if extractor.Err() != nil {
		return extractor.Err()
	}

	m = sanitizedDataMap(m)
	if err := (*oddbconv.MapData)(&dataMap).FromMap(m); err != nil {
		return err
	}

	record.ID = id
	record.ACL = acl
	record.Data = dataMap
	return nil
}

func sanitizedDataMap(m map[string]interface{}) map[string]interface{} {
	mm := map[string]interface{}{}
	for key, value := range m {
		if key[0] != '_' {
			mm[key] = value
		}
	}
	return mm
}

// mapExtractor helps to extract value of a key from a map
//
// potential candicate of a package
type mapExtractor struct {
	m   map[string]interface{}
	err error
}

func newMapExtractor(m map[string]interface{}) *mapExtractor {
	return &mapExtractor{m: m}
}

// Do execute doFunc if key exists in the map
// The key will always be removed no matter error occurred previously
func (e *mapExtractor) Do(key string, doFunc func(interface{}) error) {
	value, ok := e.m[key]
	delete(e.m, key)

	if e.err != nil {
		return
	}

	log.Printf("e.m = %#v, key = %#v", e.m, key)
	log.Printf("value = %#v, ok = %#v", value, ok)

	if ok {
		e.err = doFunc(value)
		delete(e.m, key)
	} else {
		e.err = fmt.Errorf(`no key "%s" in map`, key)
	}
}

func (e *mapExtractor) DoString(key string, doFunc func(string) error) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(string); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not string", key, i)
	})
}

func (e *mapExtractor) DoMap(key string, doFunc func(map[string]interface{}) error) {
	e.Do(key, func(i interface{}) error {
		if m, ok := i.(map[string]interface{}); ok {
			return doFunc(m)
		}
		return fmt.Errorf("key %s is of type %T, not map[string]interface{}", key, i)
	})
}

func (e *mapExtractor) DoSlice(key string, doFunc func([]interface{}) error) {
	e.Do(key, func(i interface{}) error {
		if slice, ok := i.([]interface{}); ok {
			return doFunc(slice)
		}
		return fmt.Errorf("key %s is of type %T, not map[string]interface{}", key, i)
	})
}

func (e *mapExtractor) Err() error {
	return e.err
}
