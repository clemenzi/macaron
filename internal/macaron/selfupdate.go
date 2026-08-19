package macaron

import (
	"fmt"
	"os"

	"github.com/clemenzi/macaron/internal/macaron/process"
)

const installerURL = "https://raw.githubusercontent.com/clemenzi/macaron/refs/heads/main/install.sh"

func (a *app) selfUpdate() error {
	if err := process.Attached(a.in, a.out, a.err, "sudo", "-v"); err != nil {
		return err
	}
	a.output.Info("Downloading latest Macaron installer")
	file, err := os.CreateTemp("", "macaron-install-*.sh")
	if err != nil {
		return err
	}
	path := file.Name()
	defer os.Remove(path)
	if err := file.Close(); err != nil {
		return err
	}
	if err := process.Attached(a.in, a.out, a.err, "curl", "-fsSL", installerURL, "-o", path); err != nil {
		return fmt.Errorf("download installer: %w", err)
	}
	if err := process.Attached(a.in, a.out, a.err, "sudo", "bash", path); err != nil {
		return fmt.Errorf("run installer: %w", err)
	}
	a.output.Success("Macaron updated")
	return nil
}
