package macaron

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/clemenzi/macaron/internal/macaron/process"
	"github.com/clemenzi/macaron/internal/macaron/ui"
)

func (a *app) runCheck(script, label string) error {
	a.output.Info("Running %s", label)
	cmd := process.Script(script)
	cmd.Dir = process.ServiceRoot(script)
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
	lines := streamLimited(a.output, reader, label, maxOutputLines)
	err = <-done
	if lines == 0 {
		a.output.Info("%s produced no output", label)
	}
	if err != nil {
		a.output.Error("%s failed: %v", label, err)
	} else {
		a.output.Success("%s completed", label)
	}
	return err
}

func streamLimited(output *ui.Output, r io.Reader, label string, limit int) int {
	reader := bufio.NewReader(r)
	count := 0
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			count++
			if count <= limit {
				output.Service(label, strings.TrimSuffix(line, "\n"))
			}
		}
		if err != nil {
			break
		}
	}
	return count
}
