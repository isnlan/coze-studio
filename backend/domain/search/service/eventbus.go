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

package service

import (
	"context"

	"github.com/coze-dev/coze-studio/backend/domain/search/entity"
)

// Stub implementations for backward compatibility
// TODO: Remove these along with EventBus interfaces after cleaning up all references

type projectEventBusStub struct{}

func NewProjectEventBus() ProjectEventBus {
	return &projectEventBusStub{}
}

func (p *projectEventBusStub) PublishProject(ctx context.Context, event *entity.ProjectDomainEvent) error {
	// No-op - ES sync removed
	return nil
}

type resourceEventBusStub struct{}

func NewResourceEventBus() ResourceEventBus {
	return &resourceEventBusStub{}
}

func (r *resourceEventBusStub) PublishResources(ctx context.Context, event *entity.ResourceDomainEvent) error {
	// No-op - ES sync removed
	return nil
}
