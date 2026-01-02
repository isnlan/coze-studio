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

	"gorm.io/gorm"

	common "github.com/coze-dev/coze-studio/backend/api/model/app/intelligence/common"
	resource "github.com/coze-dev/coze-studio/backend/api/model/resource/common"
	searchEntity "github.com/coze-dev/coze-studio/backend/domain/search/entity"
	"github.com/coze-dev/coze-studio/backend/domain/search/internal/dal"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

var searchInstance *searchImpl

func NewDomainService(ctx context.Context, db *gorm.DB) Search {
	return &searchImpl{
		db: db,
	}
}

type searchImpl struct {
	db *gorm.DB
}

func (s *searchImpl) SearchProjects(ctx context.Context, req *searchEntity.SearchProjectsRequest) (resp *searchEntity.SearchProjectsResponse, err error) {
	logs.CtxDebugf(ctx, "[SearchProjects] req=%+v", req)

	// Call DAL layer to query projects
	dao := dal.NewSearchProjectsDAO(s.db)
	rows, err := dao.SearchProjects(ctx, req)
	if err != nil {
		logs.CtxErrorf(ctx, "[SearchProjects] dao error: %v", err)
		return nil, err
	}

	// Convert rows to ProjectDocument
	docs := make([]*searchEntity.ProjectDocument, 0, len(rows))
	for _, row := range rows {
		doc := &searchEntity.ProjectDocument{
			ID:           row.ID,
			Type:         common.IntelligenceType(row.Type),
			Status:       common.IntelligenceStatus_Using, // Drafts are always "Using" status
			Name:         &row.Name,
			SpaceID:      &row.SpaceID,
			OwnerID:      &row.OwnerID,
			CreateTimeMS: &row.CreatedAt,
			UpdateTimeMS: &row.UpdatedAt,
		}
		docs = append(docs, doc)
	}

	resp = &searchEntity.SearchProjectsResponse{
		Data:       docs,
		HasMore:    false,
		NextCursor: "",
	}

	logs.CtxDebugf(ctx, "[SearchProjects] success, returned %d items", len(docs))
	return resp, nil
}

func (s *searchImpl) SearchResources(ctx context.Context, req *searchEntity.SearchResourcesRequest) (resp *searchEntity.SearchResourcesResponse, err error) {
	logs.CtxDebugf(ctx, "[SearchResources] req=%+v", req)

	// Call DAL layer to query resources
	dao := dal.NewSearchResourcesDAO(s.db)
	rows, err := dao.SearchResources(ctx, req)
	if err != nil {
		logs.CtxErrorf(ctx, "[SearchResources] dao error: %v", err)
		return nil, err
	}

	// Convert rows to ResourceDocument
	docs := make([]*searchEntity.ResourceDocument, 0, len(rows))
	for _, row := range rows {
		doc := &searchEntity.ResourceDocument{
			ResID:        row.ID,
			ResType:      resource.ResType(row.Type),
			ResSubType:   row.SubType,
			Name:         &row.Name,
			SpaceID:      &row.SpaceID,
			OwnerID:      &row.OwnerID,
			APPID:        row.APPID,
			BizStatus:    ptr.Of(row.Status),
			CreateTimeMS: &row.CreatedAt,
			UpdateTimeMS: &row.UpdatedAt,
		}

		// Set PublishStatus for workflow based on status field
		if row.Type == int32(resource.ResType_Workflow) {
			if row.Status > 0 {
				doc.PublishStatus = ptr.Of(resource.PublishStatus_Published)
			} else {
				doc.PublishStatus = ptr.Of(resource.PublishStatus_UnPublished)
			}
		}

		docs = append(docs, doc)
	}

	resp = &searchEntity.SearchResourcesResponse{
		Data:       docs,
		TotalHits:  ptr.Of(int64(len(docs))),
		HasMore:    false,
		NextCursor: "",
	}

	return resp, nil
}
