package testrunner

import (
	"encoding/json"
	"fmt"
)

func MatchCLIOutput(output CLIOutput, exitCode int, stdout string, stderr string) (violations []MatchViolation, err error) {
	if output.ExitCode != "" {
		exitCodeJSON, err := json.Marshal(exitCode)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal exit_code: %w", err)
		}
		exitCodeViolations, err := MatchJSON(string(exitCodeJSON), output.ExitCode)
		if err != nil {
			return nil, err
		}
		violations = append(violations, exitCodeViolations...)
	}

	if output.Stdout != "" {
		stdoutJSON, err := json.Marshal(stdout)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stdout: %w", err)
		}
		stdoutViolations, err := MatchJSON(string(stdoutJSON), output.Stdout)
		if err != nil {
			return nil, err
		}
		violations = append(violations, stdoutViolations...)
	}

	if output.Stderr != "" {
		stderrJSON, err := json.Marshal(stderr)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal stderr: %w", err)
		}
		stderrViolations, err := MatchJSON(string(stderrJSON), output.Stderr)
		if err != nil {
			return nil, err
		}
		violations = append(violations, stderrViolations...)
	}

	return violations, nil
}
