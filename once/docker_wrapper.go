package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/goaux/decowriter"
	"golang.org/x/sys/unix"
)

type ChildResult struct {
	Child        *Child
	ProcessState *os.ProcessState
	Err          error
}

type Child struct {
	cmd           *exec.Cmd
	resultChannel chan ChildResult
}

func NewChild(commandline []string, stdout io.Writer, stderr io.Writer) *Child {
	cmd := exec.Command(commandline[0], commandline[1:]...)
	// By setting Stdout and Stderr,
	// Go will copy the process stdout and the process stderr to Stdout and Stderr for us.
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	resultChannel := make(chan ChildResult)
	return &Child{
		cmd:           cmd,
		resultChannel: resultChannel,
	}
}

func (c *Child) String() string {
	return c.cmd.Path
}

func (c *Child) Start() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		var exitError *exec.ExitError
		err := c.cmd.Wait()
		if errors.As(err, &exitError) {
			c.resultChannel <- ChildResult{
				Child:        c,
				ProcessState: exitError.ProcessState,
			}
		} else if err != nil {
			c.resultChannel <- ChildResult{
				Child: c,
				Err:   err,
			}
		} else {
			c.resultChannel <- ChildResult{
				Child:        c,
				ProcessState: c.cmd.ProcessState,
			}
		}
	}()
	return nil
}

func (c *Child) Chan() <-chan ChildResult {
	return c.resultChannel
}

func (c *Child) Signal(signal os.Signal) error {
	return c.cmd.Process.Signal(signal)
}

type ParentState int

const (
	ParentStateRunning = 0
	ParentStateReaping = 1
)

func SignalName(sig os.Signal) string {
	if syscallSignal, ok := sig.(syscall.Signal); ok {
		return unix.SignalName(syscallSignal)
	}
	return sig.String()
}

func main() {
	parentStdout := decowriter.New(os.Stdout, []byte("docker_wrapper | "), []byte(""))
	//parentStderr := decowriter.New(os.Stderr, []byte("docker_wrapper | "), []byte(""))

	parentState := ParentStateRunning

	// https://pkg.go.dev/os/signal@go1.23.6#hdr-Default_behavior_of_signals_in_Go_programs
	signalsForReaping := []os.Signal{
		// For some unknown reason, the commented signals are not available on Linux.
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGTRAP,
		syscall.SIGABRT,
		// syscall.SIGSTKFLT,
		// syscall.SIGEMT,
		syscall.SIGSYS,
	}
	// signal.Notify says we must use buffered channel.
	// It also says a buffer of size 1 is usually enough.
	signalChan := make(chan os.Signal, 1)
	// We want to forward all signal to child, so listen for all signals.
	signal.Notify(signalChan)

	selectCases := []reflect.SelectCase{
		// case sig := <-signalChan:
		{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(signalChan)},
	}

	postgres := NewChild(
		[]string{"postgres"},
		decowriter.New(os.Stdout, []byte("postgres       | "), []byte("")),
		decowriter.New(os.Stderr, []byte("postgres       | "), []byte("")),
	)
	redis := NewChild(
		[]string{"redis-server", "/etc/redis/redis.conf"},
		decowriter.New(os.Stdout, []byte("redis          | "), []byte("")),
		decowriter.New(os.Stderr, []byte("redis          | "), []byte("")),
	)
	nginx := NewChild(
		[]string{"nginx"},
		decowriter.New(os.Stdout, []byte("nginx          | "), []byte("")),
		decowriter.New(os.Stderr, []byte("nginx          | "), []byte("")),
	)
	certbot := NewChild(
		[]string{"docker-certbot.py"},
		decowriter.New(os.Stdout, []byte("docker-certbot | "), []byte("")),
		decowriter.New(os.Stderr, []byte("docker-certbot | "), []byte("")),
	)
	minio := NewChild(
		[]string{"minio", "server", "/var/lib/minio/data", "--console-address", ":9001"},
		decowriter.New(os.Stdout, []byte("minio          | "), []byte("")),
		decowriter.New(os.Stderr, []byte("minio          | "), []byte("")),
	)
	childResults := []ChildResult{}
	startedChildren := []*Child{}

	start := func(child *Child) {
		err := child.Start()
		if err != nil {
			childResults = append(childResults, ChildResult{
				Child: child,
				Err:   err,
			})
		} else {
			startedChildren = append(startedChildren, child)
			// case childResult := <- child.Chan()
			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(child.Chan())})
		}
	}

	startReap := func(sig os.Signal) {
		if parentState == ParentStateReaping {
			return
		}
		parentState = ParentStateReaping
		for _, child := range startedChildren {
			fmt.Fprintf(parentStdout, "sending %v to %v\n", SignalName(sig), child)
			child.Signal(sig)
		}
	}

	start(postgres)
	start(redis)
	start(nginx)
	start(certbot)
	start(minio)

	// Some child failed to start, trigger reap now.
	if len(childResults) > 0 {
		startReap(syscall.SIGTERM)
	}

	for len(childResults) < len(startedChildren) {
		// This is a dynamic form of the select statement.
		chosenIdx, recv, _ := reflect.Select(selectCases)
		switch {
		// case sig := <-signalChan:
		case chosenIdx == 0:
			sig := recv.Interface().(os.Signal)
			fmt.Fprintf(parentStdout, "received signal: %v\n", SignalName(sig))
			for _, signalForReaping := range signalsForReaping {
				if signalForReaping == sig {
					fmt.Fprintf(parentStdout, "start reaping because signal is %v\n", SignalName(sig))
					startReap(sig)
				}
			}
		// case childResult := <-child.Chan()
		default:
			childResult := recv.Interface().(ChildResult)
			childResults = append(childResults, childResult)
			// We are supposed to be running but one of the child terminated.
			// Start reaping.
			if parentState == ParentStateRunning {
				fmt.Fprintf(parentStdout, "start reaping because %v has terminated\n", childResult.Child)
				startReap(syscall.SIGTERM)
			}
		}
	}

	// parentExitCode is normally 0.
	// If one of the child exited with non-zero, then the exit code of the parent is non-zero.
	parentExitCode := 0
	for _, childResult := range childResults {
		if childResult.Err != nil {
			fmt.Fprintf(parentStdout, "%v resulted in err: %v\n", childResult.Child, childResult.Err)
		} else if childResult.ProcessState != nil {
			childExitCode := childResult.ProcessState.ExitCode()
			fmt.Fprintf(parentStdout, "%v exited with %v\n", childResult.Child, childExitCode)
			if childExitCode != 0 {
				parentExitCode = childExitCode
			}
		}
	}
	fmt.Fprintf(parentStdout, "exiting with %v\n", parentExitCode)
	os.Exit(parentExitCode)
}
