package models

import (
	"time"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/astaxie/beego/validation"
)

const (
	// P2PPreheatJobTable is the table name for p2p preheat jobs
	P2PPreheatJobTable = "p2p_preheat_job"
	// P2PPreheatPolicyTable is table name for p2p preheat policies
	P2PPreheatPolicyTable = "p2p_preheat_policy"
	// P2PTargetTable is table name for p2p target
	P2PTargetTable = "p2p_target"
)

// P2PPreheatPolicy is the model for a p2p preheat policy, which associates a project with a target by filters and types
type P2PPreheatPolicy struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	Name              string    `orm:"column(name)"`
	ProjectID         int64     `orm:"column(project_id)" `
	TargetIDs         string    `orm:"column(target_ids)"`
	Description       string    `orm:"column(description)"`
	Filters           string    `orm:"column(filters)"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime        time.Time `orm:"column(update_time);auto_now"`
	Enabled           bool      `orm:"column(enabled)"`
	Deleted           bool      `orm:"column(deleted)"`
}

// P2PPreheatJob is the model for a p2p preheat job, which is the execution unit on job service
type P2PPreheatJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Status       string    `orm:"column(status)" json:"status"`
	Repository   string    `orm:"column(repository)" json:"repository"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	Tag          string    `orm:"column(tag)" json:"tag"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// P2PTarget is the model for a P2P target.
type P2PTarget struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	URL          string    `orm:"column(url)" json:"endpoint"`
	Name         string    `orm:"column(name)" json:"name"`
	Username     string    `orm:"column(username)" json:"username"`
	Password     string    `orm:"column(password)" json:"password"`
	Type         int       `orm:"column(type)" json:"type"`
	Insecure     bool      `orm:"column(insecure)" json:"insecure"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// Valid ...
func (pt *P2PTarget) Valid(v *validation.Validation) {
	if len(pt.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(pt.Name) > 64 {
		v.SetError("name", "max length is 64")
	}

	url, err := utils.ParseEndpoint(pt.URL)
	if err != nil {
		v.SetError("endpoint", err.Error())
	} else {
		// Prevent SSRF security issue #3755
		pt.URL = url.Scheme + "://" + url.Host + url.Path
		if len(pt.URL) > 64 {
			v.SetError("endpoint", "max length is 64")
		}
	}

	// password is encoded using base64, the length of this field
	// in DB is 64, so the max length in request is 48
	if len(pt.Password) > 48 {
		v.SetError("password", "max length is 48")
	}
}

// TableName is required by by beego orm to map P2PPreheatPolicy to table p2p_preheat_policy
func (pp *P2PPreheatPolicy) TableName() string {
	return P2PPreheatPolicyTable
}

// TableName is required by by beego orm to map P2PPreheatJob to table p2p_preheat_job
func (pj *P2PPreheatJob) TableName() string {
	return P2PPreheatJobTable
}

// TableName is required by by beego orm to map P2PTarget to table p2p_target
func (pt *P2PTarget) TableName() string {
	return P2PTargetTable
}

// P2PPreheatJobQuery holds query conditions for p2p preheat job
type P2PPreheatJobQuery struct {
	PolicyID   int64
	Repository string
	Statuses   []string
	StartTime  *time.Time
	EndTime    *time.Time
	Pagination
}
