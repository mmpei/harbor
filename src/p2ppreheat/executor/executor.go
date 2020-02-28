package executor

import (
	"github.com/goharbor/harbor/src/p2ppreheat/models"
	filter_models "github.com/goharbor/harbor/src/replication/models"
	common_models "github.com/goharbor/harbor/src/common/models"
	job_models "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/core/config"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/core/utils"
	"encoding/json"
	"github.com/goharbor/harbor/src/p2ppreheat/job"
	"github.com/goharbor/harbor/src/p2ppreheat/filter"
	"strings"
)

// Manager ...
type Executor interface {
	StartPreheat(policy *models.P2PPreheatPolicy, metadata ...map[string]interface{}) error
}

type DefaultExecutor struct {
	jobManager  job.Manager
	filter      *filter.FilterManager
	client      common_job.Client
}

// NewExecutor is the constructor of DefaultExecutor.
func NewExecutor() *DefaultExecutor {
	return  &DefaultExecutor{
		jobManager: job.NewDefaultManager(),
		filter: filter.NewFilterManager(),
		client: utils.GetJobServiceClient(),
	}
}

// DefaultExecutor is used to generate a p2p preheat job and trigger the execution.
func (de *DefaultExecutor) StartPreheat(policy *models.P2PPreheatPolicy, metadata ...map[string]interface{}) error {
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
		log.Warningf("p2p preheat candidates are null, trigger nothing")
		return nil
	}

	triggerItems, err := de.filter.DoFilter(policy, candidates)
	if err != nil {
		log.Errorf("p2p preheat do filter error: %+v", err)
		return err
	}

	if len(triggerItems) == 0 {
		log.Info("trigger items is null, trigger nothing")
		return nil
	}

	// generate jobs store then send to JS
	var retErr error
	if policy.Project.ProjectID == 0 { // registry level, not support now

	} else { // the triggered item in one project
		jobData, err := de.jobManager.Generate(policy, triggerItems)
		if err != nil {
			return fmt.Errorf("failed to get job data: %v", err)
		}

		str, err := json.Marshal(jobData)
		if err != nil {
			return fmt.Errorf("failed to marshal job data error, %v", err)
		}

		jobDetail := string(str)
		for _, target := range policy.Targets {
			id, err := dao.AddP2PPreheatJob(common_models.P2PPreheatJob{
				Status:   common_models.JobPending,
				PolicyID: policy.ID,
				Repository: jobData.Events[0].Target.Repository,
				Tag: jobData.Events[0].Target.Tag,
			})
			if err != nil {
				return err
			}

			// submit job to jobservice
			log.Debugf("submiting p2p preheat job to jobservice, policy: %v, job data: %s", policy, string(str))
			job := &job_models.JobData{
				Metadata: &job_models.JobMetadata{
					JobKind: common_job.JobKindGeneric,
				},
				StatusHook: fmt.Sprintf("%s/service/notifications/jobs/p2ppreheat/%d",
					config.InternalCoreURL(), id),
			}

			job.Name = common_job.ImageP2PPreheat
			job.Parameters = map[string]interface{}{
				"job_detail": jobDetail,
				"target":     generateURLByType(target.Type, target.URL),
			}

			uuid, err := de.client.SubmitJob(job)
			if err != nil {
				if er := dao.UpdateP2PPreheatJobStatus(id, common_models.JobError); er != nil {
					log.Errorf("failed to update the status of p2p preheat job %d: %s", id, er)
				}
				retErr = err
				continue
			}

			// create the mapping relationship between the jobs in database and jobservice
			if err = dao.SetP2PPreheatJobUUID(id, uuid); err != nil {
				retErr = err
				continue
			}
		}
	}
	return retErr
}

func generateURLByType(t int, url string) string {
	if t == 0 {// Kraken
		if !strings.HasSuffix(url, "/") {
			url = url + "/"
		}
		return url + "registry/notifications"
	} else { // Others
		return url
	}
}
