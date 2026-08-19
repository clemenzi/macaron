package macaron

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const maxOutputLines = 4

type app struct {
	in       io.Reader
	out      io.Writer
	err      io.Writer
	config   string
	services string
	disabled string
	active   string
	log      *logger
	portFree func(int) bool
}

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}
func (e *exitError) Unwrap() error { return e.err }

// Run executes Macaron and returns a process exit code.
func Run(args []string, in io.Reader, out, errOut io.Writer) int {
	a, err := newApp(in, out, errOut)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}

	err = a.run(args)
	if err == nil {
		return 0
	}
	var exit *exitError
	if errors.As(err, &exit) {
		if exit.err != nil {
			fmt.Fprintln(errOut, exit.err)
		}
		return exit.code
	}
	fmt.Fprintln(errOut, err)
	return 1
}

func newApp(in io.Reader, out, errOut io.Writer) (*app, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("find home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	config := filepath.Join(base, "macaron")
	return &app{
		in:       in,
		out:      out,
		err:      errOut,
		config:   config,
		services: filepath.Join(config, "services"),
		disabled: filepath.Join(config, "services-disabled"),
		active:   filepath.Join(config, "active-services.json"),
		log:      newLogger(out, errOut),
		portFree: portAvailable,
	}, nil
}

func (a *app) run(args []string) error {
	command := "start"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}

	switch command {
	case "start":
		if len(args) != 0 {
			return usageError("Usage: macaron start")
		}
		return a.start()
	case "doctor":
		if len(args) != 0 {
			return usageError("Usage: macaron doctor")
		}
		return a.doctor()
	case "install":
		return a.install(args)
	case "update":
		if len(args) != 0 {
			return usageError("Usage: macaron update")
		}
		return a.update()
	case "disable":
		return a.moveService(args, false)
	case "enable":
		return a.moveService(args, true)
	case "delete":
		return a.delete(args)
	case "list":
		if len(args) != 0 {
			return usageError("Usage: macaron list")
		}
		return a.list()
	case "self-update":
		if len(args) != 0 {
			return usageError("Usage: macaron self-update")
		}
		return a.selfUpdate()
	case "help", "--help", "-h":
		if len(args) != 0 {
			return usageError("Usage: macaron help")
		}
		fmt.Fprint(a.out, usage)
		return nil
	default:
		fmt.Fprintf(a.err, "Unknown command: %s\n%s", command, usage)
		return &exitError{code: 2}
	}
}

func usageError(message string) error {
	return &exitError{code: 2, err: errors.New(message)}
}

const usage = `Usage:
  macaron [start]                        | Start macaron
  macaron doctor                         | Check if everything is correctly configured and working
  macaron install [options] <source>     | Install a service from a Git repository or local directory
  macaron update                         | Update all services
  macaron disable <service>              | Disable a service
  macaron enable <service>               | Enable a service
  macaron delete <service>               | Delete a service
  macaron list                           | List all services
  macaron help                           | Show this help
  macaron self-update                    | Update macaron to the latest version
`

const installUsage = `Usage: macaron install [options] <source>

Options:
  --name <name>       Name of the service directory
  --branch <branch>   Clone a specific branch (Git repositories only)
  --skip-build        Do not run the service build script
  --skip-doctor       Do not run the service doctor
  -y, --yes           Automatically confirm the build script
  -h, --help          Show this help
`
