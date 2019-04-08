package hook

import (
	"github.com/goharbor/harbor/src/webhook/models"
	filter_models "github.com/goharbor/harbor/src/replication/models"
	common_models "github.com/goharbor/harbor/src/common/models"
	job_models "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/webhook/filter"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/core/config"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/core/utils"
	"encoding/json"
	"github.com/goharbor/harbor/src/webhook/job"
)

// Manager ...
type Manager interface {
	StartHook(policy *models.WebhookPolicy, metadata ...map[string]interface{}) error
}

type HookManager struct {
	jobManager  job.Manager
	filter      *hook.FilterManager
	client      common_job.Client
}

// NewHookManager is the constructor of HookManager.
func NewHookManager() *HookManager {
	return  &HookManager{
		jobManager: job.NewDefaultManager(),
		filter: hook.NewFilterManager(),
		client: utils.GetJobServiceClient(),
	}
}

// HookManager is used to generate a webhook job and trigger the webhook execution.
func (hm *HookManager) StartHook(policy *models.WebhookPolicy, metadata ...map[string]interface{}) error {
	// do filter
	candidates := []filter_models.FilterItem{}
	if len(metadata) > 0 {
		meta := metadata[0]["candidates"]
		if meta != nil {
			cands, ok := meta.([]filter_models.FilterItem)
			if ok {
				candidates = append(candidates, cands...)
			}
		}
	}

	if len(candidates) == 0 {
		log.Warningf("webhook candidates are null, trigger nothing")
		return nil
	}

	triggerItems, err := hm.filter.DoFilter(policy, candidates)
	if err != nil {
		log.Errorf("webhook do filter error: %+v", err)
		return err
	}

	if len(triggerItems) == 0 {
		log.Info("trigger items is null, trigger nothing")
		return nil
	}

	// generate jobs store then send to JS
	if policy.ProjectID == 0 { // registry level, not support now

	} else { // the triggered item in one project
		jobData, err := hm.jobManager.Generate(policy, triggerItems)
		if err != nil {
			return fmt.Errorf("failed to get job data: %v", err)
		}

		str, err := json.Marshal(jobData)
		if err != nil {
			return fmt.Errorf("failed to marshal job data error, %v", err)
		}
		jobDetail := string(str)
		hookType := triggerItems[0].Operation
		id, err := dao.AddWebhookJob(common_models.WebhookJob{
			Status:      common_models.JobPending,
			PolicyID:    policy.ID,
			HookType:    hookType,
			JobDetail:   jobDetail,
		})
		if err != nil {
			return err
		}

		// submit job to jobservice
		log.Debugf("submiting webhook job to jobservice, hook type: %s, policy: %v, job data: %s",
			triggerItems[0].Operation, policy, string(str))
		job := &job_models.JobData{
			Metadata: &job_models.JobMetadata{
				JobKind: common_job.JobKindGeneric,
			},
			StatusHook: fmt.Sprintf("%s/service/notifications/jobs/webhook/%d",
				config.InternalCoreURL(), id),
		}

		job.Name = common_job.ImageWebhook
		job.Parameters = map[string]interface{}{
			"hook_type":             hookType,
			"job_detail":            jobDetail,
			"target":                policy.Target,
		}

		uuid, err := hm.client.SubmitJob(job)
		if err != nil {
			if er := dao.UpdateWebhookJobStatus(id, common_models.JobError); er != nil {
				log.Errorf("failed to update the status of webhook job %d: %s", id, er)
			}
			return err
		}

		// create the mapping relationship between the jobs in database and jobservice
		if err = dao.SetWebhookJobUUID(id, uuid); err != nil {
			return err
		}
	}
	return nil
}
