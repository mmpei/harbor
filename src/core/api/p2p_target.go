// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/p2ppreheat/controller"
)

// P2PTargetAPI handles request to /api/p2p/targets/ping /api/p2p/targets/{}
type P2PTargetAPI struct {
	BaseController
	secretKey string
}

// Prepare validates the user
func (t *P2PTargetAPI) Prepare() {
	t.BaseController.Prepare()
	if !t.SecurityCtx.IsAuthenticated() {
		t.HandleUnauthorized()
		return
	}

	if !t.SecurityCtx.IsSysAdmin() {
		t.HandleForbidden(t.SecurityCtx.GetUsername())
		return
	}

	var err error
	t.secretKey, err = config.SecretKey()
	if err != nil {
		log.Errorf("failed to get secret key: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Get ...
func (t *P2PTargetAPI) Get() {
	id := t.GetIDFromURL()

	target, err := dao.GetP2PTarget(id)
	if err != nil {
		log.Errorf("failed to get p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.HandleNotFound(fmt.Sprintf("p2p target %d not found", id))
		return
	}

	target.Password = ""

	t.Data["json"] = target
	t.ServeJSON()
}

// List ...
func (t *P2PTargetAPI) List() {
	name := t.GetString("name")
	targets, err := dao.FilterP2PTargets(name)
	if err != nil {
		log.Errorf("failed to filter p2p targets %s: %v", name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, target := range targets {
		target.Password = ""
	}

	t.Data["json"] = targets
	t.ServeJSON()
	return
}

// Post ...
func (t *P2PTargetAPI) Post() {
	target := &models.P2PTarget{}
	t.DecodeJSONReqAndValidate(target)

	ta, err := dao.GetP2PTargetByName(target.Name)
	if err != nil {
		log.Errorf("failed to get p2p target %s: %v", target.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.HandleConflict("name is already used")
		return
	}

	ta, err = dao.GetP2PTargetByEndpoint(target.URL)
	if err != nil {
		log.Errorf("failed to get p2p target [ %s ]: %v", target.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.HandleConflict(fmt.Sprintf("the p2p target whose endpoint is %s already exists", target.URL))
		return
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	id, err := dao.AddP2PTarget(*target)
	if err != nil {
		log.Errorf("failed to add p2p target: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (t *P2PTargetAPI) Put() {
	id := t.GetIDFromURL()

	target, err := dao.GetP2PTarget(id)
	if err != nil {
		log.Errorf("failed to get p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.HandleNotFound(fmt.Sprintf("target %d not found", id))
		return
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	req := struct {
		Name     *string `json:"name"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	t.DecodeJSONReq(&req)

	originalName := target.Name
	originalURL := target.URL

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Endpoint != nil {
		target.URL = *req.Endpoint
	}
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Password != nil {
		target.Password = *req.Password
	}
	if req.Insecure != nil {
		target.Insecure = *req.Insecure
	}

	t.Validate(target)

	if target.Name != originalName {
		ta, err := dao.GetP2PTargetByName(target.Name)
		if err != nil {
			log.Errorf("failed to get p2p target %s: %v", target.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.HandleConflict("name is already used")
			return
		}
	}

	if target.URL != originalURL {
		ta, err := dao.GetP2PTargetByEndpoint(target.URL)
		if err != nil {
			log.Errorf("failed to get p2p target [ %s ]: %v", target.URL, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.HandleConflict(fmt.Sprintf("the p2p target whose endpoint is %s already exists", target.URL))
			return
		}
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	if err := dao.UpdateP2PTarget(*target); err != nil {
		log.Errorf("failed to update p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Delete ...
func (t *P2PTargetAPI) Delete() {
	id := t.GetIDFromURL()

	target, err := dao.GetP2PTarget(id)
	if err != nil {
		log.Errorf("failed to get p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.HandleNotFound(fmt.Sprintf("p2p target %d not found", id))
		return
	}

	policies, err := dao.GetP2PPreheatPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if len(policies) > 0 {
		log.Error("the p2p target is used by policies, can not be deleted")
		t.CustomAbort(http.StatusPreconditionFailed, "the target is used by policies, can not be deleted")
	}

	if err = dao.DeleteP2PTarget(id); err != nil {
		log.Errorf("failed to delete p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// ListPolicies ...
func (t *P2PTargetAPI) ListPolicies() {
	id := t.GetIDFromURL()

	target, err := dao.GetP2PTarget(id)
	if err != nil {
		log.Errorf("failed to get p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.HandleNotFound(fmt.Sprintf("p2p target %d not found", id))
		return
	}

	policies, err := controller.PolicyManager.GetPoliciesByTargetId(id)
	if err != nil {
		log.Errorf("failed to get policies according p2p target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, policy := range policies {
		for _, target := range policy.Targets {
			target.Password = ""
		}
	}

	t.Data["json"] = policies
	t.ServeJSON()
}
