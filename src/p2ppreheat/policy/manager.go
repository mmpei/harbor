// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/p2ppreheat/models"
	rep_models "github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"strings"
	"strconv"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Manager defines the method a policy manger should implement
type Manager interface {
	GetPolicies(projectID int64, name string, page, pageSize int64) ([]*models.P2PPreheatPolicy, int64, error)
	GetPoliciesByTargetId(targetId int64) ([]*models.P2PPreheatPolicy, error)
	GetEnabledPolicies(projectID int64) ([]*models.P2PPreheatPolicy, error)
	GetPolicy(int64) (*models.P2PPreheatPolicy, error)
	CreatePolicy(*models.P2PPreheatPolicy) (int64, error)
	UpdatePolicy(*models.P2PPreheatPolicy) error
	RemovePolicy(int64) error
}

// DefaultManager provides p2p preheat policy CURD capabilities.
type DefaultManager struct{}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// GetPolicies returns all the policies in project
func (m *DefaultManager) GetPolicies(projectID int64, name string, page, pageSize int64) ([]*models.P2PPreheatPolicy, int64, error) {
	policies, err := dao.FilterP2PPreheatPolicies(name, projectID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	count, err := dao.GetTotalOfP2PPreheatPolicies(name, projectID)
	if err != nil {
		return nil, 0, err
	}

	list := []*models.P2PPreheatPolicy{}
	for _, policy := range policies {
		ply, err := convertFromDBModel(policy)
		if err != nil {
			return nil, 0, err
		}

		list = append(list, ply)
	}

	return list, count, nil
}

// GetPoliciesByTargetId returns all the policies related to p2p target
func (m *DefaultManager) GetPoliciesByTargetId(targetId int64) ([]*models.P2PPreheatPolicy, error) {
	policies, err := dao.GetP2PPreheatPolicyByTarget(targetId)
	if err != nil {
		return nil, err
	}

	list := []*models.P2PPreheatPolicy{}
	for _, policy := range policies {
		ply, err := convertFromDBModel(policy)
		if err != nil {
			return nil, err
		}

		list = append(list, ply)
	}

	return list, nil
}

// GetEnabledPolicies returns all the enabled policies in project
func (m *DefaultManager) GetEnabledPolicies(projectID int64) ([]*models.P2PPreheatPolicy, error) {
	policies, err := dao.GetP2PPreheatPolicyByProject(projectID)
	if err != nil {
		return nil, err
	}

	list := []*models.P2PPreheatPolicy{}
	for _, policy := range policies {
		if policy.Enabled == false {
			continue
		}
		ply, err := convertFromDBModel(policy)
		if err != nil {
			return nil, err
		}

		list = append(list, ply)
	}

	return list, nil
}

// GetPolicy returns the policy with the specified ID
func (m *DefaultManager) GetPolicy(policyID int64) (*models.P2PPreheatPolicy, error) {
	policy, err := dao.GetP2PPreheatPolicy(policyID)
	if err != nil {
		return nil, err
	}

	return convertFromDBModel(policy)
}

func convertFromDBModel(policy *common_models.P2PPreheatPolicy) (*models.P2PPreheatPolicy, error) {
	if policy == nil {
		return nil, nil
	}

	ply := &models.P2PPreheatPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
		Enabled:           policy.Enabled,
	}

	prj, err := config.GlobalProjectMgr.Get(policy.ProjectID)
	if err != nil {
		return nil, err
	}
	ply.Project = prj

	if len(policy.Filters) > 0 {
		filters := []rep_models.Filter{}
		if err := json.Unmarshal([]byte(policy.Filters), &filters); err != nil {
			return nil, err
		}
		for i := range filters {
			if filters[i].Value == nil && len(filters[i].Pattern) > 0 {
				filters[i].Value = filters[i].Pattern
			}
			// convert the type of Value to int64 as the default type of
			// json Unmarshal for number is float64
			if filters[i].Kind == replication.FilterItemKindLabel {
				filters[i].Value = int64(filters[i].Value.(float64))
			}
		}
		ply.Filters = filters
	}

	if len(policy.TargetIDs) > 0 {
		targetList := []*common_models.P2PTarget{}
		ids := strings.Split(policy.TargetIDs, ",")
		for _, id := range ids {
			ID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Warningf("target id convert err: %v", err)
				continue
			}
			target, err := dao.GetP2PTarget(ID)
			if err != nil {
				log.Warningf("get p2p target err: %v", err)
				continue
			}
			targetList = append(targetList, target)
		}
		ply.Targets = targetList
	}

	return ply, nil
}

func convertToDBModel(policy *models.P2PPreheatPolicy) (*common_models.P2PPreheatPolicy, error) {
	ply := &common_models.P2PPreheatPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		ProjectID:         policy.Project.ProjectID,
		Description:       policy.Description,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
		Enabled:           policy.Enabled,
	}

	if len(policy.Targets) > 0 {
		var build strings.Builder
		for _, target := range policy.Targets {
			if build.Len() != 0 {
				build.WriteString(",")
			}
			build.WriteString(strconv.FormatInt(target.ID, 10))
		}
		ply.TargetIDs = build.String()
	}

	if len(policy.Filters) > 0 {
		filters, err := json.Marshal(policy.Filters)
		if err != nil {
			return nil, err
		}
		ply.Filters = string(filters)
	}

	return ply, nil
}

// CreatePolicy creates a new policy with the provided data;
// If creating failed, error will be returned;
// If creating succeed, ID of the new created policy will be returned.
func (m *DefaultManager) CreatePolicy(policy *models.P2PPreheatPolicy) (int64, error) {
	ply, err := convertToDBModel(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddP2PPreheatPolicy(*ply)
}

// UpdatePolicy updates the policy;
// If updating failed, error will be returned.
func (m *DefaultManager) UpdatePolicy(policy *models.P2PPreheatPolicy) error {
	ply, err := convertToDBModel(policy)
	if err != nil {
		return err
	}
	return dao.UpdateP2PPreheatPolicy(ply)
}

// RemovePolicy removes the specified policy;
// If removing failed, error will be returned.
func (m *DefaultManager) RemovePolicy(policyID int64) error {
	return dao.DeleteP2PPreheatPolicy(policyID)
}

