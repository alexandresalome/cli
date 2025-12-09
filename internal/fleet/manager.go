package fleet

import (
	"bytes"
	"io"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/legacy"
)

//
// Fleet manager
//

type FleetManager struct {
	config  *config.Config
	command *cobra.Command

	Organization *OrganizationManager
	Project      *ProjectManager
}

func NewFleetManager(cnf *config.Config, cmd *cobra.Command) *FleetManager {
	fleetManager := &FleetManager{
		config:  cnf,
		command: cmd,
	}

	fleetManager.Organization = NewOrganizationManager(fleetManager)
	fleetManager.Project = NewProjectManager(fleetManager)

	return fleetManager
}

// Check the authentication status.
// Returns (isAuthenticated, email, error)
func (m *FleetManager) Authentication() (bool, string, error) {
	arguments := []string{"auth:info", "email", "--no-auto-login"}

	output, err := m.GetExecOutput(arguments)
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

func (m *FleetManager) Exec(arguments []string) error {
	wrapper := m.createCliWrapperWithDefaultIO()
	if err := wrapper.Exec(m.command.Context(), arguments...); err != nil {
		return err
	}

	return nil
}

func (m *FleetManager) GetExecOutput(arguments []string) (string, error) {
	// buffers inmemory
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	wrapper := m.createCliWrapper(&stdout, &stderr, nil)

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

func (m *FleetManager) createCliWrapperWithDefaultIO() *legacy.CLIWrapper {
	return m.createCliWrapper(m.command.OutOrStdout(), m.command.ErrOrStderr(), m.command.InOrStdin())
}

func (m *FleetManager) createCliWrapper(stdout io.Writer, stderr io.Writer, stdin io.Reader) *legacy.CLIWrapper {
	return &legacy.CLIWrapper{
		Config:             m.config,
		Version:            config.Version,
		DisableInteraction: true,
		Stdout:             stdout,
		Stderr:             stderr,
		Stdin:              stdin,
	}
}
