package macaron

import (
	"fmt"
	"strings"
)

type systemState struct {
	remoteLogin     string
	sleepDisabled   bool
	tailscaleActive bool
}

func (a *app) readSystemState() (systemState, error) {
	remote, err := runOutput("sudo", "systemsetup", "-getremotelogin")
	if err != nil {
		return systemState{}, fmt.Errorf("read Remote Login state: %w", err)
	}
	fields := strings.Fields(strings.ToLower(remote))
	if len(fields) == 0 {
		return systemState{}, fmt.Errorf("read Remote Login state: unexpected output %q", strings.TrimSpace(remote))
	}
	state := systemState{remoteLogin: fields[len(fields)-1]}
	pmset, err := runOutput("pmset", "-g")
	if err != nil {
		return systemState{}, fmt.Errorf("read sleep state: %w", err)
	}
	for _, line := range strings.Split(pmset, "\n") {
		if strings.Contains(line, "SleepDisabled") && strings.Contains(line, "1") {
			state.sleepDisabled = true
		}
	}
	state.tailscaleActive = command("tailscale", "status").Run() == nil
	return state, nil
}

func (a *app) prepareSystem() error {
	if err := command("sudo", "systemsetup", "-setremotelogin", "on").Run(); err != nil {
		return fmt.Errorf("enable Remote Login: %w", err)
	}
	a.log.info("✅ Remote Login enabled")
	if err := runAttached(a.in, a.out, a.err, "sudo", "pmset", "-a", "disablesleep", "1"); err != nil {
		return fmt.Errorf("disable sleep: %w", err)
	}
	a.log.info("✅ Sleep disabled")
	if err := runAttached(a.in, a.out, a.err, "tailscale", "up"); err != nil {
		return fmt.Errorf("start Tailscale: %w", err)
	}
	a.log.info("✅ Tailscale started")
	return nil
}

func (a *app) restoreSystemState(state systemState) {
	a.log.info("🔙 Restoring previous settings...")
	if err := command("sudo", "systemsetup", "-setremotelogin", state.remoteLogin).Run(); err != nil {
		a.log.error("😵 Failed to restore Remote Login")
	}
	sleep := "0"
	if state.sleepDisabled {
		sleep = "1"
	}
	if err := runAttached(a.in, a.out, a.err, "sudo", "pmset", "-a", "disablesleep", sleep); err != nil {
		a.log.error("😵 Failed to restore sleep settings")
	}
	tailscale := "down"
	if state.tailscaleActive {
		tailscale = "up"
	}
	if err := runAttached(a.in, a.out, a.err, "tailscale", tailscale); err != nil {
		a.log.error("😵 Failed to restore Tailscale")
	}
}
