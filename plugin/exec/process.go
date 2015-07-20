package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	osexec "os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
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
	in, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %v", err)
	}

	hookName := fmt.Sprintf("%v:%v", recordType, trigger)
	out, err := p.runProc([]string{"hook", hookName}, in)
	if err != nil {
		return nil, fmt.Errorf("run %s: %v", hookName, err)
	}

	var recordout oddb.Record
	if err := json.Unmarshal(out, &recordout); err != nil {
		log.WithField("data", out).Error("failed to unmarshal record")
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
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
