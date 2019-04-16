package webhook

import (
	"net/http"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net"
	"bytes"
	"github.com/goharbor/harbor/src/common/job"
        "fmt"
)

// WebhookExecutor is the struct to send hook data to target
type WebhookExecutor struct {
	client    *http.Client
	logger    logger.Interface
	ctx       env.JobContext
	retry     bool
}

// MaxFails implements the interface in job/Interface
func (we *WebhookExecutor) MaxFails() uint {
	return 3
}

// ShouldRetry implements the interface in job/Interface
func (we *WebhookExecutor) ShouldRetry() bool {
	return we.retry
}

// Priority implements the interface in job/Interface
func (we *WebhookExecutor) Priority() uint {
	return job.JobPriorityHigh
}

// Validate implements the interface in job/Interface
func (we *WebhookExecutor) Validate(params map[string]interface{}) error {
	return nil
}

// Run implements the interface in job/Interface
func (we *WebhookExecutor) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := we.init(ctx, params); err != nil {
		return err
	}

	err := we.execute(ctx, params)
	we.retry = retry(err)
	return err
}

func (we *WebhookExecutor) init(ctx env.JobContext, params map[string]interface{}) error {
	we.logger = ctx.GetLogger()
	we.ctx = ctx

	we.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	we.logger.Info("initialization completed.")
	return nil
}

func (we *WebhookExecutor) execute(ctx env.JobContext, params map[string]interface{}) error {
	jobDetail := params["job_detail"].(string)
	//hookType := params["hook_type"].(string)
	target := params["target"].(string)

	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader([]byte(jobDetail)))
	if err != nil {
		return err
	}
	resp, err := we.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	} else {
		we.logger.Errorf("webhook job(target: %s, jobData: %s) response code is %d", target, jobDetail, resp.StatusCode)
                return fmt.Errorf("Response code is %d, webhook event was processed failed in remote endpoint. ", resp.StatusCode)
	}
	return nil
}

func retry(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(net.Error)
	return ok
}
