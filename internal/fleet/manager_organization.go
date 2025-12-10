package fleet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

//
// OrganizationManager
//

type OrganizationInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Label         string `json:"label"`
	Type          string `json:"type"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	OwnerID       string `json:"owner_id"`
	OwnerEmail    string `json:"owner_email"`
	OwnerUsername string `json:"owner_username"`
}

func (o OrganizationInfo) ToJson() string {
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(bytes)
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
	organizations, err := om.ListAll(context.Background())
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

func (om *OrganizationManager) ListAll(ctx context.Context) ([]OrganizationInfo, error) {
	if om.loaded {
		return om.records, nil
	}

	cached := om.fleetManager.cache.Read("organizations")
	if cached != nil {
		var organizations []OrganizationInfo
		err := json.Unmarshal(cached, &organizations)
		if err == nil {
			om.logger.Debugf("Loaded %d organizations from cache", len(organizations))
			om.loaded = true
			om.records = organizations
			return om.records, nil
		}
		om.logger.Warnf("Failed to load organizations from cache: %v", err)
	}

	result, err := om.loadAll(ctx)
	if err != nil {
		om.logger.Errorf("Failed to load organizations: %v", err)
		return nil, err
	}

	buffer, err := json.Marshal(result)
	if err != nil {
		om.logger.Errorf("Failed to marshal organizations for caching: %v", err)
		return nil, err
	}

	err = om.fleetManager.cache.Write("organizations", buffer)
	if err != nil {
		om.logger.Errorf("Failed to write organizations to cache: %v", err)
		return nil, err
	}

	om.loaded = true
	om.records = result

	return result, nil
}

func (om *OrganizationManager) loadAll(ctx context.Context) ([]OrganizationInfo, error) {
	om.logger.Info("Loading organizations")
	args := []string{"organization:list", "--format=csv", "--columns=*"}
	data, err := om.fleetManager.GetExecOutputWithCtx(args, ctx)

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
