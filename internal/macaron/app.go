package macaron

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
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
	output   *terminalOutput
	portFree func(int) bool
}

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
	var exit cli.ExitCoder
	if errors.As(err, &exit) {
		if exit.Error() != "" {
			a.output.Error("%s", exit.Error())
		}
		return exit.ExitCode()
	}
	a.output.Error("Command failed: %v", err)
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
		output:   newTerminalOutput(out, errOut),
		portFree: portAvailable,
	}, nil
}

func (a *app) run(args []string) error {
	return a.cli().Run(context.Background(), append([]string{"macaron"}, args...))
}

func usageError(message string) error {
	return cli.Exit(message, 2)
}

func (a *app) cli() *cli.Command {
	noArgs := func(usage string, action func() error) cli.ActionFunc {
		return func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() != 0 {
				return usageError(usage)
			}
			return action()
		}
	}
	oneService := func(verb string, action func([]string) error) cli.ActionFunc {
		return func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() != 1 {
				return usageError("Usage: macaron " + verb + " <service>")
			}
			return action(cmd.Args().Slice())
		}
	}

	root := &cli.Command{
		Name:                   "macaron",
		Usage:                  "Turn your Mac into a remotely accessible workstation",
		Reader:                 a.in,
		Writer:                 a.out,
		ErrWriter:              a.err,
		HideVersion:            true,
		UseShortOptionHandling: true,
		ExitErrHandler:         func(context.Context, *cli.Command, error) {},
		OnUsageError:           usageErrorHandler,
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return a.start()
			}
			return cli.Exit("Unknown command: "+cmd.Args().First(), 2)
		},
	}
	root.Commands = []*cli.Command{
		{
			Name:   "start",
			Usage:  "Start macaron",
			Action: noArgs("Usage: macaron start", a.start),
		},
		{
			Name:   "doctor",
			Usage:  "Check the configuration and services",
			Action: noArgs("Usage: macaron doctor", a.doctor),
		},
		{
			Name:      "install",
			Usage:     "Install a service from a Git repository or local directory",
			ArgsUsage: "<source>",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Usage: "Name of the service directory"},
				&cli.StringFlag{Name: "branch", Usage: "Clone a specific branch"},
				&cli.BoolFlag{Name: "skip-build", Usage: "Do not run the service build script"},
				&cli.BoolFlag{Name: "skip-doctor", Usage: "Do not run the service doctor"},
				&cli.BoolFlag{
					Name:    "yes",
					Aliases: []string{"y"},
					Usage:   "Automatically confirm the build script",
				},
			},
			Action: func(_ context.Context, cmd *cli.Command) error {
				if cmd.Args().Len() != 1 {
					return usageError("Usage: macaron install [options] <source>")
				}
				return a.install(installOptions{
					source:     cmd.Args().First(),
					name:       cmd.String("name"),
					branch:     cmd.String("branch"),
					skipBuild:  cmd.Bool("skip-build"),
					skipDoctor: cmd.Bool("skip-doctor"),
					yes:        cmd.Bool("yes"),
				})
			},
		},
		{
			Name:   "update",
			Usage:  "Update all services",
			Action: noArgs("Usage: macaron update", a.update),
		},
		{
			Name:      "disable",
			Usage:     "Disable a service",
			ArgsUsage: "<service>",
			Action: oneService("disable", func(args []string) error {
				return a.moveService(args, false)
			}),
		},
		{
			Name:      "enable",
			Usage:     "Enable a service",
			ArgsUsage: "<service>",
			Action: oneService("enable", func(args []string) error {
				return a.moveService(args, true)
			}),
		},
		{
			Name:      "delete",
			Usage:     "Delete a service",
			ArgsUsage: "<service>",
			Action:    oneService("delete", a.delete),
		},
		{
			Name:   "list",
			Usage:  "List all services",
			Action: noArgs("Usage: macaron list", a.list),
		},
		{
			Name:   "self-update",
			Usage:  "Update macaron to the latest version",
			Action: noArgs("Usage: macaron self-update", a.selfUpdate),
		},
	}
	for _, command := range root.Commands {
		command.OnUsageError = usageErrorHandler
	}
	return root
}

func usageErrorHandler(_ context.Context, _ *cli.Command, err error, _ bool) error {
	return cli.Exit(err, 2)
}
