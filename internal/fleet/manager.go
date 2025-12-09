package fleet

import (
	"bytes"
	"io"
	"os"

	"github.com/sirupsen/logrus"
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

	rootLogger *logrus.Logger
	logger     *logrus.Entry

	Organization *OrganizationManager
	Project      *ProjectManager
}

func NewFleetManager(cnf *config.Config, cmd *cobra.Command) *FleetManager {
	output := cmd.OutOrStdout()
	if output == nil {
		output = os.Stdout
	}
	logger := logrus.New()
	if os.Getenv("INFO") == "true" {
		logger.SetLevel(logrus.InfoLevel)
	} else if os.Getenv("DEBUG") == "true" {
		logger.SetLevel(logrus.DebugLevel)
	} else if os.Getenv("TRACE") == "true" {
		logger.SetLevel(logrus.TraceLevel)
	} else {
		logger.SetLevel(logrus.WarnLevel)
	}
	logger.SetOutput(output)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	fleetManager := &FleetManager{
		config:     cnf,
		command:    cmd,
		rootLogger: logger,
		logger:     logger.WithField("component", "FleetManager"),
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

	stripped := bytes.TrimSpace([]byte(output))
	if len(stripped) == 0 {
		return false, "", nil
	}

	return true, string(stripped), nil
}

// Exec executes a command in the Fleet CLI context.
// It uses the default IO streams from the command.
func (m *FleetManager) Exec(arguments []string) error {
	wrapper := m.createCliWrapperWithDefaultIO()
	if err := wrapper.Exec(m.command.Context(), arguments...); err != nil {
		return err
	}

	return nil
}

// GetExecOutput executes a command in the Fleet CLI context and returns its
// output as a string.
func (m *FleetManager) GetExecOutput(arguments []string) (string, error) {
	// buffers inmemory
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	wrapper := m.createCliWrapper(&stdout, &stderr, nil)

	m.logger.Tracef("Executing command: %s", arguments)
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
