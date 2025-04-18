package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type DockerContainer struct {
	ID    string `json:"ID"`
	Names string `json:"Names"`
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

func NewDockerRunOptionsForStarting(image string) DockerRunOptions {
	if image == "" {
		image = fmt.Sprintf("%v:%v", DefaultDockerName_NoTag, Version)
	}

	return DockerRunOptions{
		Detach: true,
		Volume: []string{fmt.Sprintf("%v:/var/lib/authgearonce", NameDockerVolume)},
		Publish: []string{
			"80:80",
			"443:443",
			"5432:5432",
			"9001:9001",
			"8090:8090",
		},
		Name:  NameDockerContainer,
		Image: image,
	}
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

func DockerLs(ctx context.Context) ([]DockerContainer, error) {
	c := exec.CommandContext(ctx, "docker", "ps", "--all", "--format", "json", "--no-trunc")
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return nil, errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}

	var cs []DockerContainer
	for line := range lines(stdout) {
		var c DockerContainer

		err = json.Unmarshal([]byte(line), &c)
		if err != nil {
			return nil, err
		}

		cs = append(cs, c)
	}
	return cs, nil
}

func DockerStart(ctx context.Context, name string) error {
	c := exec.CommandContext(ctx, "docker", "start", name)
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return nil
}

func DockerStop(ctx context.Context, name string) error {
	c := exec.CommandContext(ctx, "docker", "stop", name)
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return nil
}
