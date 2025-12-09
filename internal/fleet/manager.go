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

//
// OrganizationManager
//

type OrganizationInfo struct {
	ID            string
	Name          string
	Label         string
	Type          string
	CreatedAt     string
	UpdatedAt     string
	OwnerID       string
	OwnerEmail    string
	OwnerUsername string
}

type OrganizationManager struct {
	fleetManager *FleetManager
}

func NewOrganizationManager(fleetManager *FleetManager) *OrganizationManager {
	return &OrganizationManager{
		fleetManager: fleetManager,
	}
}

func (pm *OrganizationManager) List() ([]OrganizationInfo, error) {
	result := []OrganizationInfo{}
	args := []string{"organization:list", "--format=csv", "--columns=*"}
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		organization := OrganizationInfo{
			ID:            record["ID"],
			Name:          record["Name"],
			Label:         record["Label"],
			Type:          record["Type"],
			CreatedAt:     record["Created at"],
			UpdatedAt:     record["Updated at"],
			OwnerID:       record["Owner ID"],
			OwnerEmail:    record["Owner email"],
			OwnerUsername: record["Owner username"],
		}
		result = append(result, organization)
	}

	return result, nil
}

//
// ProjectManager
//

type ProjectInfo struct {
	ProjectID         string
	ProjectTitle      string
	Region            string
	OrganizationName  string
	OrganizationID    string
	OrganizationLabel string
	OrganizationType  string
	Status            string
	Created           string
}

type ProjectManager struct {
	fleetManager *FleetManager
}

func NewProjectManager(fleetManager *FleetManager) *ProjectManager {
	return &ProjectManager{
		fleetManager: fleetManager,
	}
}

func (pm *ProjectManager) List(organization *OrganizationInfo) ([]ProjectInfo, error) {
	result := []ProjectInfo{}
	args := []string{"project:list", "--format=csv", "--count=0", "--columns=*", "--org", organization.ID}
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		project := ProjectInfo{
			ProjectID:         record["ID"],
			ProjectTitle:      record["Title"],
			Region:            record["Region"],
			OrganizationName:  record["Org name"],
			OrganizationID:    record["Org ID"],
			OrganizationLabel: record["Org label"],
			OrganizationType:  record["Org type"],
			Status:            record["Status"],
			Created:           record["Created"],
		}
		result = append(result, project)
	}

	return result, nil
}
