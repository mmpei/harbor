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

// AddP2PTarget ...
func AddP2PTarget(target models.P2PTarget) (int64, error) {
	o := GetOrmer()

	sql := "insert into p2p_target (name, url, username, password, insecure, type) values (?, ?, ?, ?, ?, ?) RETURNING id"

	var targetID int64
	err := o.Raw(sql, target.Name, target.URL, target.Username, target.Password, target.Insecure, target.Type).QueryRow(&targetID)
	if err != nil {
		return 0, err
	}
	return targetID, nil
}

// GetP2PTarget ...
func GetP2PTarget(id int64) (*models.P2PTarget, error) {
	o := GetOrmer()
	t := models.P2PTarget{ID: id}
	err := o.Read(&t)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// GetP2PTargetByName ...
func GetP2PTargetByName(name string) (*models.P2PTarget, error) {
	o := GetOrmer()
	t := models.P2PTarget{Name: name}
	err := o.Read(&t, "Name")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// GetP2PTargetByEndpoint ...
func GetP2PTargetByEndpoint(endpoint string) (*models.P2PTarget, error) {
	o := GetOrmer()
	t := models.P2PTarget{
		URL: endpoint,
	}
	err := o.Read(&t, "URL")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// DeleteP2PTarget ...
func DeleteP2PTarget(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.P2PTarget{ID: id})
	return err
}

// UpdateP2PTarget ...
func UpdateP2PTarget(target models.P2PTarget) error {
	o := GetOrmer()

	sql := `update p2p_target 
	set url = ?, name = ?, username = ?, password = ?, insecure = ?, update_time = ?
	where id = ?`

	_, err := o.Raw(sql, target.URL, target.Name, target.Username, target.Password, target.Insecure, time.Now(), target.ID).Exec()

	return err
}

// FilterP2PTargets filters targets by name
func FilterP2PTargets(name string) ([]*models.P2PTarget, error) {
	o := GetOrmer()

	var args []interface{}

	sql := `select * from p2p_target `
	if len(name) != 0 {
		sql += `where name like ? `
		args = append(args, "%"+Escape(name)+"%")
	}
	sql += `order by creation_time desc`

	var targets []*models.P2PTarget

	if _, err := o.Raw(sql, args).QueryRows(&targets); err != nil {
		return nil, err
	}

	return targets, nil
}

// AddP2PPreheatPolicy ...
func AddP2PPreheatPolicy(policy models.P2PPreheatPolicy) (int64, error) {
	o := GetOrmer()
	sql := `insert into p2p_preheat_policy (name, project_id, target_ids, enabled, description, creation_time, update_time, filters) 
				values (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`
	params := []interface{}{}
	now := time.Now()

	params = append(params, policy.Name, policy.ProjectID, policy.TargetIDs, true, policy.Description, now, now, policy.Filters)

	var policyID int64
	err := o.Raw(sql, params...).QueryRow(&policyID)
	if err != nil {
		return 0, err
	}

	return policyID, nil
}

// GetP2PPreheatPolicy ...
func GetP2PPreheatPolicy(id int64) (*models.P2PPreheatPolicy, error) {
	o := GetOrmer()
	sql := `select * from p2p_preheat_policy where id = ? and deleted = false`

	var policy models.P2PPreheatPolicy

	if err := o.Raw(sql, id).QueryRow(&policy); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &policy, nil
}

// GetTotalOfP2PPreheatPolicies returns the total count of p2p preheat policies
func GetTotalOfP2PPreheatPolicies(name string, projectID int64) (int64, error) {
	qs := GetOrmer().QueryTable(&models.P2PPreheatPolicy{}).Filter("deleted", false)

	if len(name) != 0 {
		qs = qs.Filter("name__icontains", name)
	}

	if projectID != 0 {
		qs = qs.Filter("project_id", projectID)
	}

	return qs.Count()
}

// FilterP2PPreheatPolicies filters policies by name and project ID
func FilterP2PPreheatPolicies(name string, projectID, page, pageSize int64) ([]*models.P2PPreheatPolicy, error) {
	o := GetOrmer()

	var args []interface{}

	sql := `select * from p2p_preheat_policy where deleted = false `

	if len(name) != 0 && projectID != 0 {
		sql += `and name like ? and project_id = ? `
		args = append(args, "%"+Escape(name)+"%")
		args = append(args, projectID)
	} else if len(name) != 0 {
		sql += `and name like ? `
		args = append(args, "%"+Escape(name)+"%")
	} else if projectID != 0 {
		sql += `and project_id = ? `
		args = append(args, projectID)
	}

	sql += `order by creation_time desc`

	if page > 0 && pageSize > 0 {
		sql += ` limit ? offset ?`
		args = append(args, pageSize, (page-1)*pageSize)
	}

	var policies []*models.P2PPreheatPolicy
	if _, err := o.Raw(sql, args).QueryRows(&policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// GetP2PPreheatPolicyByName ...
func GetP2PPreheatPolicyByName(name string) (*models.P2PPreheatPolicy, error) {
	o := GetOrmer()
	sql := `select * from p2p_preheat_policy where deleted = false and name = ?`

	var policy models.P2PPreheatPolicy

	if err := o.Raw(sql, name).QueryRow(&policy); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &policy, nil
}

// GetP2PPreheatPolicyByProject ...
func GetP2PPreheatPolicyByProject(projectID int64) ([]*models.P2PPreheatPolicy, error) {
	o := GetOrmer()
	sql := `select * from p2p_preheat_policy where deleted = false and project_id = ?`

	var policies []*models.P2PPreheatPolicy

	if _, err := o.Raw(sql, projectID).QueryRows(&policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// GetP2PPreheatPolicyByTarget ...
func GetP2PPreheatPolicyByTarget(targetID int64) ([]*models.P2PPreheatPolicy, error) {
	o := GetOrmer()
	sql := `select * from p2p_preheat_policy where deleted = false and ? = ANY(STRING_TO_ARRAY(target_ids, ','))`

	var policies []*models.P2PPreheatPolicy

	if _, err := o.Raw(sql, targetID).QueryRows(&policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// UpdateP2PPreheatPolicy ...
func UpdateP2PPreheatPolicy(policy *models.P2PPreheatPolicy) error {
	o := GetOrmer()

	sql := `update p2p_preheat_policy 
		set project_id = ?, target_ids = ?, name = ?, description = ?, filters = ?, update_time = ? where id = ?`

	_, err := o.Raw(sql, policy.ProjectID, policy.TargetIDs, policy.Name, policy.Description, policy.Filters, time.Now(), policy.ID).Exec()

	return err
}

// DeleteP2PPreheatPolicy ...
func DeleteP2PPreheatPolicy(id int64) error {
	o := GetOrmer()
	policy := &models.P2PPreheatPolicy{
		ID:         id,
		Deleted:    true,
		UpdateTime: time.Now(),
	}
	_, err := o.Update(policy, "Deleted")
	return err
}

// AddP2PPreheatJob ...
func AddP2PPreheatJob(job models.P2PPreheatJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(&job)
}

// GetP2PPreheatJob ...
func GetP2PPreheatJob(id int64) (*models.P2PPreheatJob, error) {
	o := GetOrmer()
	j := models.P2PPreheatJob{ID: id}
	err := o.Read(&j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &j, nil
}

// GetTotalCountOfP2PPreheatJobs ...
func GetTotalCountOfP2PPreheatJobs(query ...*models.P2PPreheatJobQuery) (int64, error) {
	qs := p2pPreheatJobQueryConditions(query...)
	return qs.Count()
}

// GetP2PPreheatJobs ...
func GetP2PPreheatJobs(query ...*models.P2PPreheatJobQuery) ([]*models.P2PPreheatJob, error) {
	jobs := []*models.P2PPreheatJob{}

	qs := p2pPreheatJobQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-UpdateTime")

	if _, err := qs.All(&jobs); err != nil {
		return jobs, err
	}
	return jobs, nil
}

func p2pPreheatJobQueryConditions(query ...*models.P2PPreheatJobQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(new(models.P2PPreheatJob))
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if q.PolicyID != 0 {
		qs = qs.Filter("PolicyID", q.PolicyID)
	}
	if len(q.Repository) > 0 {
		qs = qs.Filter("Repository__icontains", q.Repository)
	}
	if len(q.Statuses) > 0 {
		qs = qs.Filter("Status__in", q.Statuses)
	}
	if q.StartTime != nil {
		qs = qs.Filter("CreationTime__gte", q.StartTime)
	}
	if q.EndTime != nil {
		qs = qs.Filter("CreationTime__lte", q.EndTime)
	}
	return qs
}

// DeleteP2PPreheatJob ...
func DeleteP2PPreheatJob(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.P2PPreheatJob{ID: id})
	return err
}

// UpdateP2PPreheatJobStatus ...
func UpdateP2PPreheatJobStatus(id int64, status string) error {
	o := GetOrmer()
	j := models.P2PPreheatJob{
		ID:         id,
		Status:     status,
		UpdateTime: time.Now(),
	}
	n, err := o.Update(&j, "Status", "UpdateTime")
	if n == 0 {
		log.Warningf("no records are updated when updating p2p preheat job %d", id)
	}
	return err
}

// SetP2PPreheatJobUUID ...
func SetP2PPreheatJobUUID(id int64, uuid string) error {
	o := GetOrmer()
	j := models.P2PPreheatJob{
		ID:   id,
		UUID: uuid,
	}
	n, err := o.Update(&j, "UUID")
	if n == 0 {
		log.Warningf("no records are updated when updating p2p preheat job %d", id)
	}
	return err
}
