package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"iter"
	"os/exec"
	"strings"
)

const (
	DockerVolumeScopeLocal = "local"
)

type DockerVolume struct {
	Name  string `json:"Name"`
	Scope string `json:"Scope"`
}

func runCmd(c *exec.Cmd) (stdout string, stderr string, err error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	c.Stdout = &stdoutBuf
	c.Stderr = &stderrBuf
	err = c.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	return
}

type CmdError struct {
	Stdout string
	Stderr string
}

func (e *CmdError) Error() string {
	return e.Stderr
}

func lines(stdout string) iter.Seq[string] {
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	return func(yield func(string) bool) {
		for scanner.Scan() {
			line := scanner.Text()
			if !yield(line) {
				return
			}
		}
	}
}

func DockerVolumeLs(ctx context.Context) ([]DockerVolume, error) {
	c := exec.CommandContext(ctx, "docker", "volume", "ls", "--format", "json")
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return nil, errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}

	var vs []DockerVolume
	for line := range lines(stdout) {
		var v DockerVolume

		err = json.Unmarshal([]byte(line), &v)
		if err != nil {
			return nil, err
		}

		vs = append(vs, v)
	}
	return vs, nil
}

func DockerVolumeCreate(ctx context.Context, name string) error {
	c := exec.CommandContext(ctx, "docker", "volume", "create", name)
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return nil
}

type DockerRunOptions struct {
	Detach  bool
	Rm      bool
	Volume  []string
	Publish []string
	Env     []string
	Name    string
	Image   string
	Command []string
}

func DockerRun(ctx context.Context, opts DockerRunOptions) error {
	args := []string{"run"}

	if opts.Detach {
		args = append(args, "--detach")
	}
	if opts.Rm {
		args = append(args, "--rm")
	}
	for _, v := range opts.Volume {
		args = append(args, "--volume", v)
	}
	for _, p := range opts.Publish {
		args = append(args, "--publish", p)
	}
	for _, e := range opts.Env {
		args = append(args, "--env", e)
	}
	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	args = append(args, opts.Image)
	for _, c := range opts.Command {
		args = append(args, c)
	}

	c := exec.CommandContext(ctx, "docker", args...)
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return nil
}
