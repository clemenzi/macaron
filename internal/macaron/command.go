package macaron

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func scriptCommand(script string) *exec.Cmd {
	info, err := os.Stat(script)
	if err == nil && info.Mode()&0o111 != 0 {
		return command(script)
	}
	return command("bash", script)
}

func serviceRoot(script string) string {
	return filepath.Dir(filepath.Dir(script))
}

func (a *app) runCheck(script, label string) error {
	a.log.script("╭─ %s logs", label)
	cmd := scriptCommand(script)
	cmd.Dir = serviceRoot(script)
	reader, writer, err := os.Pipe()
	if err != nil {
		return err
	}
	defer reader.Close()
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Start(); err != nil {
		writer.Close()
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
		writer.Close()
	}()
	lines := streamLimited(a.log, reader, maxOutputLines)
	err = <-done
	if lines == 0 {
		a.log.empty("│ no output from script")
	}
	a.log.script("╰─ logs end")
	return err
}

func streamLimited(log *logger, r io.Reader, limit int) int {
	reader := bufio.NewReader(r)
	count := 0
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			count++
			if count <= limit {
				log.script("│ %s", strings.TrimSuffix(line, "\n"))
			}
		}
		if err != nil {
			break
		}
	}
	return count
}

func runOutput(name string, args ...string) (string, error) {
	cmd := command(name, args...)
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
