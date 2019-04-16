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

package dao

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// AddWebhookPolicy ...
func AddWebhookPolicy(policy models.WebhookPolicy) (int64, error) {
	o := GetOrmer()
	sql := `insert into webhook_policy (name, project_id, target, hook_types, enabled, description, filters, creation_time, update_time) 
				values (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`
	params := []interface{}{}
	now := time.Now()

	params = append(params, policy.Name, policy.ProjectID, policy.Target, policy.HookTypes, policy.Enabled,
		policy.Description, policy.Filters, now, now)

	var policyID int64
	err := o.Raw(sql, params...).QueryRow(&policyID)
	if err != nil {
		return 0, err
	}

	return policyID, nil
}

// GetWebhookPolicy ...
func GetWebhookPolicy(id int64) (*models.WebhookPolicy, error) {
	o := GetOrmer()
	sql := `select * from webhook_policy where id = ? and deleted = false`

	var policy models.WebhookPolicy

	if err := o.Raw(sql, id).QueryRow(&policy); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &policy, nil
}

// GetWebhookPolicyByName ...
func GetWebhookPolicyByName(projectID int64, name string) ([]*models.WebhookPolicy, error) {
	o := GetOrmer()
	sql := `select * from webhook_policy where deleted = false and project_id = ?`

	var args []interface{}
	args = append(args, projectID)
	if len(name) != 0 {
		sql += ` and name like ?`
		args = append(args, "%"+Escape(name)+"%")
	}

	var policy []*models.WebhookPolicy
	_, err := o.Raw(sql, args).QueryRows(&policy)

	return policy, err
}

// GetEnabledWebhookPolicyByProject ...
func GetEnabledWebhookPolicyByProject(projectID int64) ([]*models.WebhookPolicy, error) {
	o := GetOrmer()
	sql := `select * from webhook_policy where enabled = true and deleted = false and project_id = ?`

	var policies []*models.WebhookPolicy

	if _, err := o.Raw(sql, projectID).QueryRows(&policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// UpdateWebhookPolicy ...
func UpdateWebhookPolicy(policy *models.WebhookPolicy) error {
	o := GetOrmer()

	sql := `update webhook_policy 
		set project_id = ?, target = ?, name = ?, description = ?, filters = ?, hook_types = ?, enabled = ?, update_time = ? 
		where id = ?`

	_, err := o.Raw(sql, policy.ProjectID, policy.Target, policy.Name, policy.Description, policy.Filters,
		policy.HookTypes, policy.Enabled, time.Now(), policy.ID).Exec()

	return err
}

// DeleteWebhookPolicy ...
func DeleteWebhookPolicy(id int64) error {
	o := GetOrmer()
	policy := &models.WebhookPolicy{
		ID:         id,
		Deleted:    true,
		UpdateTime: time.Now(),
	}
	_, err := o.Update(policy, "Deleted")
	return err
}

// AddWebhookJob ...
func AddWebhookJob(job models.WebhookJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(&job)
}

// GetWebhookJob ...
func GetWebhookJob(id int64) (*models.WebhookJob, error) {
	o := GetOrmer()
	j := models.WebhookJob{ID: id}
	err := o.Read(&j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &j, nil
}

// GetTotalCountOfWebhookJobs ...
func GetTotalCountOfWebhookJobs(query ...*models.WebhookJobQuery) (int64, error) {
	qs := webhookJobQueryConditions(query...)
	return qs.Count()
}

// GetWebhookJobs ...
func GetWebhookJobs(query ...*models.WebhookJobQuery) ([]*models.WebhookJob, error) {
	jobs := []*models.WebhookJob{}

	qs := webhookJobQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-UpdateTime")

	_, err := qs.All(&jobs)
	return jobs, err
}

// DeleteWebhookJob ...
func DeleteWebhookJob(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.WebhookJob{ID: id})
	return err
}

// DeleteWebhookJobs ...
func DeleteWebhookJobs(policyID int64) (int64, error) {
	o := GetOrmer()
	return o.Delete(&models.WebhookJob{PolicyID: policyID})
}

// UpdateWebhookJobStatus ...
func UpdateWebhookJobStatus(id int64, status string) error {
	o := GetOrmer()
	j := models.WebhookJob{
		ID:         id,
		Status:     status,
		UpdateTime: time.Now(),
	}
	n, err := o.Update(&j, "Status", "UpdateTime")
	if n == 0 {
		log.Warningf("no records are updated when updating webhook job %d", id)
	}
	return err
}

// SetWebhookJobUUID ...
func SetWebhookJobUUID(id int64, uuid string) error {
	o := GetOrmer()
	j := models.WebhookJob{
		ID:   id,
		UUID: uuid,
	}
	n, err := o.Update(&j, "UUID")
	if n == 0 {
		log.Warningf("no records are updated when updating webhook job %d", id)
	}
	return err
}

func webhookJobQueryConditions(query ...*models.WebhookJobQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.WebhookJob{})
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if q.PolicyID != 0 {
		qs = qs.Filter("PolicyID", q.PolicyID)
	}
	//if len(q.OpUUID) > 0 {
	//	qs = qs.Filter("OpUUID__exact", q.OpUUID)
	//}
	if len(q.Statuses) > 0 {
		qs = qs.Filter("Status__in", q.Statuses)
	}
	if len(q.HookTypes) > 0 {
		qs = qs.Filter("HookType__in", q.HookTypes)
	}
	if q.StartTime != nil {
		qs = qs.Filter("CreationTime__gte", q.StartTime)
	}
	if q.EndTime != nil {
		qs = qs.Filter("CreationTime__lte", q.EndTime)
	}
	return qs
}
