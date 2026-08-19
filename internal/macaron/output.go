package macaron

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type logger struct {
	out, err   io.Writer
	white, dim string
	dimItalic  string
	reset      string
	mu         sync.Mutex
}

func newLogger(out, errOut io.Writer) *logger {
	l := &logger{out: out, err: errOut}
	file, ok := out.(*os.File)
	if !ok || os.Getenv("NO_COLOR") != "" {
		return l
	}
	info, err := file.Stat()
	if err == nil && info.Mode()&os.ModeCharDevice != 0 {
		l.white = "\x1b[37m"
		l.dim = "\x1b[2;90m"
		l.dimItalic = "\x1b[2;3;90m"
		l.reset = "\x1b[0m"
	}
	return l
}

func (l *logger) line(w io.Writer, color, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprint(w, color)
	fmt.Fprintf(w, format, args...)
	fmt.Fprintln(w, l.reset)
}

func (l *logger) info(format string, args ...any)   { l.line(l.out, l.white, format, args...) }
func (l *logger) warn(format string, args ...any)   { l.info(format, args...) }
func (l *logger) error(format string, args ...any)  { l.line(l.err, l.white, format, args...) }
func (l *logger) script(format string, args ...any) { l.line(l.out, l.dim, format, args...) }
func (l *logger) empty(format string, args ...any)  { l.line(l.out, l.dimItalic, format, args...) }
