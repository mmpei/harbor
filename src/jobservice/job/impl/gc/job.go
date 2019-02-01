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

package gc

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/common/utils"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/registryctl/client"
	"github.com/goharbor/harbor/src/common/job"
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	blobPrefix            = "blobs::*"
	repoPrefix            = "repository::*"
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	registryCtlClient client.Client
	logger            logger.Interface
	coreclient        *common_http.Client
	CoreURL           string
	insecure          bool
	redisURL          string
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// Priority implements the interface in job/Interface
func (gc *GarbageCollector) Priority() uint {
	return job.JobPriorityNormal
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(params map[string]interface{}) error {
	return nil
}

// Run implements the interface in job/Interface
func (gc *GarbageCollector) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}
	readOnlyCur, err := gc.getReadOnly()
	if err != nil {
		return err
	}
	if readOnlyCur != true {
		if err := gc.setReadOnly(true); err != nil {
			return err
		}
		defer gc.setReadOnly(readOnlyCur)
	}
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as registry controller is unreachable: %v", err)
		return err
	}
	gc.logger.Infof("start to run gc in job.")
	gcr, err := gc.registryCtlClient.StartGC()
	if err != nil {
		gc.logger.Errorf("failed to get gc result: %v", err)
		return err
	}
	if err := gc.cleanCache(); err != nil {
		return err
	}
	gc.logger.Infof("GC results: status: %t, message: %s, start: %s, end: %s.", gcr.Status, gcr.Msg, gcr.StartTime, gcr.EndTime)
	gc.logger.Infof("success to run gc in job.")
	return nil
}

func (gc *GarbageCollector) init(ctx env.JobContext, params map[string]interface{}) error {
	registryctl.Init()
	gc.registryCtlClient = registryctl.RegistryCtlClient
	gc.logger = ctx.GetLogger()
	cred := auth.NewSecretAuthorizer(os.Getenv("JOBSERVICE_SECRET"))
	gc.insecure = false
	gc.coreclient = common_http.NewClient(&http.Client{
		Transport: reg.GetHTTPTransport(gc.insecure),
	}, cred)
	errTpl := "Failed to get required property: %s"
	if v, ok := ctx.Get(common.CoreURL); ok && len(v.(string)) > 0 {
		gc.CoreURL = v.(string)
	} else {
		return fmt.Errorf(errTpl, common.CoreURL)
	}
	gc.redisURL = params["redis_url_reg"].(string)
	return nil
}

func (gc *GarbageCollector) getReadOnly() (bool, error) {
	cfgs := map[string]interface{}{}
	if err := gc.coreclient.Get(fmt.Sprintf("%s/api/configs", gc.CoreURL), &cfgs); err != nil {
		return false, err
	}
	return utils.SafeCastBool(cfgs[common.ReadOnly]), nil
}

func (gc *GarbageCollector) setReadOnly(switcher bool) error {
	if err := gc.coreclient.Put(fmt.Sprintf("%s/api/configurations", gc.CoreURL), struct {
		ReadOnly bool `json:"read_only"`
	}{
		ReadOnly: switcher,
	}); err != nil {
		gc.logger.Errorf("failed to send readonly request to %s: %v", gc.CoreURL, err)
		return err
	}
	gc.logger.Info("the readonly request has been sent successfully")
	return nil
}

// cleanCache is to clean the registry cache for GC.
// To do this is because the issue https://github.com/docker/distribution/issues/2094
func (gc *GarbageCollector) cleanCache() error {

	con, err := redis.DialURL(
		gc.redisURL,
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)

	if err != nil {
		gc.logger.Errorf("failed to connect to redis %v", err)
		return err
	}
	defer con.Close()

	// clean all keys in registry redis DB.

	// sample of keys in registry redis:
	// 1) "blobs::sha256:1a6fd470b9ce10849be79e99529a88371dff60c60aab424c077007f6979b4812"
	// 2) "repository::library/hello-world::blobs::sha256:4ab4c602aa5eed5528a6620ff18a1dc4faef0e1ab3a5eddeddb410714478c67f"
	err = delKeys(con, blobPrefix)
	if err != nil {
		gc.logger.Errorf("failed to clean registry cache %v, pattern blobs::*", err)
		return err
	}
	err = delKeys(con, repoPrefix)
	if err != nil {
		gc.logger.Errorf("failed to clean registry cache %v, pattern repository::*", err)
		return err
	}

	return nil
}

func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
		}
	}
	return nil
}
