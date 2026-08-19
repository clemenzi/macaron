package macaron

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/clemenzi/macaron/internal/macaron/process"
	servicefs "github.com/clemenzi/macaron/internal/macaron/service"
	"github.com/urfave/cli/v3"
)

type installOptions struct {
	source, name, branch  string
	skipBuild, skipDoctor bool
	yes                   bool
}

func (a *app) install(opts installOptions) error {
	if opts.name == "" {
		opts.name = strings.TrimSuffix(filepath.Base(filepath.Clean(opts.source)), ".git")
	}
	a.output.Info("Installing service from %s", opts.source)
	if strings.HasPrefix(opts.name, "macaron-service-") {
		opts.name = strings.TrimPrefix(opts.name, "macaron-service-")
	}
	if !servicefs.ValidName(opts.name) {
		return usageError("Invalid service name: " + opts.name)
	}

	if err := os.MkdirAll(a.services, 0o755); err != nil {
		return fmt.Errorf("create services directory: %w", err)
	}
	destination := filepath.Join(a.services, opts.name)
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("cannot install %s: a service with that name already exists", opts.name)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	info, statErr := os.Stat(opts.source)
	if statErr == nil && info.IsDir() {
		if opts.branch != "" {
			return usageError("--branch cannot be used with a local directory")
		}
		if err := servicefs.CopyDir(opts.source, destination); err != nil {
			return fmt.Errorf("copy service: %w", err)
		}
	} else {
		cloneArgs := []string{"clone", "--quiet"}
		if opts.branch != "" {
			cloneArgs = append(cloneArgs, "--branch", opts.branch)
		}
		cloneArgs = append(cloneArgs, opts.source, destination)
		cmd := process.Command("git", cloneArgs...)
		cmd.Stdout, cmd.Stderr = a.out, a.err
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("clone service: %w", err)
		}
	}

	build := filepath.Join(destination, ".macaron", "build")
	if opts.skipBuild {
		a.output.Info("Build skipped for %s", opts.name)
	} else if servicefs.RegularFile(build) {
		run := opts.yes
		if !opts.yes {
			fmt.Fprintf(a.out, "%s has a build script. Run it? [y/N] ", opts.name)
			answer, _ := bufio.NewReader(a.in).ReadString('\n')
			run = strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y")
		}
		if run {
			a.output.Info("Building %s", opts.name)
			if err := a.runCheck(build, opts.name+" · build"); err != nil {
				return fmt.Errorf("build %s: %w", opts.name, err)
			}
		} else {
			a.output.Warning("Build skipped for %s; the service may not work correctly", opts.name)
		}
	}

	doctor := filepath.Join(destination, ".macaron", "doctor")
	if opts.skipDoctor {
		a.output.Info("Doctor skipped for %s", opts.name)
	} else if servicefs.RegularFile(doctor) {
		if err := a.runCheck(doctor, opts.name+" · doctor"); err != nil {
			a.output.Error("Doctor failed for %s; run 'macaron doctor' for details", opts.name)
		} else {
			a.output.Success("Doctor passed for %s", opts.name)
		}
	}
	a.output.Success("Service %s installed", opts.name)
	return nil
}

func (a *app) moveService(args []string, enable bool) error {
	verb := "disable"
	if enable {
		verb = "enable"
	}
	if len(args) != 1 {
		return usageError("Usage: macaron " + verb + " <service>")
	}
	name := filepath.Base(args[0])
	from, to := a.services, a.disabled
	state, done := "enabled", "disabled"
	if enable {
		from, to, state, done = to, from, done, state
	}
	source, destination := filepath.Join(from, name), filepath.Join(to, name)
	if !servicefs.Directory(source) {
		a.output.Info("Service %s is already %s", name, state)
		return nil
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("cannot %s %s: a %s service with that name already exists", verb, name, done)
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, destination); err != nil {
		return err
	}
	a.output.Success("Service %s is now %s", name, done)
	return nil
}

