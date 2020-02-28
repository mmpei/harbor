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

package event

import (
	"errors"
	"fmt"
	"reflect"

	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/event/notification"
	"github.com/goharbor/harbor/src/replication/event/topic"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication/trigger"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/webhook/controller"
	"github.com/goharbor/harbor/src/webhook"
	p2p_controller "github.com/goharbor/harbor/src/p2ppreheat/controller"
)

// OnPushHandler implements the notification handler interface to handle image on push event.
type OnPushHandler struct{}

// Handle implements the same method of notification handler interface
func (oph *OnPushHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("OnPushHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.OnPushNotification" {
		return fmt.Errorf("Mismatch value type of OnPushHandler, expect %s but got %s", "notification.OnPushNotification", vType.String())
	}

	notification := value.(notification.OnPushNotification)

	handleResult := true
	msg := fmt.Sprintf("image %s: ", notification.Image)
	err := checkAndTriggerReplication(notification.Image, common_models.RepOpTransfer)
	if err != nil {
		handleResult = false
		log.Error(err)
		msg += "trigger replication failed;"
	}

	err = checkAndTriggerWebhook(notification.Image)
	if err != nil {
		handleResult = false
		log.Error(err)
		msg += "trigger webhook failed;"
	}

	err = checkAndTriggerP2PPreheat(notification.Image)
	if err != nil {
		handleResult = false
		log.Error(err)
		msg += "trigger p2p preheat failed;"
	}

	if handleResult {
		return nil
	} else {
		return fmt.Errorf("%s", msg)
	}
}

// IsStateful implements the same method of notification handler interface
func (oph *OnPushHandler) IsStateful() bool {
	// Statless
	return false
}

// checks whether replication policy is set on the resource, if is, trigger the replication
func checkAndTriggerReplication(image, operation string) error {
	project, _ := utils.ParseRepository(image)
	watchItems, err := trigger.DefaultWatchList.Get(project, operation)
	if err != nil {
		return fmt.Errorf("failed to get watch list for resource %s, operation %s: %v",
			image, operation, err)
	}
	if len(watchItems) == 0 {
		log.Debugf("no replication should be triggered for resource %s, operation %s, skip", image, operation)
		return nil
	}

	for _, watchItem := range watchItems {
		item := models.FilterItem{
			Kind:      replication.FilterItemKindTag,
			Value:     image,
			Operation: operation,
		}

		if err := notifier.Publish(topic.StartReplicationTopic, notification.StartReplicationNotification{
			PolicyID: watchItem.PolicyID,
			Metadata: map[string]interface{}{
				"candidates": []models.FilterItem{item},
			},
		}); err != nil {
			return fmt.Errorf("failed to publish replication topic for resource %s, operation %s, policy %d: %v",
				image, operation, watchItem.PolicyID, err)
		}
		log.Infof("replication topic for resource %s, operation %s, policy %d triggered",
			image, operation, watchItem.PolicyID)
	}
	return nil
}

func checkAndTriggerWebhook(image string) error {
	project, _ := utils.ParseRepository(image)
	prj, err := config.GlobalProjectMgr.Get(project)
	if err != nil {
		return fmt.Errorf("failed to get project %s, image %s: %v", project, image, err)
	}

	policies, err := controller.PolicyManager.GetPolicies(prj.ProjectID, "")
	if err != nil {
		return fmt.Errorf("failed to get webhook policies projectID %d, image %s: %v", prj.ProjectID, image, err)
	}

	if len(policies) == 0 {
		return nil
	}

	for _, policy := range policies {
		if len(policy.HookTypes) == 0 {
			log.Warningf("webhook policy doesn't contain any hook type, policyID: %d", policy.ID)
			continue
		}
		shouldTrigger := false
		for _, hookType := range policy.HookTypes {
			if hookType == webhook.ImagePushEvent {
				shouldTrigger = true
				break
			}
		}

		if !shouldTrigger {
			continue
		}

		item := models.FilterItem{
			Kind:      replication.FilterItemKindTag,
			Value:     image,
			Operation: webhook.ImagePushEvent,
		}
		if err := notifier.Publish(topic.StartWebhookTopic, notification.StartWebhookNotification{
			Policy: policy,
			Metadata: map[string]interface{}{
				"candidates": []models.FilterItem{item},
			},
		}); err != nil {
			return fmt.Errorf("failed to publish webhook topic for resource %s, policy %d: %v",
				image, policy.ID, err)
		}
		log.Infof("webhook topic for resource %s, policy %d triggered", image, policy.ID)
	}
	return nil
}

func checkAndTriggerP2PPreheat(image string) error {
	project, _ := utils.ParseRepository(image)
	prj, err := config.GlobalProjectMgr.Get(project)
	if err != nil {
		return fmt.Errorf("failed to get project %s, image %s: %v", project, image, err)
	}

	policies, _, err := p2p_controller.PolicyManager.GetPolicies(prj.ProjectID, "", 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get p2p preheat policies projectID %d, image %s: %v", prj.ProjectID, image, err)
	}

	if len(policies) == 0 {
		return nil
	}

	for _, policy := range policies {
		item := models.FilterItem{
			Kind:  replication.FilterItemKindTag,
			Value: image,
		}
		if err := notifier.Publish(topic.StartP2PPreheatTopic, notification.StartP2PPreheatNotification{
			Policy: policy,
			Metadata: map[string]interface{}{
				"candidates": []models.FilterItem{item},
			},
		}); err != nil {
			return fmt.Errorf("failed to publish p2p preheat topic for resource %s, policy %d: %v",
				image, policy.ID, err)
		}
		log.Infof("p2p preheat topic for resource %s, policy %d triggered", image, policy.ID)
	}
	return nil
}
