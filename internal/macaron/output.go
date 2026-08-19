package macaron

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type terminalOutput struct {
	mu      sync.Mutex
	out     io.Writer
	err     io.Writer
	useANSI bool
}

func newTerminalOutput(out, errOut io.Writer) *terminalOutput {
	return &terminalOutput{out: out, err: errOut, useANSI: isTerminal(out) && os.Getenv("NO_COLOR") == ""}
}

func (o *terminalOutput) Info(format string, args ...any) {
	o.print(o.out, "🔘 ", format, args...)
}

func (o *terminalOutput) Success(format string, args ...any) {
	o.print(o.out, "🟢 ", format, args...)
}

func (o *terminalOutput) Warning(format string, args ...any) {
	o.print(o.out, "🟡  ", format, args...)
}

func (o *terminalOutput) Error(format string, args ...any) {
	o.print(o.err, "🔴 ", format, args...)
}

func (o *terminalOutput) Section(icon, title string) {
	o.print(o.out, "\n"+icon+" ", "%s", title)
}

func (o *terminalOutput) Service(name, line string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.useANSI {
		fmt.Fprintf(o.out, "\033[2m🪵 %s  %s\033[0m\n", name, line)
		return
	}
	fmt.Fprintf(o.out, "🪵 %s  %s\n", name, line)
}

func (o *terminalOutput) print(writer io.Writer, prefix, format string, args ...any) {
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
