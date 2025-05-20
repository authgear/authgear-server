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
	"slices"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/slice"
)

const (
	DockerVolumeScopeLocal = "local"
)

const (
	CertbotExitCode10 = 10
)

var DockerPublishedPorts []int = []int{80, 443}

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

func getExitError(err error) (*exec.ExitError, bool) {
	var e *exec.ExitError
	ok := errors.As(err, &e)
	if ok {
		return e, true
	}
	return nil, false
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
	Restart string
	Rm      bool
	Volume  []string
	Publish []string
	Env     []string
	Name    string
	Image   string
	Command []string
}

func NewDockerRunOptionsForStarting(image string) DockerRunOptions {
	return DockerRunOptions{
		Detach:  true,
		Restart: "always",
		Volume:  []string{fmt.Sprintf("%v:/var/lib/authgearonce", NameDockerVolume)},
		Publish: slice.Map(DockerPublishedPorts, dockerPublishPortOnAllInterfaces),
		Name:    NameDockerContainer,
		Image:   image,
	}
}

// dockerPublishPortOnAllInterfaces takes port and returns port:port.
func dockerPublishPortOnAllInterfaces(port int) string {
	return fmt.Sprintf("%v:%v", port, port)
}

type DockerRunResult struct {
	Stdout string
	Stderr string
}

func DockerRun(ctx context.Context, opts DockerRunOptions) (*DockerRunResult, error) {
	args := []string{"run"}

	if opts.Detach {
		args = append(args, "--detach")
	}
	if opts.Restart != "" {
		args = append(args, "--restart", opts.Restart)
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
		return nil, errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return &DockerRunResult{Stdout: stdout, Stderr: stderr}, nil
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

type DockerRmOptions struct {
	Force bool
}

func DockerRm(ctx context.Context, name string, options DockerRmOptions) error {
	args := []string{"rm"}

	if options.Force {
		args = append(args, "-f")
	}

	args = append(args, name)

	c := exec.CommandContext(ctx, "docker", args...)
	stdout, stderr, err := runCmd(c)
	if err != nil {
		return errors.Join(&CmdError{Stdout: stdout, Stderr: stderr}, err)
	}
	return nil
}

func GetPersistentEnvironmentVariableInVolume(ctx context.Context, envVarName string) (image string, err error) {
	opts := DockerRunOptions{
		Rm:     true,
		Volume: []string{fmt.Sprintf("%v:/var/lib/authgearonce", NameDockerVolume)},
		// Use busybox to inspect the volume.
		Image: "busybox:1",
		Command: []string{
			"sh",
			"-c",
			fmt.Sprintf(`</var/lib/authgearonce/env.sh awk -F = '/%v/ { print $2 }'`, envVarName),
		},
	}

	result, err := DockerRun(ctx, opts)
	if err != nil {
		return
	}

	image = strings.TrimSpace(result.Stdout)
	return
}

// CheckAllPublishedPortsNotListening loops through DockerPublishedPorts
// and checks if any of them are already listening on the host.
// It returns an error if any of the ports are already in use.
func CheckAllPublishedPortsNotListening() error {
	var errs []error

	for _, port := range DockerPublishedPorts {
		if err := CheckTCPPortIsListening(port); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func CheckVolumeExists(ctx context.Context) (bool, error) {
	volumes, err := DockerVolumeLs(ctx)
	if err != nil {
		return false, err
	}
	if slices.ContainsFunc(volumes, func(v DockerVolume) bool {
		return v.Name == NameDockerVolume && v.Scope == DockerVolumeScopeLocal
	}) {
		return true, nil
	}
	return false, nil
}

func DockerRunWithCertbotErrorHandling(ctx context.Context, opts DockerRunOptions) (*DockerRunResult, error) {

	result, err := DockerRun(ctx, opts)
	if err != nil {
		if exitErr, ok := getExitError(err); ok {
			if exitErr.ProcessState.ExitCode() == CertbotExitCode10 {
				err = errors.Join(ErrCertbotExitCode10, err)
			}
		}
	}
	return result, err
}
