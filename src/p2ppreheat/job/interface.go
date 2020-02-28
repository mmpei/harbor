package job

import (
	"github.com/goharbor/harbor/src/p2ppreheat/models"
	filter_models "github.com/goharbor/harbor/src/replication/models"
)

// Interface is certain mechanism to generate JobData.
type Interface interface {
	Generate(policy *models.P2PPreheatPolicy, triggerItems []filter_models.FilterItem) (*models.JobData, error)
}
