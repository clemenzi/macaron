package macaron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
)

type activeService struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type managedService struct {
	activeService
	cmd     *exec.Cmd
	done    chan struct{}
	waitErr error
}

func (a *app) start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runAttached(a.in, a.out, a.err, "sudo", "-v"); err != nil {
		return err
	}
	state, err := a.readSystemState()
	if err != nil {
		return err
	}
	var services []*managedService
	defer func() {
		fmt.Fprintln(a.out)
		a.stopServices(services)
		a.cleanupServices()
		if err := os.Remove(a.active); err != nil && !errors.Is(err, os.ErrNotExist) {
			a.log.error("😵 Failed to remove active services file: %v", err)
		}
		a.restoreSystemState(state)
	}()
	if ctx.Err() != nil {
		return nil
	}

	if err := a.prepareSystem(); err != nil {
		return err
	}
	fmt.Fprintln(a.out)
	a.log.info("🙈 Environment ready; loading services...")
	ip, err := runOutput("tailscale", "ip", "-4")
	if err != nil {
		return fmt.Errorf("get Tailscale IP: %w", err)
	}
	tailscaleIP := strings.Fields(ip)
	if len(tailscaleIP) == 0 {
		return errors.New("Tailscale did not return an IPv4 address")
	}

	dirs, err := serviceDirs(a.services)
	if err != nil {
		return err
	}
	port := 49001
	found := false
	for _, dir := range dirs {
		script := filepath.Join(dir, ".macaron", "start")
		if !regularFile(script) {
			continue
		}
		found = true
		for !a.portFree(port) {
			port++
		}
		service, err := a.launchService(filepath.Base(dir), script, port)
		if err != nil {
			a.log.error("😵 Failed to start %s: %v", filepath.Base(dir), err)
			port++
			continue
		}
		services = append(services, service)
		port++
	}

	time.Sleep(time.Second)
	active := make([]activeService, 0, len(services))
	for _, service := range services {
		select {
		case <-service.done:
			if service.waitErr == nil {
				a.log.error("😵 %s exited immediately; .macaron/start must keep the service in the foreground", service.Name)
			} else {
				a.log.error("😵 Failed to start %s (%s)", service.Name, exitDescription(service.waitErr))
			}
		default:
			a.log.info("✅ %s started", service.Name)
			active = append(active, service.activeService)
		}
	}
	if err := a.writeActiveServices(active); err != nil {
		return err
	}
	if !found {
		a.log.warn("⚠️  No services found in %s", a.services)
	}
	fmt.Fprintln(a.out)
	a.log.info("🏃 Macaron is running")
	a.log.info("   → SSH: ssh %s@%s", currentUsername(), tailscaleIP[0])
	for _, service := range active {
		a.log.info("   -> %s: http://%s:%d", service.Name, tailscaleIP[0], service.Port)
	}

	<-ctx.Done()
	return nil
}

func (a *app) launchService(name, script string, port int) (*managedService, error) {
	cmd := scriptCommand(script)
	cmd.Dir = serviceRoot(script)
	cmd.Env = append(os.Environ(), "MACARON_AVAILABLE_PORT="+strconv.Itoa(port))
	stdout := newLineWriter(a.log, name)
	stderr := newLineWriter(a.log, name)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	service := &managedService{activeService: activeService{Name: name, Port: port}, cmd: cmd, done: make(chan struct{})}
	go func() {
		service.waitErr = cmd.Wait()
		stdout.flush()
		stderr.flush()
		close(service.done)
	}()
	return service, nil
}

type lineWriter struct {
	log     *logger
	name    string
	mu      sync.Mutex
	pending bytes.Buffer
}

func newLineWriter(log *logger, name string) *lineWriter { return &lineWriter{log: log, name: name} }
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
		w.log.script("🪵 (%s) LOG %s", w.name, strings.TrimSuffix(line, "\n"))
	}
	return n, nil
}

func (w *lineWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pending.Len() == 0 {
		return
	}
	w.log.script("🪵 (%s) LOG %s", w.name, w.pending.String())
	w.pending.Reset()
}

func (a *app) stopServices(services []*managedService) {
	if len(services) == 0 {
		return
	}
	a.log.info("🛑 Stopping services...")
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
					a.log.warn("⚠️  Service process %d did not stop gracefully; stopping it", remaining.cmd.Process.Pid)
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
	a.log.info("🧹 Cleaning up services...")
	dirs, err := serviceDirs(a.services)
	if err != nil {
		a.log.error("😵 Failed to list services: %v", err)
		return
	}
	for _, dir := range dirs {
		script := filepath.Join(dir, ".macaron", "cleanup")
		if !regularFile(script) {
			continue
		}
		name := filepath.Base(dir)
		if err := a.runCheck(script, name+" · cleanup"); err != nil {
			a.log.error("😵 Failed to clean up %s", name)
		} else {
			a.log.info("✅ %s cleaned up", name)
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
func runAttached(in io.Reader, out, errOut io.Writer, name string, args ...string) error {
	cmd := command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = in, out, errOut
	return cmd.Run()
}
