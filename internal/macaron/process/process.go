// Package process runs external commands and Macaron service scripts.
package process

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func Script(script string) *exec.Cmd {
	info, err := os.Stat(script)
	if err == nil && info.Mode()&0o111 != 0 {
		return Command(script)
	}
	return Command("bash", script)
}

func ServiceRoot(script string) string {
	return filepath.Dir(filepath.Dir(script))
}

func Output(name string, args ...string) (string, error) {
	cmd := Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		message := strings.TrimSpace(out.String())
		if message != "" {
			return out.String(), fmt.Errorf("%s: %w", message, err)
		}
	}
	return out.String(), err
}

func QuietScript(script string) error {
	cmd := Script(script)
	cmd.Dir = ServiceRoot(script)
	output, err := cmd.CombinedOutput()
	if err != nil && strings.TrimSpace(string(output)) != "" {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return err
}

func Attached(in io.Reader, out, errOut io.Writer, name string, args ...string) error {
	cmd := Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = in, out, errOut
	return cmd.Run()
}
