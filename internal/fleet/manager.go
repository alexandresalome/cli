package fleet

import (
	"bytes"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/legacy"
)

type FleetManager struct {
	config  *config.Config
	command *cobra.Command
}

func NewFleetManager(cnf *config.Config, cmd *cobra.Command) *FleetManager {
	return &FleetManager{
		config:  cnf,
		command: cmd,
	}
}

// Check the authentication status.
// Returns (isAuthenticated, email, error)
func (m *FleetManager) Authentication() (bool, string, error) {
	arguments := []string{"auth:info", "email", "--no-auto-login"}

	output, err := m.getExecOutput(arguments)
	if err != nil {
		return false, "", err
	}

	// strip whitespace
	stripped := bytes.TrimSpace([]byte(output))
	if len(stripped) == 0 {
		return false, "", nil
	}

	return true, string(stripped), nil
}

func (m *FleetManager) getExecOutput(arguments []string) (string, error) {
	// buffers inmemory
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	wrapper := &legacy.CLIWrapper{
		Config:             m.config,
		Version:            config.Version,
		DisableInteraction: true,
		Stdout:             &stdout,
		Stderr:             &stderr,
		Stdin:              nil,
	}

	if err := wrapper.Exec(m.command.Context(), arguments...); err != nil {
		newError := err
		if stderr.Len() > 0 {
			newError = &FleetManagerProcessError{
				Inner:      err,
				StderrData: stderr.String(),
			}
		}
		return "", newError
	}

	return stdout.String(), nil
}

func (m *FleetManager) Exec(arguments []string) error {
	wrapper := &legacy.CLIWrapper{
		Config:             m.config,
		Version:            config.Version,
		DisableInteraction: true,
		Stdout:             m.command.OutOrStdout(),
		Stderr:             m.command.ErrOrStderr(),
		Stdin:              m.command.InOrStdin(),
	}

	if err := wrapper.Exec(m.command.Context(), arguments...); err != nil {
		return err
	}

	return nil
}
