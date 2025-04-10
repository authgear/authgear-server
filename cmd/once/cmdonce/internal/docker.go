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
		return nil, errors.Join(&CmdError{Stderr: stderr}, err)
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
	_, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stderr: stderr}, err)
	}
	return nil
}
