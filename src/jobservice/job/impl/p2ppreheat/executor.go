package p2ppreheat

import (
	"net/http"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"net"
	"bytes"
	"github.com/goharbor/harbor/src/common/job"
    "fmt"
)

// P2PPreheatExecutor is the struct to send preheat request to p2p target
type P2PPreheatExecutor struct {
	client    *http.Client
	logger    logger.Interface
	ctx       env.JobContext
	retry     bool
}

// MaxFails implements the interface in job/Interface
func (pe *P2PPreheatExecutor) MaxFails() uint {
	return 3
}

// ShouldRetry implements the interface in job/Interface
func (pe *P2PPreheatExecutor) ShouldRetry() bool {
	return pe.retry
}

// Priority implements the interface in job/Interface
func (pe *P2PPreheatExecutor) Priority() uint {
	return job.JobPriorityHigh
}

// Validate implements the interface in job/Interface
func (pe *P2PPreheatExecutor) Validate(params map[string]interface{}) error {
	return nil
}

// Run implements the interface in job/Interface
func (pe *P2PPreheatExecutor) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := pe.init(ctx, params); err != nil {
		return err
	}

	err := pe.execute(ctx, params)
	pe.retry = retry(err)
	return err
}

func (pe *P2PPreheatExecutor) init(ctx env.JobContext, params map[string]interface{}) error {
	pe.logger = ctx.GetLogger()
	pe.ctx = ctx

	pe.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	pe.logger.Info("initialization completed.")
	return nil
}

func (pe *P2PPreheatExecutor) execute(ctx env.JobContext, params map[string]interface{}) error {
	jobDetail := params["job_detail"].(string)
	target := params["target"].(string)

	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader([]byte(jobDetail)))
	if err != nil {
		pe.logger.Errorf("p2p preheat job(target: %s, jobData: %s) error: %v", target, jobDetail, err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := pe.client.Do(req)
	if err != nil {
		pe.logger.Errorf("p2p preheat job(target: %s, jobData: %s) error: %v", target, jobDetail, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		pe.logger.Infof("p2p preheat job(target: %s, jobData: %s) send successfully.", target, jobDetail)
		return nil
	} else {
		pe.logger.Errorf("p2p preheat job(target: %s, jobData: %s) response code is %d", target, jobDetail, resp.StatusCode)
        return fmt.Errorf("Response code is %d, p2p preheat event was processed failed in remote endpoint. ", resp.StatusCode)
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
