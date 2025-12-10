package fleet

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type ProjectInfo struct {
	ProjectID             string            `json:"id"`
	ProjectTitle          string            `json:"title"`
	Region                string            `json:"region"`
	Organization          *OrganizationInfo `json:"organization"`
	OrganizationName      string
	OrganizationID        string
	OrganizationLabel     string
	OrganizationType      string
	Status                string           `json:"status"`
	Created               string           `json:"created_at"`
	ProductionLoaded      bool             `json:"production_loaded"`
	ProductionEnvironment *EnvironmentInfo `json:"production_environment"`
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
			ProjectID:             record["ID"],
			ProjectTitle:          record["Title"],
			Region:                record["Region"],
			Organization:          org,
			OrganizationName:      record["Org name"],
			OrganizationID:        record["Org ID"],
			OrganizationLabel:     record["Org label"],
			OrganizationType:      record["Org type"],
			Status:                record["Status"],
			Created:               record["Created"],
			ProductionLoaded:      false,
			ProductionEnvironment: nil,
		}
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
			return
		}

		// First, send all projects without environments
		for _, project := range projects {
			ch <- project
		}

		// Then, load and send environments for each project
		for _, project := range projects {
			env, err := pm.fleetManager.Environment.FindProductionEnvironment(&project, ctx)
			if err == nil && env != nil {
				project.ProductionEnvironment = env
				project.ProductionLoaded = true
			}
			select {
			case <-ctx.Done():
				return
			case ch <- project:
			}
		}
	}()

	return ch
}
