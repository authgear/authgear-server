package editor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var ErrEditorCancelled = errors.New("edit cancelled")

var envs = []string{
	"VISUAL",
	"EDITOR",
}

func getEditor() []string {
	for _, env := range envs {
		editor := os.Getenv(env)
		if editor != "" {
			return strings.Split(editor, " ")
		}
	}

	if runtime.GOOS == "windows" {
		return []string{"notepad"}
	}
	return []string{"vi"}
}

func editInEditor(filePrefix, fileExt string, r io.Reader) ([]byte, error) {
	f, err := os.CreateTemp("", filePrefix+"-*."+fileExt)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	path := f.Name()
	if _, err := io.Copy(f, r); err != nil {
		os.Remove(path)
		return nil, err
	}
	f.Close()

	// launch editor
	args := getEditor()
	args = append(args, path)
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("failed to launch editor: %w", err)
		return nil, err
	}

	bytes, err := ioutil.ReadFile(path)
	os.Remove(path)
	return bytes, err
}

func EditYAML(content []byte, previousError error, filePrefix, fileExt string) ([]byte, error) {
	original := removeComments(content)
	buf := &bytes.Buffer{}
	if previousError != nil {
		addErrorHeaders(buf, previousError)
	}
	buf.Write(original)
	edited, err := editInEditor(filePrefix, fileExt, buf)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(removeComments(edited), removeComments(original)) {
		// cancel
		return nil, ErrEditorCancelled
	}
	return edited, nil
}

func addErrorHeaders(w io.Writer, err error) {
	addHashToLineBreak := func(input string) string {
		lines := strings.Split(input, "\n")
		return strings.Join(lines, "\n# ")
	}
	fmt.Fprintf(w, `# Error:
# %s
#
`, addHashToLineBreak(err.Error()))
}

func removeComments(b []byte) []byte {
	newLines := [][]byte{}
	lines := bytes.Split(b, []byte("\n"))
	for _, line := range lines {
		if !bytes.HasPrefix(line, []byte("#")) {
			newLines = append(newLines, line)
		}
	}
	return bytes.Join(newLines, []byte("\n"))
}
