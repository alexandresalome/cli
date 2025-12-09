package fleet

//
// ProjectManager
//

type ProjectInfo struct {
	ProjectID          string
	ProjectTitle       string
	Region             string
	OrganizationName   string
	OrganizationID     string
	OrganizationLabel  string
	OrganizationType   string
	Status             string
	Created            string
	EnvironmentsLoaded bool
	Environments       []ProjectEnvironmentInfo
}

type ProjectEnvironmentInfo struct {
	ID          string
	MachineName string
	Title       string
	Status      string
	Type        string
	Created     string
	Updated     string
}

type ProjectManager struct {
	fleetManager *FleetManager
}

func NewProjectManager(fleetManager *FleetManager) *ProjectManager {
	return &ProjectManager{
		fleetManager: fleetManager,
	}
}

func (pm *ProjectManager) ListAll() ([]ProjectInfo, error) {
	return pm.List(nil)
}

func (pm *ProjectManager) List(organization *OrganizationInfo) ([]ProjectInfo, error) {
	result := []ProjectInfo{}
	args := []string{"project:list", "--format=csv", "--count=0", "--columns=*"}
	if organization != nil {
		args = append(args, "--org", organization.ID)
	}
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
			ProjectID:          record["ID"],
			ProjectTitle:       record["Title"],
			Region:             record["Region"],
			OrganizationName:   record["Org name"],
			OrganizationID:     record["Org ID"],
			OrganizationLabel:  record["Org label"],
			OrganizationType:   record["Org type"],
			Status:             record["Status"],
			Created:            record["Created"],
			EnvironmentsLoaded: false,
			Environments:       nil,
		}
		result = append(result, project)
	}

	return result, nil
}

func (pm *ProjectManager) GetEnvironments(projectID string) ([]ProjectEnvironmentInfo, error) {
	result := []ProjectEnvironmentInfo{}
	args := []string{"env", "--format=csv", "--project", projectID, "--columns=*"}
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		environment := ProjectEnvironmentInfo{
			ID:          record["ID"],
			MachineName: record["Machine name"],
			Title:       record["Title"],
			Status:      record["Status"],
			Type:        record["Type"],
			Created:     record["Created"],
			Updated:     record["Updated"],
		}

		result = append(result, environment)
	}

	return result, nil
}

func (pm *ProjectManager) Subscribe() chan ProjectInfo {
	ch := make(chan ProjectInfo)

	go func() {
		projects, err := pm.ListAll()
		defer func() {
			close(ch)
		}()

		if err != nil {
			return
		}

		// First, send all projects without environments
		for _, project := range projects {
			ch <- project
		}

		// Then, load and send environments for each project
		for _, project := range projects {
			envs, err := pm.GetEnvironments(project.ProjectID)
			if err == nil {
				project.Environments = envs
				project.EnvironmentsLoaded = true
			}
			ch <- project
		}

		close(ch)
	}()

	return ch
}

//
// OrganizationManager
//
