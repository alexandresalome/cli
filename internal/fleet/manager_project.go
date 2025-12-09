package fleet

import "github.com/sirupsen/logrus"

type ProjectInfo struct {
	ProjectID          string
	ProjectTitle       string
	Region             string
	Organization       *OrganizationInfo
	OrganizationName   string
	OrganizationID     string
	OrganizationLabel  string
	OrganizationType   string
	Status             string
	Created            string
	EnvironmentsLoaded bool
	Environments       []EnvironmentInfo
}

func (p *ProjectInfo) ProductionEnvironment() *EnvironmentInfo {
	for _, env := range p.Environments {
		if env.IsProduction() {
			return &env
		}
	}
	return nil
}

type ProjectManager struct {
	logger       *logrus.Entry
	fleetManager *FleetManager
	loaded       bool
	records      []ProjectInfo
}

func NewProjectManager(fleetManager *FleetManager) *ProjectManager {
	return &ProjectManager{
		fleetManager: fleetManager,
		logger:       fleetManager.rootLogger.WithField("component", "ProjectManager"),
		loaded:       false,
		records:      []ProjectInfo{},
	}
}

func (pm *ProjectManager) List(organization *OrganizationInfo) ([]ProjectInfo, error) {
	allProjects, err := pm.ListAll()
	if err != nil {
		return nil, err
	}

	if organization == nil {
		return allProjects, nil
	}

	filtered := []ProjectInfo{}
	for _, project := range allProjects {
		if project.OrganizationID == organization.ID {
			filtered = append(filtered, project)
		}
	}

	return filtered, nil
}

func (pm *ProjectManager) ListAll() ([]ProjectInfo, error) {
	if pm.loaded {
		return pm.records, nil
	}

	result := []ProjectInfo{}
	args := []string{"project:list", "--format=csv", "--count=0", "--columns=*"}

	pm.logger.Info("Loading projects")
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		org, err := pm.fleetManager.Organization.GetByID(record["Org ID"])
		if err != nil {
			pm.logger.Warnf("Failed to get organization %s: %v", record["Org ID"], err)
			continue
		}
		project := ProjectInfo{
			ProjectID:          record["ID"],
			ProjectTitle:       record["Title"],
			Region:             record["Region"],
			Organization:       org,
			OrganizationName:   record["Org name"],
			OrganizationID:     record["Org ID"],
			OrganizationLabel:  record["Org label"],
			OrganizationType:   record["Org type"],
			Status:             record["Status"],
			Created:            record["Created"],
			EnvironmentsLoaded: false,
			Environments:       []EnvironmentInfo{},
		}
		result = append(result, project)
	}

	pm.loaded = true
	pm.records = result
	pm.logger.Debugf("Loaded %d projects", len(pm.records))

	return result, nil
}

func (pm *ProjectManager) Subscribe(organization *OrganizationInfo) chan ProjectInfo {
	ch := make(chan ProjectInfo)

	go func() {
		projects, err := pm.List(organization)
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
			envs, err := pm.fleetManager.Environment.List(&project)
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
