package macaron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCLIHelpAndUsageErrors(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out, errOut bytes.Buffer
	if code := Run([]string{"help"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("help exit code = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "install") {
		t.Fatalf("help missing install command: %s", out.String())
	}
	out.Reset()
	errOut.Reset()
	if code := Run([]string{"unknown"}, strings.NewReader(""), &out, &errOut); code != 2 {
		t.Fatalf("unknown exit code = %d", code)
	}
	if !strings.Contains(errOut.String(), "Unknown command: unknown") {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
	out.Reset()
	errOut.Reset()
	if code := Run([]string{"doctor", "extra"}, strings.NewReader(""), &out, &errOut); code != 2 {
		t.Fatalf("invalid doctor exit code = %d", code)
	}
	for _, args := range [][]string{
		{"install"},
		{"install", "one", "two"},
		{"install", "--name"},
		{"install", "--unknown", "source"},
	} {
		out.Reset()
		errOut.Reset()
		if code := Run(args, strings.NewReader(""), &out, &errOut); code != 2 {
			t.Errorf("Run(%q) exit code = %d, stderr = %s", args, code, errOut.String())
		}
	}
}

func TestServiceListOutput(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("NO_COLOR", "1")
	mustMkdir(t, filepath.Join(base, "macaron", "services", "api"))
	mustMkdir(t, filepath.Join(base, "macaron", "services-disabled", "docs"))
	var out, errOut bytes.Buffer
	if code := Run([]string{"list"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("list exit code = %d, stderr = %s", code, errOut.String())
	}
	for _, expected := range []string{"📦 Services", "✅ api", "enabled", "*️⃣ docs", "disabled"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("output missing %q: %s", expected, out.String())
		}
	}
}

func TestServiceOutputIsDimmedOnTerminal(t *testing.T) {
	var out bytes.Buffer
	output := &terminalOutput{out: &out, err: io.Discard, useANSI: true}
	output.Service("api", "ready")
	if got := out.String(); got != "\x1b[2m🪵 api  ready\x1b[0m\n" {
		t.Fatalf("service output = %q", got)
	}
}

func TestInstallManageAndDeleteLocalService(t *testing.T) {
	configBase := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configBase)
	t.Setenv("NO_COLOR", "1")
	source := filepath.Join(t.TempDir(), "macaron-service-demo")
	mustMkdir(t, filepath.Join(source, ".macaron"))
	mustWrite(t, filepath.Join(source, ".hidden"), "kept", 0o644)
	mustWrite(t, filepath.Join(source, ".macaron", "build"), "pwd > build.cwd\nprintf 'one\\ntwo\\nthree\\nfour\\nfive\\n'\n", 0o644)
	mustWrite(t, filepath.Join(source, ".macaron", "doctor"), "test -f .hidden\n", 0o644)

	var out, errOut bytes.Buffer
	if code := Run([]string{"install", "--yes", source}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("install exit code = %d, stdout = %s, stderr = %s", code, out.String(), errOut.String())
	}
	service := filepath.Join(configBase, "macaron", "services", "demo")
	if got, err := os.ReadFile(filepath.Join(service, ".hidden")); err != nil || string(got) != "kept" {
		t.Fatalf("hidden file not copied: %q, %v", got, err)
	}
	workingDir, err := os.ReadFile(filepath.Join(service, "build.cwd"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(workingDir)) != service {
		t.Fatalf("build cwd = %q, want %q", workingDir, service)
	}
	if strings.Contains(out.String(), "five") {
		t.Fatalf("script output was not limited: %s", out.String())
	}

	if code := Run([]string{"disable", "demo"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("disable code = %d", code)
	}
	if !directory(filepath.Join(configBase, "macaron", "services-disabled", "demo")) {
		t.Fatal("service was not disabled")
	}
	if code := Run([]string{"enable", "demo"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("enable code = %d", code)
	}
	if code := Run([]string{"list"}, strings.NewReader(""), &out, &errOut); code != 0 || !strings.Contains(out.String(), "demo") {
		t.Fatalf("list failed: code %d, %s", code, out.String())
	}
	if code := Run([]string{"delete", "demo"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("delete code = %d", code)
	}
	if directory(service) {
		t.Fatal("service was not deleted")
	}
}

func TestDoctor(t *testing.T) {
	configBase := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configBase)
	t.Setenv("NO_COLOR", "1")
	bin := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	mustWrite(t, filepath.Join(bin, "tailscale"), "#!/bin/sh\nprintf '{\"BackendState\":\"Running\"}\\n'\n", 0o755)
	service := filepath.Join(configBase, "macaron", "services", "ok", ".macaron")
	mustMkdir(t, service)
	mustWrite(t, filepath.Join(service, "doctor"), "#!/bin/sh\necho healthy\n", 0o755)
	var out, errOut bytes.Buffer
	if code := Run([]string{"doctor"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("doctor code = %d, stderr = %s", code, errOut.String())
	}
	want := "✅ Tailscale OK\n✅ ok doctor passed\n"
	if out.String() != want {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestInstallAndUpdateGitService(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	configBase := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configBase)
	t.Setenv("NO_COLOR", "1")
	source := filepath.Join(t.TempDir(), "source")
	mustMkdir(t, filepath.Join(source, ".macaron"))
	runGit(t, source, "init", "-b", "main")
	runGit(t, source, "config", "user.email", "test@example.com")
	runGit(t, source, "config", "user.name", "Macaron Test")
	mustWrite(t, filepath.Join(source, "version"), "one\n", 0o644)
	mustWrite(t, filepath.Join(source, ".macaron", "build"), "#!/bin/sh\necho built >> builds\n", 0o755)
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial")

	var out, errOut bytes.Buffer
	url := "file://" + source
	if code := Run([]string{"install", "--branch", "main", "--yes", "--name", "git-service", url}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("git install code = %d, stdout = %s, stderr = %s", code, out.String(), errOut.String())
	}
	mustWrite(t, filepath.Join(source, "version"), "two\n", 0o644)
	runGit(t, source, "add", "version")
	runGit(t, source, "commit", "-m", "update")
	if code := Run([]string{"update"}, strings.NewReader(""), &out, &errOut); code != 0 {
		t.Fatalf("update code = %d, stderr = %s", code, errOut.String())
	}
	destination := filepath.Join(configBase, "macaron", "services", "git-service")
	version, err := os.ReadFile(filepath.Join(destination, "version"))
	if err != nil || string(version) != "two\n" {
		t.Fatalf("updated version = %q, err = %v", version, err)
	}
	builds, err := os.ReadFile(filepath.Join(destination, "builds"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(builds), "built") != 2 {
		t.Fatalf("build script runs = %q, want two", builds)
	}
}

func TestActiveServicesFile(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	a, err := newApp(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	want := []activeService{{Name: "quoted \\\" service", Port: 49001}}
	if err := a.writeActiveServices(want); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(a.active)
	if err != nil {
		t.Fatal(err)
	}
	var got []activeService
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid JSON: %v: %s", err, data)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("active services = %#v, want %#v", got, want)
	}
}

func TestPortAvailable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot open a test port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if portAvailable(port) {
		t.Fatalf("port %d reported available while listening", port)
	}
	listener.Close()
	if !portAvailable(port) {
		t.Fatalf("port %d reported unavailable after close", port)
	}
}

func TestStartLifecycleHelper(t *testing.T) {
	if os.Getenv("MACARON_TEST_START_HELPER") != "1" {
		return
	}
	a, err := newApp(os.Stdin, os.Stdout, os.Stderr)
	if err == nil {
		a.portFree = func(int) bool { return true }
		err = a.run([]string{"start"})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestStartLifecycleAndRestore(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signals and shell scripts are Unix-specific")
	}
	root := t.TempDir()
	configBase := filepath.Join(root, "config")
	bin := filepath.Join(root, "bin")
	mustMkdir(t, bin)
	calls := filepath.Join(root, "calls")
	cleanup := filepath.Join(root, "cleanup")
	mustWrite(t, filepath.Join(bin, "sudo"), `#!/bin/sh
echo "sudo $*" >> "$MACARON_TEST_CALLS"
case "$1 $2" in
  "-v ") exit 0 ;;
  "systemsetup -getremotelogin") echo "Remote Login: Off" ;;
esac
exit 0
`, 0o755)
	mustWrite(t, filepath.Join(bin, "pmset"), `#!/bin/sh
echo "pmset $*" >> "$MACARON_TEST_CALLS"
if [ "$1" = "-g" ]; then echo " SleepDisabled 0"; fi
`, 0o755)
	mustWrite(t, filepath.Join(bin, "tailscale"), `#!/bin/sh
echo "tailscale $*" >> "$MACARON_TEST_CALLS"
case "$1 $2" in
  "status --json") echo '{"BackendState":"Running"}' ;;
  "status ") exit 1 ;;
  "ip -4") echo '100.64.0.10' ;;
esac
exit 0
`, 0o755)
	service := filepath.Join(configBase, "macaron", "services", "web", ".macaron")
	mustMkdir(t, service)
	mustWrite(t, filepath.Join(service, "start"), `#!/bin/sh
trap 'exit 0' TERM INT
echo "port=$MACARON_AVAILABLE_PORT"
while :; do sleep 1; done
`, 0o755)
	mustWrite(t, filepath.Join(service, "cleanup"), fmt.Sprintf("#!/bin/sh\ntouch %q\n", cleanup), 0o755)
	failing := filepath.Join(configBase, "macaron", "services", "broken", ".macaron")
	mustMkdir(t, failing)
	mustWrite(t, filepath.Join(failing, "start"), "#!/bin/sh\nexit 7\n", 0o755)

	cmd := exec.Command(os.Args[0], "-test.run=^TestStartLifecycleHelper$")
	cmd.Env = append(os.Environ(),
		"MACARON_TEST_START_HELPER=1", "MACARON_TEST_CALLS="+calls, "XDG_CONFIG_HOME="+configBase,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "NO_COLOR=1", "USER=tester")
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	activeFile := filepath.Join(configBase, "macaron", "active-services.json")
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(activeFile); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(activeFile); err != nil {
		callData, _ := os.ReadFile(calls)
		t.Fatalf("active services not created: %v\noutput:\n%s\ncalls:\n%s", err, output.String(), callData)
	}
	data, err := os.ReadFile(activeFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"name":"web"`) || !strings.Contains(string(data), `"port":49002`) {
		t.Fatalf("unexpected active services: %s", data)
	}
	if strings.Contains(string(data), `"name":"broken"`) {
		t.Fatalf("failed service included in active services: %s", data)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("start helper failed: %v\n%s", err, output.String())
	}
	if _, err := os.Stat(activeFile); !os.IsNotExist(err) {
		t.Fatalf("active services file still exists: %v", err)
	}
	if _, err := os.Stat(cleanup); err != nil {
		t.Fatalf("cleanup script did not run: %v", err)
	}
	callData, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	callLog := string(callData)
	for _, expected := range []string{"sudo systemsetup -setremotelogin off", "sudo pmset -a disablesleep 0", "tailscale down"} {
		if !strings.Contains(callLog, expected) {
			t.Errorf("restore call %q missing from:\n%s", expected, callLog)
		}
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
func mustWrite(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = directory
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
func waitFor(t *testing.T, timeout time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}
