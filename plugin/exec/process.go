package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	osexec "os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/oddb"
	odplugin "github.com/oursky/ourd/plugin"
	"github.com/oursky/ourd/plugin/common"
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
		Result json.RawMessage   `json:"result"`
		Err    *common.ExecError `json:"error"`
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

func (p execTransport) RunHook(recordType string, trigger string, record *oddb.Record, originalRecord *oddb.Record) (*oddb.Record, error) {
	param := map[string]interface{}{
		"record":   (*common.JSONRecord)(record),
		"original": (*common.JSONRecord)(originalRecord),
	}
	in, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %v", err)
	}

	hookName := fmt.Sprintf("%v:%v", recordType, trigger)
	out, err := p.runProc([]string{"hook", hookName}, in)
	if err != nil {
		return nil, fmt.Errorf("run %s: %v", hookName, err)
	}

	var recordout oddb.Record
	if err := json.Unmarshal(out, (*common.JSONRecord)(&recordout)); err != nil {
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

func (p execTransport) RunProvider(request *odplugin.AuthRequest) (*odplugin.AuthResponse, error) {
	req := map[string]interface{}{
		"auth_data": request.AuthData,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %v", err)
	}

	out, err := p.runProc([]string{"provider", request.ProviderName, request.Action}, in)
	if err != nil {
		name := fmt.Sprintf("%v:%v", request.ProviderName, request.Action)
		return nil, fmt.Errorf("run %s: %v", name, err)
	}

	resp := odplugin.AuthResponse{}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &resp, nil
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
