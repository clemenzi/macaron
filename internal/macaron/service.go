package macaron

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type installOptions struct {
	source, name, branch  string
	skipBuild, skipDoctor bool
	yes                   bool
}

func (a *app) install(args []string) error {
	opts, help, err := parseInstall(args)
	if help {
		fmt.Fprint(a.out, installUsage)
		return nil
	}
	if err != nil {
		fmt.Fprintln(a.err, err)
		fmt.Fprint(a.err, installUsage)
		var exit *exitError
		if errors.As(err, &exit) {
			return &exitError{code: exit.code}
		}
		return err
	}

	if opts.name == "" {
		opts.name = strings.TrimSuffix(filepath.Base(filepath.Clean(opts.source)), ".git")
	}
	a.log.info("📦 Installing from %s...", opts.source)
	if strings.HasPrefix(opts.name, "macaron-service-") {
		opts.name = strings.TrimPrefix(opts.name, "macaron-service-")
		a.log.info("ℹ️  Stripped macaron-service- prefix from name")
	}
	if !validServiceName(opts.name) {
		return usageError("Invalid service name: " + opts.name)
	}

	if err := os.MkdirAll(a.services, 0o755); err != nil {
		return fmt.Errorf("create services directory: %w", err)
	}
	destination := filepath.Join(a.services, opts.name)
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("😵 Cannot install %s: a service with that name already exists", opts.name)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	info, statErr := os.Stat(opts.source)
	if statErr == nil && info.IsDir() {
		if opts.branch != "" {
			return usageError("😵 --branch cannot be used with a local directory")
		}
		if err := copyDir(opts.source, destination); err != nil {
			return fmt.Errorf("copy service: %w", err)
		}
	} else {
		cloneArgs := []string{"clone", "--quiet"}
		if opts.branch != "" {
			cloneArgs = append(cloneArgs, "--branch", opts.branch)
		}
		cloneArgs = append(cloneArgs, opts.source, destination)
		cmd := command("git", cloneArgs...)
		cmd.Stdout, cmd.Stderr = a.out, a.err
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("clone service: %w", err)
		}
	}

	build := filepath.Join(destination, ".macaron", "build")
	if opts.skipBuild {
		a.log.info("ℹ️  Build skipped")
	} else if regularFile(build) {
		run := opts.yes
		if !opts.yes {
			fmt.Fprintf(a.out, "ℹ️  %s has a build script. Run it? [y/N] ", opts.name)
			answer, _ := bufio.NewReader(a.in).ReadString('\n')
			run = strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y")
		}
		if run {
			a.log.info("🔨 Building %s...", opts.name)
			if err := a.runCheck(build, opts.name+" · build"); err != nil {
				return fmt.Errorf("build %s: %w", opts.name, err)
			}
		} else {
			a.log.warn("⚠️  Build skipped; the service may not work correctly")
		}
	}

	doctor := filepath.Join(destination, ".macaron", "doctor")
	if opts.skipDoctor {
		a.log.info("ℹ️  Doctor skipped")
	} else if regularFile(doctor) {
		if err := a.runCheck(doctor, opts.name+" · doctor"); err != nil {
			a.log.error("😵 Doctor failed; run 'macaron doctor' for details")
		} else {
			a.log.info("✅ Doctor passed")
		}
	}
	a.log.info("✅ Installed from %s", opts.source)
	return nil
}

func parseInstall(args []string) (installOptions, bool, error) {
	var opts installOptions
	for len(args) > 0 {
		switch args[0] {
		case "--name", "--branch":
			flag := args[0]
			if len(args) < 2 {
				return opts, false, usageError("Missing value for " + flag)
			}
			if flag == "--name" {
				opts.name = args[1]
			} else {
				opts.branch = args[1]
			}
			args = args[2:]
		case "--skip-build":
			opts.skipBuild = true
			args = args[1:]
		case "--skip-doctor":
			opts.skipDoctor = true
			args = args[1:]
		case "-y", "--yes":
			opts.yes = true
			args = args[1:]
		case "-h", "--help":
			return opts, true, nil
		case "--":
			args = args[1:]
			if len(args) == 0 {
				return opts, false, usageError("Missing source")
			}
			if opts.source != "" || len(args) != 1 {
				return opts, false, usageError("Only one source can be installed at a time")
			}
			opts.source, args = args[0], nil
		default:
			if strings.HasPrefix(args[0], "-") {
				return opts, false, usageError("Unknown install option: " + args[0])
			}
			if opts.source != "" {
				return opts, false, usageError("Only one source can be installed at a time")
			}
			opts.source, args = args[0], args[1:]
		}
	}
	if opts.source == "" {
		return opts, false, usageError("Missing source")
	}
	return opts, false, nil
}

func validServiceName(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.Base(name) == name && !strings.ContainsAny(name, `/\\`)
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
	if !directory(source) {
		a.log.info("ℹ️  %s is not %s", name, state)
		return nil
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("😵 Cannot %s %s: a %s service with that name already exists", verb, name, done)
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, destination); err != nil {
		return err
	}
	a.log.info("✅ %s %sd", name, verb)
	return nil
}

