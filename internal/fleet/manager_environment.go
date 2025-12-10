package fleet

import (
	"context"
	"encoding/json"
	"time"

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

type Environment struct {
	ProjectID        string              `json:"project_id"`
	Project          *ProjectInfo        `json:"-"`
	Ref              string              `json:"ref"`
	InfoFetchedAt    time.Time           `json:"info_fetched_at"`
	DetailsFetchedAt time.Time           `json:"details_fetched_at"`
	Info             *EnvironmentInfo    `json:"info"`
	Details          *EnvironmentDetails `json:"details"`
}

func (env *Environment) ToJson() []byte {
	bytes, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}
	return bytes
}

func (env *Environment) IsDefault() bool {
	return env.Ref == "." || env.Info != nil && env.Info.isDefault()
}

// Fetched from upsun env:list
type EnvironmentInfo struct {
	MachineName string `json:"machine_name"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Created     string `json:"created_at"`
	Updated     string `json:"updated_at"`
}

func (env *EnvironmentInfo) isDefault() bool {
	return env.Type == "production"
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
	logger       *logrus.Entry
	fleetManager *FleetManager
	records      []Environment
}

func NewEnvironmentManager(fleetManager *FleetManager) *EnvironmentManager {
	return &EnvironmentManager{
		fleetManager: fleetManager,
		logger:       fleetManager.rootLogger.WithField("component", "EnvironmentManager"),
		records:      []Environment{},
	}
}

func (em *EnvironmentManager) GetOrCreate(projectInfo *ProjectInfo, ref string) *Environment {
	// Check if we already have the environment loaded
	for _, env := range em.records {
		if env.ProjectID == projectInfo.ProjectID && env.Ref == ref {
			return &env
		}
	}

	// Try to load from cache
	if env := em.loadFromCache(projectInfo, ref); env != nil {
		return env
	}

	// Create a new environment
	env := &Environment{
		ProjectID:        projectInfo.ProjectID,
		Project:          projectInfo,
		Ref:              ref,
		InfoFetchedAt:    time.Unix(0, 0),
		DetailsFetchedAt: time.Unix(0, 0),
		Info:             nil,
		Details:          nil,
	}

	em.records = append(em.records, *env)

	return env
}

func (em *EnvironmentManager) Save(environment *Environment) error {
	cacheKey := getCacheKeyForEnvironment(environment.Project, environment.Ref)
	err := em.fleetManager.cache.Write(cacheKey, environment.ToJson())
	if err != nil {
		return err
	}

	return nil
}

func (em *EnvironmentManager) loadFromCache(projectInfo *ProjectInfo, ref string) *Environment {
	cacheKey := getCacheKeyForEnvironment(projectInfo, ref)
	data := em.fleetManager.cache.Read(cacheKey)
	var environment Environment
	err := json.Unmarshal(data, &environment)
	if err != nil {
		em.logger.Warnf("Failed to unmarshal environment from cache: %v", err)
		return nil
	}
	environment.Project = projectInfo
	return &environment
}

func getCacheKeyForEnvironment(projectInfo *ProjectInfo, ref string) string {
	return "env/" + projectInfo.ProjectID + "-" + ref
}

func (em *EnvironmentManager) LoadDefaultEnvironment(projectInfo *ProjectInfo, ctx context.Context) (*Environment, error) {
	return em.loadEnvironment(projectInfo, ".", ctx)
}

func (em *EnvironmentManager) loadEnvironment(projectInfo *ProjectInfo, ref string, ctx context.Context) (*Environment, error) {
	logger := em.logger.WithFields(logrus.Fields{"project": projectInfo.ProjectID, "ref": ref})
	logger.Debug("Loading environment")

	environment := em.GetOrCreate(projectInfo, ref)

	args := []string{"env:info", "--format=csv", "--project", projectInfo.ProjectID, "-e", ref, "--columns=*"}
	data, err := em.fleetManager.GetExecOutputWithCtx(args, ctx)
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
		MachineName: values["machine_name"],
		Title:       values["title"],
		Status:      values["status"],
		Type:        values["type"],
		Created:     values["created_at"],
		Updated:     values["updated_at"],
	}

	environment.Info = environmentInfo
	environment.Details = environmentDetails
	environment.InfoFetchedAt = time.Now()
	environment.DetailsFetchedAt = time.Now()

	if environmentInfo.isDefault() {
		projectInfo.DefaultEnvironment = environment
	}

	logger.Debug("Finished loading the environment")

	em.Save(environment)

	return environment, nil
}

// This method is not used and kept for potential future usage
func (em *EnvironmentManager) List(projectInfo *ProjectInfo) ([]Environment, error) {
	em.logger.Infof("Loading all environments for project %s", projectInfo.ProjectID)

	result := []Environment{}
	args := []string{"env:list", "--format=csv", "--project", projectInfo.ProjectID, "--columns=*"}
	data, err := em.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		ref := record["ID"]

		// 1. Record update or creation
		environment := em.GetOrCreate(projectInfo, ref)

		info := &EnvironmentInfo{
			MachineName: record["Machine name"],
			Title:       record["Title"],
			Status:      record["Status"],
			Type:        record["Type"],
			Created:     record["Created"],
			Updated:     record["Updated"],
		}
		environment.Info = info
		environment.InfoFetchedAt = time.Now()
		em.Save(environment)
		result = append(result, *environment)
	}

	return result, nil
}
