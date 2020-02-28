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

package notification

import (
	"github.com/goharbor/harbor/src/webhook/models"
	p2p_models "github.com/goharbor/harbor/src/p2ppreheat/models"
)

// OnPushNotification contains the data required by this handler
type OnPushNotification struct {
	// The name of the image that is being pushed
	Image string
}

// OnDeletionNotification contains the data required by this handler
type OnDeletionNotification struct {
	// The name of the image that is being deleted
	Image string
}

// StartReplicationNotification contains data required by this handler
type StartReplicationNotification struct {
	// ID of the policy
	PolicyID int64
	Metadata map[string]interface{}
}

// StartWebhookNotification contains data required by this handler
type StartWebhookNotification struct {
	// the policy trigger the notification
	Policy *models.WebhookPolicy
	Metadata map[string]interface{}
}

// StartP2PPreheatNotification contains data required by this handler
type StartP2PPreheatNotification struct {
	// the policy trigger the notification
	Policy *p2p_models.P2PPreheatPolicy
	Metadata map[string]interface{}
}
