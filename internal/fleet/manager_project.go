package fleet

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type ProjectInfo struct {
	ProjectID          string            `json:"id"`
	ProjectTitle       string            `json:"title"`
	Region             string            `json:"region"`
	Organization       *OrganizationInfo `json:"organization"`
	OrganizationName   string
	OrganizationID     string
	OrganizationLabel  string
	OrganizationType   string
	Status             string       `json:"status"`
	Created            string       `json:"created_at"`
	DefaultEnvironment *Environment `json:"default_environment"`
}

func (p ProjectInfo) ToJson() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(bytes)
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

func (pm *ProjectManager) List(organization *OrganizationInfo, ctx context.Context) ([]ProjectInfo, error) {
	allProjects, err := pm.ListAll(ctx)
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

func (pm *ProjectManager) ListAll(ctx context.Context) ([]ProjectInfo, error) {
	if pm.loaded {
		return pm.records, nil
	}

	result := []ProjectInfo{}
	args := []string{"project:list", "--format=csv", "--count=0", "--columns=*"}

	pm.logger.Info("Loading projects")
	data, err := pm.fleetManager.GetExecOutputWithCtx(args, ctx)

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
			DefaultEnvironment: nil,
		}
		defaultEnv, err := pm.fleetManager.Environment.GetOrCreate(&project, ".")
		if err != nil {
			pm.logger.Warnf("Failed to get default environment for project %s: %v", project.ProjectID, err)

			return nil, err
		}

		project.DefaultEnvironment = defaultEnv
		result = append(result, project)
	}

	pm.loaded = true
	pm.records = result
	pm.logger.Debugf("Loaded %d projects", len(pm.records))

	return result, nil
}

func (pm *ProjectManager) SubscribeAll(ctx context.Context) chan ProjectInfo {
	return pm.Subscribe(nil, ctx)
}

func (pm *ProjectManager) Subscribe(organization *OrganizationInfo, ctx context.Context) chan ProjectInfo {
	ch := make(chan ProjectInfo)

	go func() {
		defer func() {
			close(ch)
		}()
		projects, err := pm.List(organization, ctx)

		if err != nil {
			pm.logger.Errorf("Failed to list projects for subscription: %v", err)

			return
		}
		pm.logger.Debugf("Subscribing to %d projects", len(projects))

		// First, send all projects without environments
		for _, project := range projects {
			ch <- project
			pm.logger.Tracef("Sent project update for %s", project.ProjectID)
		}

		pm.logger.Debug("All projects sent, loading environments")
		// Then, load and send environments for each project
		for i := range projects {
			project := &projects[i]
			env, err := pm.fleetManager.Environment.LoadDefaultEnvironment(project, ctx)
			if err != nil {
				pm.logger.Warnf("Failed to load environment for project %s: %v", project.ProjectID, err)
				continue
			} else if project.DefaultEnvironment != env {
				pm.logger.Errorf("Loaded environment does not match cached one for project %s.", project.ProjectID)
				continue
			}
			pm.logger.Tracef("Project: %+v", project.ToJson())
			select {
			case <-ctx.Done():
				pm.logger.Debug("Project subscription cancelled")
				return
			case ch <- *project:
				pm.logger.Tracef("Sent project update for %s", project.ProjectID)
			}
		}
	}()

	return ch
}
