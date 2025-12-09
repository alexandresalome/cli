package fleet

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

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
	logger       *logrus.Entry
	fleetManager *FleetManager
	loaded       bool
	records      []OrganizationInfo
}

func NewOrganizationManager(fleetManager *FleetManager) *OrganizationManager {
	return &OrganizationManager{
		logger:       fleetManager.rootLogger.WithField("component", "OrganizationManager"),
		fleetManager: fleetManager,
		loaded:       false,
		records:      []OrganizationInfo{},
	}
}

func (om *OrganizationManager) GetByID(id string) (*OrganizationInfo, error) {
	organizations, err := om.List()
	if err != nil {
		return nil, err
	}

	for _, org := range organizations {
		if org.ID == id {
			return &org, nil
		}
	}

	return nil, fmt.Errorf("organization with ID %s not found", id)
}

func (om *OrganizationManager) List() ([]OrganizationInfo, error) {
	if om.loaded {
		return om.records, nil
	}

	om.logger.Info("Loading organizations")
	args := []string{"organization:list", "--format=csv", "--columns=*"}
	data, err := om.fleetManager.GetExecOutput(args)

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
		om.records = append(om.records, organization)
	}

	om.logger.Debugf("Loaded %d organizations", len(om.records))
	om.loaded = true

	return om.records, nil
}
