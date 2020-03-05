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

package models

import (
	"time"

	"github.com/astaxie/beego/validation"
	common_models "github.com/goharbor/harbor/src/common/models"
	rep_models "github.com/goharbor/harbor/src/replication/models"
)

// P2PPreheatPolicy defines the data model used in API level
type P2PPreheatPolicy struct {
	ID                        int64                      `json:"id"`
	Name                      string                     `json:"name"`
	Description               string                     `json:"description"`
	Filters                   []rep_models.Filter        `json:"filters"`
	Project                   *common_models.Project     `json:"project"`
	Targets                   []*common_models.P2PTarget `json:"targets"`
	CreationTime              time.Time                  `json:"creation_time"`
	UpdateTime                time.Time                  `json:"update_time"`
	Enabled                   bool                       `json:"enabled"`
}

// Valid ...
func (ppp *P2PPreheatPolicy) Valid(v *validation.Validation) {
	if len(ppp.Name) == 0 {
		v.SetError("name", "can not be empty")
	}

	if len(ppp.Name) > 256 {
		v.SetError("name", "max length is 256")
	}

	if len(ppp.Targets) == 0 {
		v.SetError("targets", "can not be empty")
	}

	for i := range ppp.Filters {
		ppp.Filters[i].Valid(v)
	}
}

