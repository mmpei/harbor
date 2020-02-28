package job

import (
   "github.com/goharbor/harbor/src/p2ppreheat/models"
   filter_models "github.com/goharbor/harbor/src/replication/models"
   "fmt"
   "strings"
   "github.com/goharbor/harbor/src/common/utils"
   "github.com/goharbor/harbor/src/core/config"
   "time"
   common_models "github.com/goharbor/harbor/src/common/models"
   "github.com/goharbor/harbor/src/common/utils/log"
   "github.com/docker/distribution/manifest/schema2"
   coreutils "github.com/goharbor/harbor/src/core/utils"
   "errors"
)

// Manager defines the method a job manger should implement
type Manager interface {
   Generate(policy *models.P2PPreheatPolicy, triggerItems []filter_models.FilterItem) (*common_models.Notification, error)
}

// DefaultManager ...
type DefaultManager struct{
}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// Generate returns the p2p preheat request body data(job data)
func (dm *DefaultManager) Generate(policy *models.P2PPreheatPolicy, triggerItems []filter_models.FilterItem) (*common_models.Notification, error) {
   if policy == nil {
      return nil, fmt.Errorf("generate job data: policy is nil")
   }
   if len(triggerItems) == 0 {
      return nil, fmt.Errorf("generate job data: trigger items are nil")
   }

   var events = []common_models.Event{}
   for _, item := range triggerItems {
      strs := strings.SplitN(item.Value, ":", 2)
      if len(strs) != 2 {
         log.Warningf("invalid image '%s'", item.Value)
         continue
      }
      digest, err := getDigest(strs[0], strs[1])
      if err != nil {
         return nil, err
      }
      extURL, _ := config.ExtURL()
      event := common_models.Event{
         ID: utils.GenerateRandomString(),
         TimeStamp: time.Now().UTC(),
         Action: "push",
         Target: &common_models.Target{
            MediaType: schema2.MediaTypeManifest,
            Digest: digest,
            Repository: strs[0],
            URL: fmt.Sprintf("%s/%s", extURL, item.Value),
            Tag: strs[1],
         },
      }
      events = append(events, event)
   }
   var payload = common_models.Notification{
      Events: events,
   }
   return &payload, nil
}

func getDigest(repoName, tag string) (string, error) {
   client, err := coreutils.NewRepositoryClientForUI("harbor-core", repoName)
   if err != nil {
      return "", err
   }
   digest, exist, err := client.ManifestExist(tag)
   if err != nil {
      return "", err
   }
   if !exist {
      return "", errors.New("digest not found")
   }
   return digest, nil
}
