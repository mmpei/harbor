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
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/p2ppreheat/controller"
	"errors"
)

// P2PPreheatJobAPI handles request to /api/jobs/p2ppreheat /api/jobs/p2ppreheat/:id/log
type P2PPreheatJobAPI struct {
	BaseController
	jobID int64
}

// Prepare validates that whether user has system admin role
func (pja *P2PPreheatJobAPI) Prepare() {
	pja.BaseController.Prepare()
	if !pja.SecurityCtx.IsAuthenticated() {
		pja.HandleUnauthorized()
		return
	}

	if len(pja.GetStringFromPath(":id")) != 0 {
		id, err := pja.GetInt64FromPath(":id")
		if err != nil {
			pja.HandleBadRequest(fmt.Sprintf("invalid ID: %s", pja.GetStringFromPath(":id")))
			return
		}
		pja.jobID = id
	}
}

// List filters jobs according to the parameters
func (pja *P2PPreheatJobAPI) List() {
	policyID, err := pja.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		pja.HandleBadRequest(fmt.Sprintf("invalid policy_id: %s", pja.GetString("policy_id")))
		return
	}

	policy, err := controller.PolicyManager.GetPolicy(policyID)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", policyID, err)
		pja.CustomAbort(http.StatusInternalServerError, "")
	}

	if policy == nil {
		pja.HandleNotFound(fmt.Sprintf("policy %d not found", policyID))
		return
	}

	if _, err := pja.doAuth(policy.Project.ProjectID); err != nil {
		return
	}

	query := &models.P2PPreheatJobQuery{
		PolicyID: policyID,
	}

	query.Statuses = pja.GetStrings("status")
	//query.OpUUID = pja.GetString("op_uuid")

	startTimeStr := pja.GetString("start_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			pja.HandleBadRequest(fmt.Sprintf("invalid start_time: %s", startTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.StartTime = &t
	}

	endTimeStr := pja.GetString("end_time")
	if len(endTimeStr) != 0 {
		i, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			pja.HandleBadRequest(fmt.Sprintf("invalid end_time: %s", endTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.EndTime = &t
	}

	query.Page, query.Size = pja.GetPaginationParams()

	total, err := dao.GetTotalCountOfP2PPreheatJobs(query)
	if err != nil {
		pja.HandleInternalServerError(fmt.Sprintf("failed to get total count of p2p preheat jobs of policy %d: %v", policyID, err))
		return
	}
	jobs, err := dao.GetP2PPreheatJobs(query)
	if err != nil {
		pja.HandleInternalServerError(fmt.Sprintf("failed to get p2p preheat jobs, query: %v :%v", query, err))
		return
	}

	pja.SetPaginationHeader(total, query.Page, query.Size)

	pja.Data["json"] = jobs
	pja.ServeJSON()
}

// Delete ...
func (pja *P2PPreheatJobAPI) Delete() {
	if pja.jobID == 0 {
		pja.HandleBadRequest("ID is nil")
		return
	}

	job, err := dao.GetP2PPreheatJob(pja.jobID)
	if err != nil {
		log.Errorf("failed to get job %d: %v", pja.jobID, err)
		pja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// check job
	if job == nil {
		pja.HandleNotFound(fmt.Sprintf("job %d not found", pja.jobID))
		return
	}
	if job.Status == models.JobPending || job.Status == models.JobRunning {
		pja.HandleBadRequest(fmt.Sprintf("job is %s, can not be deleted", job.Status))
		return
	}

	// admin can delete all jobs including jobs without policy
	if !pja.SecurityCtx.IsSysAdmin() {
		policy, err := controller.PolicyManager.GetPolicy(job.PolicyID)
		if err != nil {
			pja.ParseAndHandleError(fmt.Sprintf("failed to get p2p preheat policy %d", job.PolicyID), err)
			return
		}
		if policy.ID == 0 {
			pja.HandleNotFound(fmt.Sprintf("policy %d not found", job.PolicyID))
			return
		}
		if _, err := pja.doAuth(policy.Project.ProjectID); err != nil {
			return
		}
	}

	if err = dao.DeleteWebhookJob(pja.jobID); err != nil {
		log.Errorf("failed to deleted job %d: %v", pja.jobID, err)
		pja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// GetLog ...
func (pja *P2PPreheatJobAPI) GetLog() {
	if pja.jobID == 0 {
		pja.HandleBadRequest("ID is nil")
		return
	}

	job, err := dao.GetP2PPreheatJob(pja.jobID)
	if err != nil {
		pja.HandleInternalServerError(fmt.Sprintf("failed to get p2p preheat job %d: %v", pja.jobID, err))
		return
	}

	if job == nil {
		pja.HandleNotFound(fmt.Sprintf("p2p preheat job %d not found", pja.jobID))
		return
	}

	// admin can get all job logs including jobs without policy
	if !pja.SecurityCtx.IsSysAdmin() {
		policy, err := controller.PolicyManager.GetPolicy(job.PolicyID)
		if err != nil {
			pja.ParseAndHandleError(fmt.Sprintf("failed to get p2p preheat policy %d", job.PolicyID), err)
			return
		}
		if policy.ID == 0 {
			pja.HandleNotFound(fmt.Sprintf("policy %d not found", job.PolicyID))
			return
		}
		if _, err := pja.doAuth(policy.Project.ProjectID); err != nil {
			return
		}
	}

	logBytes, err := utils.GetJobServiceClient().GetJobLog(job.UUID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			pja.RenderError(httpErr.Code, "")
			log.Errorf(fmt.Sprintf("failed to get log of job %d: %d %s",
				pja.jobID, httpErr.Code, httpErr.Message))
			return
		}
		pja.HandleInternalServerError(fmt.Sprintf("failed to get log of job %s: %v",
			job.UUID, err))
		return
	}
	pja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	pja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = pja.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		pja.HandleInternalServerError(fmt.Sprintf("failed to write log of job %s: %v", job.UUID, err))
		return
	}
}

func (pja *P2PPreheatJobAPI) doAuth(projectID int64) (*models.Project, error) {
	project, err := pja.ProjectMgr.Get(projectID)
	if err != nil {
		pja.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return nil, err
	}
	if project == nil {
		pja.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return nil, errors.New("project not found")
	}

	if !(pja.Ctx.Input.IsGet() && pja.SecurityCtx.HasReadPerm(projectID) ||
		pja.SecurityCtx.HasAllPerm(projectID)) {
		pja.HandleForbidden(pja.SecurityCtx.GetUsername())
		return nil, errors.New("forbidden")
	}
	return project, nil
}
