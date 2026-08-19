// Package ui formats Macaron's terminal output.
package ui

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// Output writes status and service messages to a terminal.
type Output struct {
	mu      sync.Mutex
	out     io.Writer
	err     io.Writer
	useANSI bool
}

// New creates an Output for the provided stdout and stderr streams.
func New(out, errOut io.Writer) *Output {
	return &Output{out: out, err: errOut, useANSI: isTerminal(out) && os.Getenv("NO_COLOR") == ""}
}

func (o *Output) Info(format string, args ...any) {
	o.print(o.out, "ℹ️ ", format, args...)
}

func (o *Output) Success(format string, args ...any) {
	o.print(o.out, "✅ ", format, args...)
}

func (o *Output) Warning(format string, args ...any) {
	o.print(o.out, "⚠️  ", format, args...)
}

func (o *Output) Error(format string, args ...any) {
	o.print(o.err, "❌ ", format, args...)
}

func (o *Output) Section(icon, title string) {
	o.print(o.out, "\n"+icon+" ", "%s", title)
}

func (o *Output) Service(name, line string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.useANSI {
		fmt.Fprintf(o.out, "\033[2m🪵 %s  %s\033[0m\n", name, line)
		return
	}
	fmt.Fprintf(o.out, "🪵 %s  %s\n", name, line)
}

func (o *Output) print(writer io.Writer, prefix, format string, args ...any) {
	o.mu.Lock()
	defer o.mu.Unlock()
	fmt.Fprint(writer, prefix)
	fmt.Fprintf(writer, format, args...)
	fmt.Fprintln(writer)
}

func isTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
