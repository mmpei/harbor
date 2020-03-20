// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"

	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/p2ppreheat/controller"
	"github.com/goharbor/harbor/src/replication"
	p2p_models "github.com/goharbor/harbor/src/p2ppreheat/models"
	"errors"
)

// P2PPreheatPolicyAPI handles /api/policies/p2ppreheat/:id([0-9]+) /api/policies/p2ppreheat
type P2PPreheatPolicyAPI struct {
	BaseController
}

// Prepare ...
func (ppa *P2PPreheatPolicyAPI) Prepare() {
	ppa.BaseController.Prepare()
	if !ppa.SecurityCtx.IsAuthenticated() {
		ppa.HandleUnauthorized()
		return
	}
}

// Get ...
func (ppa *P2PPreheatPolicyAPI) Get() {
	id := ppa.GetIDFromURL()
	policy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get p2p preheat policy %d: %v", id, err)
		ppa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy == nil {
		ppa.HandleNotFound(fmt.Sprintf("p2p preheat policy %d not found", id))
		return
	}

	if policy.ID == 0 {
		ppa.HandleNotFound(fmt.Sprintf("p2p preheat policy %d not found", id))
		return
	}

	_, err = ppa.doAuth(policy.Project.ProjectID)
	if err != nil {
		return
	}

	for _, target := range policy.Targets {
		target.Password = ""
	}

	ppa.Data["json"] = policy
	ppa.ServeJSON()
}

// List ...
func (ppa *P2PPreheatPolicyAPI) List() {
	name := ppa.GetString("name")
	projectIDStr := ppa.GetString("project_id")
	var projectID int64
	if len(projectIDStr) > 0 {
		var err error
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			ppa.HandleBadRequest(fmt.Sprintf("invalid project ID: %s", projectIDStr))
			return
		}
	}

	_, err := ppa.doAuth(projectID)
	if err != nil {
		return
	}

	page, pageSize := ppa.GetPaginationParams()
	policies, count, err := controller.PolicyManager.GetPolicies(projectID, name, page, pageSize)
	if err != nil {
		log.Errorf("failed to get p2p preheat policies: %v, projectID: %d, name: %s", err, projectID, name)
		ppa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, policy := range policies {
		for _, target := range policy.Targets {
			target.Password = ""
		}
	}

	ppa.SetPaginationHeader(count, page, pageSize)
	ppa.Data["json"] = policies
	ppa.ServeJSON()
}

// Post creates a p2p preheat policy
func (ppa *P2PPreheatPolicyAPI) Post() {
	policy := &p2p_models.P2PPreheatPolicy{}
	ppa.DecodeJSONReqAndValidate(policy)

	_, err := ppa.doAuth(policy.Project.ProjectID)
	if err != nil {
		return
	}

	// check the name
	exist, err := p2pPreheatPolicyExist(policy.Name)
	if err != nil {
		ppa.HandleInternalServerError(fmt.Sprintf("failed to check the existence of policy %s: %v", policy.Name, err))
		return
	}

	if exist {
		ppa.HandleConflict(fmt.Sprintf("name %s is already used", policy.Name))
		return
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				ppa.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				ppa.HandleNotFound(fmt.Sprintf("label %d not found", labelID))
				return
			}
		}
	}

	// check existence of target
	for _, target := range policy.Targets {
		t, err := dao.GetP2PTarget(target.ID)
		if err != nil {
			ppa.HandleInternalServerError(fmt.Sprintf("failed to find target %d: %v", target.ID, err))
			return
		}
		if t == nil {
			ppa.HandleNotFound(fmt.Sprintf("target %d not found", target.ID))
			return
		}
	}

	id, err := controller.PolicyManager.CreatePolicy(policy)
	if err != nil {
		ppa.HandleInternalServerError(fmt.Sprintf("failed to create p2p preheat policy: %v", err))
		return
	}

	ppa.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