func (a *app) delete(args []string) error {
	if len(args) != 1 {
		return usageError("Usage: macaron delete <service>")
	}
	name := filepath.Base(args[0])
	enabled, disabled := filepath.Join(a.services, name), filepath.Join(a.disabled, name)
	if servicefs.Directory(enabled) && servicefs.Directory(disabled) {
		return fmt.Errorf("cannot delete %s: it exists in both enabled and disabled services", name)
	}
	a.output.Info("Deleting service %s", name)
	target := enabled
	if !servicefs.Directory(target) {
		target = disabled
	}
	if !servicefs.Directory(target) {
		a.output.Info("Service %s is not installed", name)
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	a.output.Success("Service %s deleted", name)
	return nil
}

func (a *app) update() error {
	a.output.Info("Updating services")
	services, err := servicefs.Dirs(a.services)
	if err != nil {
		return err
	}
	for _, dir := range services {
		name := filepath.Base(dir)
		if cmd := process.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree"); cmd.Run() != nil {
			a.output.Info("Skipping %s because its source is a local directory", name)
			continue
		}
		cmd := process.Command("git", "-C", dir, "pull", "--quiet")
		cmd.Stdout, cmd.Stderr = a.out, a.err
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("update %s: %w", name, err)
		}
		revision, err := process.Output("git", "-C", dir, "rev-parse", "--short", "HEAD")
		if err != nil {
			return fmt.Errorf("read %s revision: %w", name, err)
		}
		a.output.Success("Service %s updated to %s", name, strings.TrimSpace(revision))
		build := filepath.Join(dir, ".macaron", "build")
		if servicefs.RegularFile(build) {
			a.output.Info("Building %s", name)
			if err := a.runCheck(build, name+" · build"); err != nil {
				return fmt.Errorf("build %s: %w", name, err)
			}
		}
	}
	a.output.Success("Updated %d services", len(services))
	return nil
}

func (a *app) list() error {
	enabled, err := servicefs.Dirs(a.services)
	if err != nil {
		return err
	}
	disabled, err := servicefs.Dirs(a.disabled)
	if err != nil {
		return err
	}
	a.output.Section("📦", "Services")
	for _, dir := range enabled {
		a.output.Success("%s  enabled", filepath.Base(dir))
	}
	for _, dir := range disabled {
		a.output.Info("%s  disabled", filepath.Base(dir))
	}
	if len(enabled)+len(disabled) == 0 {
		a.output.Info("No services installed")
	}
	return nil
}

func (a *app) doctor() error {
	failed := false
	path, err := exec.LookPath("tailscale")
	if err != nil {
		a.output.Error("Tailscale is not installed or not on PATH")
		failed = true
	} else {
		output, err := process.Output(path, "status", "--json")
		if err != nil {
			a.output.Error("Unable to determine Tailscale login status: %v", err)
			failed = true
		} else {
			var status struct {
				BackendState string `json:"BackendState"`
			}
			if json.Unmarshal([]byte(output), &status) == nil && status.BackendState == "NeedsLogin" {
				a.output.Error("Tailscale is not logged in")
				failed = true
			} else {
				a.output.Success("Tailscale OK")
			}
		}
	}
	services, err := servicefs.Dirs(a.services)
	if err != nil {
		return err
	}
	found := false
	for _, dir := range services {
		check := filepath.Join(dir, ".macaron", "doctor")
		if !servicefs.RegularFile(check) {
			continue
		}
		found = true
		name := filepath.Base(dir)
		if err := process.QuietScript(check); err != nil {
			a.output.Error("%s doctor failed: %v", name, err)
			failed = true
		} else {
			a.output.Success("%s doctor passed", name)
		}
	}
	if len(services) == 0 {
		a.output.Info("No services configured in %s", a.services)
	} else if !found {
		a.output.Info("No service doctor checks found")
	}
	if failed {
		return cli.Exit("", 1)
	}
	return nil
}
