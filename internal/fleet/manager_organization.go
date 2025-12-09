package fleet

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