func (a *app) delete(args []string) error {
	if len(args) != 1 {
		return usageError("Usage: macaron delete <service>")
	}
	name := filepath.Base(args[0])
	enabled, disabled := filepath.Join(a.services, name), filepath.Join(a.disabled, name)
	if directory(enabled) && directory(disabled) {
		return fmt.Errorf("😵 Cannot delete %s: it exists in both enabled and disabled services", name)
	}
	a.log.info("🗑️  Deleting %s...", name)
	target := enabled
	if !directory(target) {
		target = disabled
	}
	if !directory(target) {
		a.log.info("ℹ️  %s is not installed", name)
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	a.log.info("✅ %s deleted", name)
	return nil
}

func (a *app) update() error {
	a.log.info("🔄 Updating services...")
	services, err := serviceDirs(a.services)
	if err != nil {
		return err
	}
	for _, dir := range services {
		name := filepath.Base(dir)
		if cmd := command("git", "-C", dir, "rev-parse", "--is-inside-work-tree"); cmd.Run() != nil {
			a.log.info("ℹ️  %s was installed from a local directory; update skipped", name)
			continue
		}
		cmd := command("git", "-C", dir, "pull", "--quiet")
		cmd.Stdout, cmd.Stderr = a.out, a.err
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("update %s: %w", name, err)
		}
		revision, err := runOutput("git", "-C", dir, "rev-parse", "--short", "HEAD")
		if err != nil {
			return fmt.Errorf("read %s revision: %w", name, err)
		}
		a.log.info("✅ %s updated (%s)", name, strings.TrimSpace(revision))
		build := filepath.Join(dir, ".macaron", "build")
		if regularFile(build) {
			a.log.info("🔨 Building %s...", name)
			if err := a.runCheck(build, name+" · build"); err != nil {
				return fmt.Errorf("build %s: %w", name, err)
			}
		}
	}
	a.log.info("✅ Update complete")
	return nil
}

func (a *app) list() error {
	a.log.info("ℹ️  Installed services:")
	for _, dir := range mustServiceDirs(a.services) {
		a.log.info("   - %s", filepath.Base(dir))
	}
	a.log.info("ℹ️  Disabled services:")
	for _, dir := range mustServiceDirs(a.disabled) {
		a.log.info("   - %s", filepath.Base(dir))
	}
	return nil
}

func (a *app) doctor() error {
	failed := false
	path, err := exec.LookPath("tailscale")
	if err != nil {
		a.log.error("😵 Tailscale: not installed or not on PATH")
		failed = true
	} else {
		output, err := runOutput(path, "status", "--json")
		if err != nil {
			a.log.error("😵 Tailscale: unable to determine login status")
			failed = true
		} else {
			var status struct {
				BackendState string `json:"BackendState"`
			}
			if json.Unmarshal([]byte(output), &status) == nil && status.BackendState == "NeedsLogin" {
				a.log.error("😵 Tailscale: not logged in")
				failed = true
			} else {
				a.log.info("✅ Tailscale: logged in")
			}
		}
	}
	fmt.Fprintln(a.out)
	a.log.info("🔍 Checking services...")
	services, err := serviceDirs(a.services)
	if err != nil {
		return err
	}
	found := false
	for _, dir := range services {
		check := filepath.Join(dir, ".macaron", "doctor")
		if !regularFile(check) {
			continue
		}
		found = true
		name := filepath.Base(dir)
		if err := a.runCheck(check, name+" · doctor"); err != nil {
			a.log.error("  - %s: not ok", name)
			failed = true
		} else {
			a.log.info("  - %s: ok", name)
		}
	}
	if len(services) == 0 {
		a.log.info("  - no services configured (%s)", a.services)
	} else if !found {
		a.log.info("  - no service checks found")
	}
	fmt.Fprintln(a.out)
	if failed {
		a.log.error("😵 Doctor: one or more checks failed")
		return &exitError{code: 1}
	}
	a.log.info("✅ Doctor: all checks passed")
	return nil
}

func serviceDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(root, entry.Name()))
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func mustServiceDirs(root string) []string { dirs, _ := serviceDirs(root); return dirs }
func directory(path string) bool           { info, err := os.Stat(path); return err == nil && info.IsDir() }
func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func copyDir(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.Mkdir(destination, info.Mode().Perm()); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.Type()&os.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		case entry.IsDir():
			return os.Mkdir(target, info.Mode().Perm())
		case entry.Type().IsRegular():
			in, err := os.Open(path)
			if err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
			if err != nil {
				in.Close()
				return err
			}
			_, copyErr := io.Copy(out, in)
			readCloseErr := in.Close()
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if readCloseErr != nil {
				return readCloseErr
			}
			return closeErr
		default:
			return nil
		}
	})
}
