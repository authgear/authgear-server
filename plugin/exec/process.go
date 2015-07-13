package exec

import (
	"bufio"
	"fmt"

	log "github.com/Sirupsen/logrus"

	odplugin "github.com/oursky/ourd/plugin"
	osexec "os/exec"
)

func startCommand(cmd *osexec.Cmd, in []byte) (out []byte, err error) {
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

	out, err = startCommand(cmd, in)
	log.Debugf("Called process %s %s %s", p.Path, finalArgs, in)
	return
}

func (p execTransport) RunInit() (out []byte, err error) {
	out, err = p.run([]string{"init"}, []byte{})
	return
}

func (p execTransport) RunLambda(name string, in []byte) (out []byte, err error) {
	out, err = p.run([]string{"op", name}, in)
	return
}

func (p execTransport) RunHandler(name string, in []byte) (out []byte, err error) {
	out, err = p.run([]string{"handler", name}, in)
	return
}

func (p execTransport) RunHook(recordType string, trigger string, in []byte) (out []byte, err error) {
	hookName := fmt.Sprintf("%v:%v", recordType, trigger)
	out, err = p.run([]string{"hook", hookName}, in)
	return
}

func (p execTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	out, err = p.run([]string{"timer", name}, in)
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
