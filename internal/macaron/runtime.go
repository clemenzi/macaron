package macaron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/clemenzi/macaron/internal/macaron/discover"
	"github.com/clemenzi/macaron/internal/macaron/process"
	"github.com/clemenzi/macaron/internal/macaron/service"
	"github.com/clemenzi/macaron/internal/macaron/ui"
)

const (
	firstServicePort   = 49001
	serviceStartupWait = time.Second
)

type activeService = service.Endpoint

type managedService struct {
	activeService
	cmd     *exec.Cmd
	done    chan struct{}
	waitErr error
}

func (a *app) start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := process.Attached(a.in, a.out, a.err, "sudo", "-v"); err != nil {
		return err
	}
	state, err := a.readSystemState()
	if err != nil {
		return err
	}
	var services []*managedService
	defer func() {
		a.stopServices(services)
		a.cleanupServices()
		if err := os.Remove(a.active); err != nil && !errors.Is(err, os.ErrNotExist) {
			a.output.Error("Failed to remove %s: %v", a.active, err)
		}
		a.restoreSystemState(state)
	}()
	if ctx.Err() != nil {
		return nil
	}

	if err := a.prepareSystem(); err != nil {
		return err
	}
	a.output.Success("Environment ready")
	a.output.Info("Loading services")
	ip, err := process.Output("tailscale", "ip", "-4")
	if err != nil {
		return fmt.Errorf("get Tailscale IP: %w", err)
	}
	tailscaleIP := strings.Fields(ip)
	if len(tailscaleIP) == 0 {
		return errors.New("Tailscale did not return an IPv4 address")
	}

	launched, hasStartScripts, err := a.launchServices()
	if err != nil {
		return err
	}
	services = launched

	time.Sleep(serviceStartupWait)
	active := a.runningServices(services)

	if err := a.writeActiveServices(active); err != nil {
		return err
	}
	if !hasStartScripts {
		a.output.Warning("No services found in %s", a.services)
	}

	a.serveDiscovery(ctx, active)

	a.output.Section("🚀", "Macaron is running")
	a.output.Info("SSH  ssh %s@%s", currentUsername(), tailscaleIP[0])
	a.output.Info("Discover  http://%s:%d", tailscaleIP[0], discover.Port)
	for _, service := range active {
		a.output.Info("%s  http://%s:%d", service.Name, tailscaleIP[0], service.Port)
	}

	<-ctx.Done()
	return nil
}

func (a *app) launchServices() ([]*managedService, bool, error) {
	dirs, err := service.Dirs(a.services)
	if err != nil {
		return nil, false, err
	}

	var services []*managedService
	port := firstServicePort
	hasStartScripts := false

	for _, dir := range dirs {
		script := filepath.Join(dir, ".macaron", "start")
		if !service.RegularFile(script) {
			continue
		}

		hasStartScripts = true
		for !a.portFree(port) {
			port++
		}

		name := filepath.Base(dir)
		running, err := a.launchService(name, script, port)
		if err != nil {
			a.output.Error("Failed to start %s: %v", name, err)
			port++
			continue
		}

		services = append(services, running)
		port++
	}

	return services, hasStartScripts, nil
}

func (a *app) runningServices(services []*managedService) []activeService {
	active := make([]activeService, 0, len(services))

	for _, running := range services {
		select {
		case <-running.done:
			if running.waitErr == nil {
				a.output.Error("%s exited immediately; .macaron/start must stay in the foreground", running.Name)
			} else {
				a.output.Error("%s failed during startup: %s", running.Name, exitDescription(running.waitErr))
			}
		default:
			a.output.Success("%s started on port %d", running.Name, running.Port)
			active = append(active, running.activeService)
		}
	}

	return active
}

func (a *app) launchService(name, script string, port int) (*managedService, error) {
	cmd := process.Script(script)
	cmd.Dir = process.ServiceRoot(script)
	cmd.Env = append(os.Environ(), "MACARON_AVAILABLE_PORT="+strconv.Itoa(port))
	stdout := newLineWriter(a.output, name)
	stderr := newLineWriter(a.output, name)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	running := &managedService{
		activeService: activeService{Name: name, Port: port},
		cmd:           cmd,
		done:          make(chan struct{}),
	}
	go func() {
		running.waitErr = cmd.Wait()
		stdout.flush()
		stderr.flush()
		close(running.done)
	}()

	return running, nil
}

type lineWriter struct {
	output  *ui.Output
	name    string
	mu      sync.Mutex
	pending bytes.Buffer
}

func newLineWriter(output *ui.Output, name string) *lineWriter {
	return &lineWriter{output: output, name: name}
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	w.pending.Write(p)
	for {
		line, err := w.pending.ReadString('\n')
		if err != nil {
			w.pending.WriteString(line)
			break
		}
		w.output.Service(w.name, strings.TrimSuffix(line, "\n"))
	}
	return n, nil
}

func (w *lineWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pending.Len() == 0 {
		return
	}
	w.output.Service(w.name, w.pending.String())
	w.pending.Reset()
}

func (a *app) stopServices(services []*managedService) {
	if len(services) == 0 {
		return
	}
	a.output.Info("Stopping %d services", len(services))
	for _, service := range services {
		select {
		case <-service.done:
		default:
			_ = service.cmd.Process.Signal(syscall.SIGTERM)
		}
	}
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for _, service := range services {
		select {
		case <-service.done:
		case <-deadline.C:
			for _, remaining := range services {
				select {
				case <-remaining.done:
				default:
					a.output.Warning("%s did not stop gracefully; killing process %d", remaining.Name, remaining.cmd.Process.Pid)
					_ = remaining.cmd.Process.Kill()
				}
			}
			for _, remaining := range services {
				select {
				case <-remaining.done:
				case <-time.After(time.Second):
				}
			}
			return
		}
	}
}

func (a *app) cleanupServices() {
	a.output.Info("Cleaning up services")
	dirs, err := service.Dirs(a.services)
	if err != nil {
		a.output.Error("Failed to list services: %v", err)
		return
	}
	for _, dir := range dirs {
		script := filepath.Join(dir, ".macaron", "cleanup")
		if !service.RegularFile(script) {
			continue
		}
		name := filepath.Base(dir)
		if err := a.runCheck(script, name+" · cleanup"); err != nil {
			a.output.Error("Cleanup failed for %s: %v", name, err)
		} else {
			a.output.Success("Cleaned up %s", name)
		}
	}
}

func (a *app) writeActiveServices(services []activeService) error {
	if err := os.MkdirAll(a.config, 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(a.config, ".active-services.*")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(services); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(0o644); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, a.active); err != nil {
		return fmt.Errorf("replace active services file: %w", err)
	}
	return nil
}

func (a *app) serveDiscovery(ctx context.Context, services []activeService) {
	go func() {
		if err := discover.Serve(ctx, services); err != nil {
			a.output.Error("Discovery service failed: %v", err)
		}
	}()
}

func portAvailable(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func exitDescription(err error) string {
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return fmt.Sprintf("exit %d", exit.ExitCode())
	}
	return err.Error()
}

func currentUsername() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	return "user"
}
