package fleet

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type httpAccessConfig struct {
	Enabled bool `yaml:"is_enabled"`
}

type deploymentStateConfig struct {
	LastDeploymentSuccessful bool   `yaml:"last_deployment_successful"`
	LastDeploymentAt         string `yaml:"last_deployment_at"`
}

// Fetched from upsun env:list
type EnvironmentInfo struct {
	Project       *ProjectInfo        `json:"-"`
	ID            string              `json:"id"`
	MachineName   string              `json:"machine_name"`
	Title         string              `json:"title"`
	Status        string              `json:"status"`
	Type          string              `json:"type"`
	Created       string              `json:"created_at"`
	Updated       string              `json:"updated_at"`
	DetailsLoaded bool                `json:"details_loaded"`
	Details       *EnvironmentDetails `json:"details,omitempty"`
}

// Fetched from upsun env:info
type EnvironmentDetails struct {
	DefaultDomain            string `json:"default_domain"`
	HttpAccessEnabled        bool   `json:"http_access_enabled"`
	RobotsRestrictionEnabled bool   `json:"robots_restriction_enabled"`
	LastDeploymentSuccessful bool   `json:"last_deployment_successful"`
	LastDeploymentAt         string `json:"last_deployment_at"`
}

func (env *EnvironmentInfo) IsProduction() bool {
	return env.Type == "production"
}

type EnvironmentManager struct {
	logger                             *logrus.Entry
	fleetManager                       *FleetManager
	records                            map[string]EnvironmentInfo
	projectIDToEnvironmentIDs          map[string][]string
	projectIDToProductionEnvironmentID map[string]string
}

func NewEnvironmentManager(fleetManager *FleetManager) *EnvironmentManager {
	return &EnvironmentManager{
		fleetManager:                       fleetManager,
		logger:                             fleetManager.rootLogger.WithField("component", "EnvironmentManager"),
		records:                            map[string]EnvironmentInfo{},
		projectIDToEnvironmentIDs:          map[string][]string{},
		projectIDToProductionEnvironmentID: map[string]string{},
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (pm *EnvironmentManager) FindProductionEnvironment(projectInfo *ProjectInfo) (*EnvironmentInfo, error) {
	args := []string{"env:info", "--format=csv", "--project", projectInfo.ProjectID, "-e", ".", "--columns=*"}
	data, err := pm.fleetManager.GetExecOutput(args)
	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	values := map[string]string{}
	for _, record := range parser.GetRecords() {
		property := record["Property"]
		value := record["Value"]
		values[property] = value
	}

	httpAccess := values["http_access"]
	yamlAccess := httpAccessConfig{}
	err = yaml.Unmarshal([]byte(httpAccess), &yamlAccess)
	if err != nil {
		return nil, err
	}
	deploymentState := values["deployment_state"]
	yamlDeployment := deploymentStateConfig{}
	err = yaml.Unmarshal([]byte(deploymentState), &yamlDeployment)
	if err != nil {
		return nil, err
	}

	environmentDetails := &EnvironmentDetails{
		DefaultDomain:            values["default_domain"],
		HttpAccessEnabled:        yamlAccess.Enabled,
		RobotsRestrictionEnabled: values["restrict_robots"] == "true",
		LastDeploymentSuccessful: yamlDeployment.LastDeploymentSuccessful,
		LastDeploymentAt:         yamlDeployment.LastDeploymentAt,
	}

	environmentInfo := &EnvironmentInfo{
		Project:       projectInfo,
		ID:            values["id"],
		MachineName:   values["machine_name"],
		Title:         values["title"],
		Status:        values["status"],
		Type:          values["type"],
		Created:       values["created_at"],
		Updated:       values["updated_at"],
		DetailsLoaded: true,
		Details:       environmentDetails,
	}

	projectInfo.ProductionEnvironment = environmentInfo
	projectInfo.ProductionLoaded = true

	pm.records[environmentInfo.ID] = *environmentInfo

	return environmentInfo, nil
}

func (pm *EnvironmentManager) List(projectInfo *ProjectInfo) ([]EnvironmentInfo, error) {
	// Check if we have already loaded the environments for this project
	if envs, exists := pm.projectIDToEnvironmentIDs[projectInfo.ProjectID]; exists {
		result := []EnvironmentInfo{}
		for _, env := range envs {
			result = append(result, pm.records[env])
		}
		return result, nil
	}

	pm.logger.Infof("Loading all environments for project %s", projectInfo.ProjectID)
	result := []EnvironmentInfo{}
	args := []string{"env:list", "--format=csv", "--project", projectInfo.ProjectID, "--columns=*"}
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		recordID := record["ID"]

		// 1. Record update or creation
		existingRecord, exists := pm.records[recordID]
		if !exists {
			existingRecord = EnvironmentInfo{
				Project:       projectInfo,
				ID:            record["ID"],
				MachineName:   record["Machine name"],
				Title:         record["Title"],
				Status:        record["Status"],
				Type:          record["Type"],
				Created:       record["Created"],
				Updated:       record["Updated"],
				DetailsLoaded: false,
				Details:       nil,
			}
			pm.records[recordID] = existingRecord
		} else {
			existingRecord.Project = projectInfo
			existingRecord.MachineName = record["Machine name"]
			existingRecord.Title = record["Title"]
			existingRecord.Status = record["Status"]
			existingRecord.Type = record["Type"]
			existingRecord.Created = record["Created"]
			existingRecord.Updated = record["Updated"]
		}

		// 2. Update the projectIDToEnvironmentIDs map
		if _, found := pm.projectIDToEnvironmentIDs[projectInfo.ProjectID]; !found {
			pm.projectIDToEnvironmentIDs[projectInfo.ProjectID] = []string{}
		}

		if !contains(pm.projectIDToEnvironmentIDs[projectInfo.ProjectID], recordID) {
			pm.projectIDToEnvironmentIDs[projectInfo.ProjectID] = append(pm.projectIDToEnvironmentIDs[projectInfo.ProjectID], recordID)
		}

		environment := EnvironmentInfo{
			Project:       projectInfo,
			ID:            record["ID"],
			MachineName:   record["Machine name"],
			Title:         record["Title"],
			Status:        record["Status"],
			Type:          record["Type"],
			Created:       record["Created"],
			Updated:       record["Updated"],
			DetailsLoaded: false,
			Details:       nil,
		}

		if environment.IsProduction() {
			pm.projectIDToProductionEnvironmentID[projectInfo.ProjectID] = recordID
		}

		result = append(result, environment)
	}

	return result, nil
}