func p2pPreheatPolicyExist(name string) (bool, error) {
	policy, err := dao.GetP2PPreheatPolicyByName(name)
	if err != nil {
		return false, err
	}

	if policy == nil {
		return false, nil
	}
	return true, nil
}

// Put updates the p2p preheat policy
func (ppa *P2PPreheatPolicyAPI) Put() {
	id := ppa.GetIDFromURL()

	originalPolicy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		ppa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if originalPolicy == nil {
		ppa.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	if _, err := ppa.doAuth(originalPolicy.Project.ProjectID); err != nil {
		return
	}

	policy := &p2p_models.P2PPreheatPolicy{}
	ppa.DecodeJSONReqAndValidate(policy)

	policy.ID = id

	// check projectID, projectID shouldn't be changed
	if policy.Project == nil || originalPolicy.Project.ProjectID != policy.Project.ProjectID {
		ppa.HandleBadRequest("project shouldn't be changed")
		return
	}

	// check the name
	if policy.Name != originalPolicy.Name {
		exist, err := p2pPreheatPolicyExist(policy.Name)
		if err != nil {
			ppa.HandleInternalServerError(fmt.Sprintf("failed to check the existence of policy %s: %v", policy.Name, err))
			return
		}

		if exist {
			ppa.HandleConflict(fmt.Sprintf("name %s is already used", policy.Name))
			return
		}
	}

	// check hook type
	for _, target := range policy.Targets {
		t, err := dao.GetP2PTarget(target.ID)
		if err != nil {
			ppa.HandleInternalServerError(fmt.Sprintf("failed to find target %d: %v", target.ID, err))
			return
		}
		if t == nil {
			ppa.HandleNotFound(fmt.Sprintf("target %d not found", target.ID))
			return
		}
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				ppa.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				ppa.HandleNotFound(fmt.Sprintf("label %d not found", labelID))
				return
			}
		}
	}

	if err = controller.PolicyManager.UpdatePolicy(policy); err != nil {
		ppa.HandleInternalServerError(fmt.Sprintf("failed to update policy %d: %v", id, err))
		return
	}
}

// Delete the p2p preheat policy
func (ppa *P2PPreheatPolicyAPI) Delete() {
	id := ppa.GetIDFromURL()

	policy, err := controller.PolicyManager.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		ppa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy == nil {
		ppa.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	if _, err := ppa.doAuth(policy.Project.ProjectID); err != nil {
		return
	}

	count, err := dao.GetTotalCountOfP2PPreheatJobs(&models.P2PPreheatJobQuery{
		PolicyID: id,
		Statuses: []string{models.JobRunning, models.JobRetrying, models.JobPending},
	})
	if err != nil {
		log.Errorf("failed to filter jobs of policy %d: %v", id, err)
		ppa.CustomAbort(http.StatusInternalServerError, "")
	}
	if count > 0 {
		ppa.CustomAbort(http.StatusPreconditionFailed, "policy has running/retrying/pending jobs, can not be deleted")
	}

	// maybe this will make the jobs related to this policy to orphan
	// however removing policy is a logical operate, this makes stuffs acceptable
	if err = controller.PolicyManager.RemovePolicy(id); err != nil {
		log.Errorf("failed to delete policy %d: %v", id, err)
		ppa.CustomAbort(http.StatusInternalServerError, "")
	}
}

func (ppa *P2PPreheatPolicyAPI) doAuth(projectID int64) (*models.Project, error) {
	project, err := ppa.ProjectMgr.Get(projectID)
	if err != nil {
		ppa.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return nil, err
	}
	if project == nil {
		ppa.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return nil, errors.New("project not found")
	}

	if !(ppa.Ctx.Input.IsGet() && ppa.SecurityCtx.HasReadPerm(projectID) ||
		ppa.SecurityCtx.HasAllPerm(projectID)) {
		ppa.HandleForbidden(ppa.SecurityCtx.GetUsername())
		return nil, errors.New("forbidden")
	}
	return project, nil
}
