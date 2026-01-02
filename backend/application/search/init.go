/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package search

import (
	"context"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/application/singleagent"
	app "github.com/coze-dev/coze-studio/backend/domain/app/service"
	connector "github.com/coze-dev/coze-studio/backend/domain/connector/service"
	knowledge "github.com/coze-dev/coze-studio/backend/domain/knowledge/service"
	database "github.com/coze-dev/coze-studio/backend/domain/memory/database/service"
	"github.com/coze-dev/coze-studio/backend/domain/plugin/service"
	prompt "github.com/coze-dev/coze-studio/backend/domain/prompt/service"
	search "github.com/coze-dev/coze-studio/backend/domain/search/service"
	user "github.com/coze-dev/coze-studio/backend/domain/user/service"
	"github.com/coze-dev/coze-studio/backend/domain/workflow"
	"github.com/coze-dev/coze-studio/backend/infra/cache"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
)

type ServiceComponents struct {
	DB                   *gorm.DB
	Cache                cache.Cmdable
	TOS                  storage.Storage
	ProjectEventBus      ProjectEventBus
	ResourceEventBus     ResourceEventBus
	SingleAgentDomainSVC singleagent.SingleAgent
	APPDomainSVC         app.AppService
	KnowledgeDomainSVC   knowledge.Knowledge
	PluginDomainSVC      service.PluginService
	WorkflowDomainSVC    workflow.Service
	UserDomainSVC        user.User
	ConnectorDomainSVC   connector.Connector
	PromptDomainSVC      prompt.Prompt
	DatabaseDomainSVC    database.Database
}

func InitService(ctx context.Context, s *ServiceComponents) (*SearchApplicationService, error) {
	searchDomainSVC := search.NewDomainService(ctx, s.DB)

	SearchSVC.DomainSVC = searchDomainSVC
	SearchSVC.ServiceComponents = s

	// Event consumers removed - no longer syncing to Elasticsearch
	// TODO: Remove event consumer registration and handler code

	return SearchSVC, nil
}

type (
	ResourceEventBus = search.ResourceEventBus
	ProjectEventBus  = search.ProjectEventBus
)

func NewResourceEventBus() search.ResourceEventBus {
	return search.NewResourceEventBus()
}

func NewProjectEventBus() search.ProjectEventBus {
	return search.NewProjectEventBus()
}
