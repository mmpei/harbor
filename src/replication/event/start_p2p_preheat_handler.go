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

	"github.com/goharbor/harbor/src/replication/event/notification"
	"github.com/goharbor/harbor/src/p2ppreheat/controller"
)

// StartP2PPreheatHandler implements the notification handler interface to handle start p2p preheat requests.
type StartP2PPreheatHandler struct{}

// Handle implements the same method of notification handler interface
func (sph *StartP2PPreheatHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("StartP2PPreheatHandler can not handle nil value")
	}

	vType := reflect.TypeOf(value)
	if vType.Kind() != reflect.Struct || vType.String() != "notification.StartP2PPreheatNotification" {
		return fmt.Errorf("Mismatch value type of StartP2PPreheatHandler, expect %s but got %s",
			"notification.StartP2PPreheatNotification", vType.String())
	}

	notification := value.(notification.StartP2PPreheatNotification)
	if notification.Policy == nil {
		return errors.New("Invalid policy")
	}

	// Start p2p preheat
	return controller.Executor.StartPreheat(notification.Policy , notification.Metadata)
}

// IsStateful implements the same method of notification handler interface
func (sph *StartP2PPreheatHandler) IsStateful() bool {
	// Stateless
	return false
}
