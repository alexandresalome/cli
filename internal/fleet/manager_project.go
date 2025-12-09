package fleet

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
