package fleet

import "github.com/sirupsen/logrus"

type EnvironmentInfo struct {
	Project     *ProjectInfo
	ID          string
	MachineName string
	Title       string
	Status      string
	Type        string
	Created     string
	Updated     string
}

func (env *EnvironmentInfo) IsProduction() bool {
	return env.Type == "production"
}

type EnvironmentManager struct {
	logger       *logrus.Entry
	fleetManager *FleetManager
	records      map[string][]EnvironmentInfo
}

func NewEnvironmentManager(fleetManager *FleetManager) *EnvironmentManager {
	return &EnvironmentManager{
		fleetManager: fleetManager,
		logger:       fleetManager.rootLogger.WithField("component", "EnvironmentManager"),
		records:      map[string][]EnvironmentInfo{},
	}
}

func (pm *EnvironmentManager) List(projectInfo *ProjectInfo) ([]EnvironmentInfo, error) {
	if envs, exists := pm.records[projectInfo.ProjectID]; exists {
		return envs, nil
	}

	pm.logger.Infof("Loading environments for project %s", projectInfo.ProjectID)
	result := []EnvironmentInfo{}
	args := []string{"env", "--format=csv", "--project", projectInfo.ProjectID, "--columns=*"}
	data, err := pm.fleetManager.GetExecOutput(args)

	if err != nil {
		return nil, err
	}

	parser, err := NewCsvParser(data)
	if err != nil {
		return nil, err
	}

	for _, record := range parser.GetRecords() {
		environment := EnvironmentInfo{
			Project:     projectInfo,
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

	pm.records[projectInfo.ProjectID] = result

	return result, nil
}
