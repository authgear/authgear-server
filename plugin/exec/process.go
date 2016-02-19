package exec

import (
	"bufio"
	"encoding/json"
	"fmt"
	osexec "os/exec"

	log "github.com/Sirupsen/logrus"

	skyplugin "github.com/oursky/skygear/plugin"
	"github.com/oursky/skygear/plugin/common"
	"github.com/oursky/skygear/skyconfig"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skyconv"
	"golang.org/x/net/context"
)

var startCommand = func(cmd *osexec.Cmd, in []byte) (out []byte, err error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
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

	pluginLog := []byte{}
	stdErr := bufio.NewScanner(stderr)
	for stdErr.Scan() {
		pluginLog = append(pluginLog, []byte("\n")...)
		pluginLog = append(pluginLog, stdErr.Bytes()...)
	}
	log.Debug("exec stderr : ", string(pluginLog))

	s := bufio.NewScanner(stdout)
	for s.Scan() {
		out = append(out, s.Bytes()...)
	}

	err = stdout.Close()
	if err != nil {
		return
	}

	err = cmd.Wait()
	return
}

type execTransport struct {
	Path        string
	Args        []string
	DBConfig    string
	Config      skyconfig.Configuration
	initHandler skyplugin.TransportInitHandler
	state       skyplugin.TransportState
}

func (p *execTransport) run(args []string, env []string, in []byte) (out []byte, err error) {
	finalArgs := make([]string, len(p.Args)+len(args))
	for i, arg := range p.Args {
		finalArgs[i] = arg
	}
	for i, arg := range args {
		finalArgs[i+len(p.Args)] = arg
	}

	encodedConfig, err := common.EncodeBase64JSON(p.Config)
	if err != nil {
		return nil, err
	}

	cmd := osexec.Command(p.Path, finalArgs...)
	cmd.Env = []string{
		"DATABASE_URL=" + p.DBConfig,
		fmt.Sprintf("SKYGEAR_CONFIG=%s", encodedConfig),
	}
	for _, envLine := range env {
		cmd.Env = append(cmd.Env, envLine)
	}
	log.Debugf("Calling with Env %v", cmd.Env)
	log.Debugf("Calling %s %s with     : %s", cmd.Path, cmd.Args, in)
	out, err = startCommand(cmd, in)
	log.Debugf("Called  %s %s returning: %s", cmd.Path, cmd.Args, out)

	return
}

// runProc unwrap inner error returned from run
func (p *execTransport) runProc(args []string, env []string, in []byte) (out []byte, err error) {
	var data []byte
	data, err = p.run(args, env, in)
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

func (p *execTransport) State() skyplugin.TransportState {
	return p.state
}

func (p *execTransport) SetInitHandler(f skyplugin.TransportInitHandler) {
	p.initHandler = f
}

func (p *execTransport) setState(state skyplugin.TransportState) {
	if state != p.state {
		oldState := p.state
		p.state = state
		log.Infof("Transport state changes from %v to %v.", oldState, p.state)
	}
}

func (p *execTransport) RequestInit() {
	out, err := p.RunInit()
	if p.initHandler != nil {
		handlerError := p.initHandler(out, err)
		if err != nil || handlerError != nil {
			p.setState(skyplugin.TransportStateError)
			return
		}
	}
	p.setState(skyplugin.TransportStateReady)
}

func (p *execTransport) RunInit() (out []byte, err error) {
	out, err = p.run([]string{"init"}, []string{}, []byte{})
	return
}

func (p *execTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	pluginCtx := skyplugin.ContextMap(ctx)
	encodedCtx, err := common.EncodeBase64JSON(pluginCtx)
	if err != nil {
		return nil, err
	}
	env := []string{
		fmt.Sprintf("SKYGEAR_CONTEXT=%s", encodedCtx),
	}
	out, err = p.runProc([]string{"op", name}, env, in)
	return
}

func (p *execTransport) RunHandler(name string, in []byte) (out []byte, err error) {
	out, err = p.runProc([]string{"handler", name}, []string{}, in)
	return
}

func (p *execTransport) RunHook(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
	param := map[string]interface{}{
		"record":   (*skyconv.JSONRecord)(record),
		"original": (*skyconv.JSONRecord)(originalRecord),
	}
	in, err := json.Marshal(param)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %v", err)
	}

	pluginCtx := skyplugin.ContextMap(ctx)
	encodedCtx, err := common.EncodeBase64JSON(pluginCtx)
	if err != nil {
		return nil, err
	}
	env := []string{
		fmt.Sprintf("SKYGEAR_CONTEXT=%s", encodedCtx),
	}
	out, err := p.runProc([]string{"hook", hookName}, env, in)
	if err != nil {
		return nil, err
	}

	var recordout skydb.Record
	if err := json.Unmarshal(out, (*skyconv.JSONRecord)(&recordout)); err != nil {
		log.WithField("data", string(out)).Error("failed to unmarshal record")
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
	recordout.OwnerID = record.OwnerID
	recordout.CreatedAt = record.CreatedAt
	recordout.CreatorID = record.CreatorID
	recordout.UpdatedAt = record.UpdatedAt
	recordout.UpdaterID = record.UpdaterID

	return &recordout, nil
}

func (p *execTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	out, err = p.runProc([]string{"timer", name}, []string{}, in)
	return
}

func (p *execTransport) RunProvider(request *skyplugin.AuthRequest) (*skyplugin.AuthResponse, error) {
	req := map[string]interface{}{
		"auth_data": request.AuthData,
	}

	in, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %v", err)
	}

	out, err := p.runProc([]string{"provider", request.ProviderName, request.Action}, []string{}, in)
	if err != nil {
		return nil, err
	}

	resp := skyplugin.AuthResponse{}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &resp, nil
}

type execTransportFactory struct {
}

func (f execTransportFactory) Open(path string, args []string, config skyconfig.Configuration) (transport skyplugin.Transport) {
	log.Debugf("plugin exec args %v", args)
	if !config.App.DevMode {
		log.Warn("plugin exec transport is development use only")
	}
	if path == "" {
		path = "py-skygear"
	}
	args = append(args, "--subprocess")
	transport = &execTransport{
		Path:     path,
		Args:     args,
		DBConfig: config.DB.Option,
		Config:   config,
		state:    skyplugin.TransportStateUninitialized,
	}
	return
}

func init() {
	skyplugin.RegisterTransport("exec", execTransportFactory{})
}
