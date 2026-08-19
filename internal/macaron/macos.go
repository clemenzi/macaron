package macaron

import (
	"fmt"
	"strings"

	"github.com/clemenzi/macaron/internal/macaron/process"
)

type systemState struct {
	remoteLogin     string
	sleepDisabled   bool
	tailscaleActive bool
}

func (a *app) readSystemState() (systemState, error) {
	remote, err := process.Output("sudo", "systemsetup", "-getremotelogin")
	if err != nil {
		return systemState{}, fmt.Errorf("read Remote Login state: %w", err)
	}
	fields := strings.Fields(strings.ToLower(remote))
	if len(fields) == 0 {
		return systemState{}, fmt.Errorf("read Remote Login state: unexpected output %q", strings.TrimSpace(remote))
	}
	state := systemState{remoteLogin: fields[len(fields)-1]}
	pmset, err := process.Output("pmset", "-g")
	if err != nil {
		return systemState{}, fmt.Errorf("read sleep state: %w", err)
	}
	for _, line := range strings.Split(pmset, "\n") {
		if strings.Contains(line, "SleepDisabled") && strings.Contains(line, "1") {
			state.sleepDisabled = true
		}
	}
	state.tailscaleActive = process.Command("tailscale", "status").Run() == nil
	return state, nil
}

func (a *app) prepareSystem() error {
	if err := process.Command("sudo", "systemsetup", "-setremotelogin", "on").Run(); err != nil {
		return fmt.Errorf("enable Remote Login: %w", err)
	}
	a.output.Success("Remote Login enabled")
	if err := process.Attached(a.in, a.out, a.err, "sudo", "pmset", "-a", "disablesleep", "1"); err != nil {
		return fmt.Errorf("disable sleep: %w", err)
	}
	a.output.Success("Sleep disabled")
	if err := process.Attached(a.in, a.out, a.err, "tailscale", "up"); err != nil {
		return fmt.Errorf("start Tailscale: %w", err)
	}
	a.output.Success("Tailscale started")
	return nil
}

func (a *app) restoreSystemState(state systemState) {
	a.output.Info("Restoring previous system settings")
	if err := process.Command("sudo", "systemsetup", "-setremotelogin", state.remoteLogin).Run(); err != nil {
		a.output.Error("Failed to restore Remote Login: %v", err)
	}
	sleep := "0"
	if state.sleepDisabled {
		sleep = "1"
	}
	if err := process.Attached(a.in, a.out, a.err, "sudo", "pmset", "-a", "disablesleep", sleep); err != nil {
		a.output.Error("Failed to restore sleep settings: %v", err)
	}
	tailscale := "down"
	if state.tailscaleActive {
		tailscale = "up"
	}
	if err := process.Attached(a.in, a.out, a.err, "tailscale", tailscale); err != nil {
		a.output.Error("Failed to restore Tailscale: %v", err)
	}
}
