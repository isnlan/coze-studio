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

package dal

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	resource "github.com/coze-dev/coze-studio/backend/api/model/resource/common"
	searchEntity "github.com/coze-dev/coze-studio/backend/domain/search/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

type SearchResourcesDAO struct {
	db *gorm.DB
}

func NewSearchResourcesDAO(db *gorm.DB) *SearchResourcesDAO {
	return &SearchResourcesDAO{db: db}
}

// SearchResourceRow represents a row in the unified resource search result
type SearchResourceRow struct {
	ID          int64
	Type        int32
	OwnerID     int64
	SpaceID     int64
	APPID       *int64
	Name        string
	Description string
	IconURI     string
	Status      int64
	SubType     *int32
	CreatedAt   int64
	UpdatedAt   int64
}

// SearchResources queries multiple resource tables and returns unified results
func (dao *SearchResourcesDAO) SearchResources(ctx context.Context, req *searchEntity.SearchResourcesRequest) ([]*SearchResourceRow, error) {
	// Determine which resource types to query
	typesToQuery := dao.getResTypesToQuery(req.ResTypeFilter)
	if len(typesToQuery) == 0 {
		logs.CtxDebugf(ctx, "[SearchResourcesDAO] No resource types to query")
		return []*SearchResourceRow{}, nil
	}

	logs.CtxDebugf(ctx, "[SearchResourcesDAO] Querying resource types: %v", typesToQuery)

	// Build subqueries for each type
	var subqueries []*gorm.DB
	for _, resType := range typesToQuery {
		subquery := dao.buildSubqueryForType(ctx, req, resType)
		if subquery != nil {
			subqueries = append(subqueries, subquery)
		}
	}

	if len(subqueries) == 0 {
		logs.CtxDebugf(ctx, "[SearchResourcesDAO] No valid subqueries built")
		return []*SearchResourceRow{}, nil
	}

	// Combine with UNION ALL
	unionQuery := dao.combineWithUnion(subqueries)

	// Execute query
	var results []*SearchResourceRow
	err := unionQuery.Scan(&results).Error
	if err != nil {
		logs.CtxErrorf(ctx, "[SearchResourcesDAO] Query failed: %v", err)
		return nil, err
	}

	logs.CtxDebugf(ctx, "[SearchResourcesDAO] Found %d resources", len(results))
	return results, nil
}

func (dao *SearchResourcesDAO) getResTypesToQuery(filter []resource.ResType) []resource.ResType {
	// If filter is empty, return all types
	if len(filter) == 0 {
		return dao.getAllSupportedTypes()
	}

	// Check if filter contains -1 (special value for "all types")
	for _, resType := range filter {
		if resType == -1 {
			return dao.getAllSupportedTypes()
		}
	}

	return filter
}

// getAllSupportedTypes returns all resource types supported for search
func (dao *SearchResourcesDAO) getAllSupportedTypes() []resource.ResType {
	return []resource.ResType{
		resource.ResType_Workflow,
		resource.ResType_Imageflow,
		resource.ResType_Prompt,
		resource.ResType_Plugin,
		resource.ResType_Knowledge,
		resource.ResType_Database,
	}
}

func (dao *SearchResourcesDAO) buildSubqueryForType(ctx context.Context, req *searchEntity.SearchResourcesRequest, resType resource.ResType) *gorm.DB {
	switch resType {
	case resource.ResType_Workflow:
		return dao.buildWorkflowQuery(req)
	case resource.ResType_Imageflow:
		return dao.buildImageflowQuery(req)
	case resource.ResType_Prompt:
		return dao.buildPromptQuery(req)
	case resource.ResType_Plugin:
		return dao.buildPluginQuery(req)
	case resource.ResType_Knowledge:
		return dao.buildKnowledgeQuery(req)
	case resource.ResType_Database:
		return dao.buildDatabaseQuery(req)
	default:
		logs.CtxWarnf(ctx, "[SearchResourcesDAO] Unsupported resource type: %d", resType)
		return nil
	}
}

func (dao *SearchResourcesDAO) buildWorkflowQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	query := dao.db.Table("workflow_meta").
		Select("id, 2 as type, creator_id as owner_id, space_id, app_id, "+
			"CAST(name AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(description AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST(icon_uri AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"status, mode as sub_type, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL AND mode = 0", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("creator_id = ?", req.OwnerID)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.APPID > 0 {
		query = query.Where("app_id = ?", req.APPID)
	}
	if req.PublishStatusFilter > 0 {
		// workflow_meta.status: 0=unpublished, 1=published
		if req.PublishStatusFilter == resource.PublishStatus_Published {
			query = query.Where("status = ?", 1)
		} else if req.PublishStatusFilter == resource.PublishStatus_UnPublished {
			query = query.Where("status = ?", 0)
		}
	}

	return query
}

func (dao *SearchResourcesDAO) buildImageflowQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	query := dao.db.Table("workflow_meta").
		Select("id, 3 as type, creator_id as owner_id, space_id, app_id, "+
			"CAST(name AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(description AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST(icon_uri AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"status, mode as sub_type, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL AND mode = 1", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("creator_id = ?", req.OwnerID)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.APPID > 0 {
		query = query.Where("app_id = ?", req.APPID)
	}
	if req.PublishStatusFilter > 0 {
		// workflow_meta.status: 0=unpublished, 1=published
		if req.PublishStatusFilter == resource.PublishStatus_Published {
			query = query.Where("status = ?", 1)
		} else if req.PublishStatusFilter == resource.PublishStatus_UnPublished {
			query = query.Where("status = ?", 0)
		}
	}

	return query
}

func (dao *SearchResourcesDAO) buildPromptQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	// Prompt resources don't belong to any app, so skip if APPID is specified
	if req.APPID > 0 {
		return nil
	}

	query := dao.db.Table("prompt_resource").
		Select("id, 6 as type, creator_id as owner_id, space_id, NULL as app_id, "+
			"CAST(name AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(description AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST('' AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"status, NULL as sub_type, created_at, updated_at").
		Where("space_id = ? AND status = 1", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("creator_id = ?", req.OwnerID)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	return query
}

func (dao *SearchResourcesDAO) buildPluginQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	query := dao.db.Table("plugin_draft").
		Select("id, 1 as type, developer_id as owner_id, space_id, app_id, "+
			"CAST(COALESCE(JSON_UNQUOTE(JSON_EXTRACT(manifest, '$.name_for_human')), 'Unnamed Plugin') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(COALESCE(JSON_UNQUOTE(JSON_EXTRACT(manifest, '$.description_for_human')), '') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST(icon_uri AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"1 as status, NULL as sub_type, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("developer_id = ?", req.OwnerID)
	}
	if req.APPID > 0 {
		query = query.Where("app_id = ?", req.APPID)
	}

	return query
}

func (dao *SearchResourcesDAO) buildKnowledgeQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	query := dao.db.Table("knowledge").
		Select("id, 4 as type, creator_id as owner_id, space_id, app_id, "+
			"CAST(name AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(COALESCE(description, '') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST(COALESCE(icon_uri, '') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"status, format_type as sub_type, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("creator_id = ?", req.OwnerID)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.APPID > 0 {
		query = query.Where("app_id = ?", req.APPID)
	}

	return query
}

func (dao *SearchResourcesDAO) buildDatabaseQuery(req *searchEntity.SearchResourcesRequest) *gorm.DB {
	query := dao.db.Table("draft_database_info").
		Select("id, 7 as type, creator_id as owner_id, space_id, app_id, "+
			"CAST(table_name AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as name, "+
			"CAST(COALESCE(table_desc, '') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as description, "+
			"CAST(COALESCE(icon_uri, '') AS CHAR CHARACTER SET utf8mb4) COLLATE utf8mb4_unicode_ci as icon_uri, "+
			"is_visible as status, NULL as sub_type, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL", req.SpaceID)

	// Apply optional filters
	if req.OwnerID > 0 {
		query = query.Where("creator_id = ?", req.OwnerID)
	}
	if req.APPID > 0 {
		query = query.Where("app_id = ?", req.APPID)
	}

	return query
}

func (dao *SearchResourcesDAO) combineWithUnion(subqueries []*gorm.DB) *gorm.DB {
	if len(subqueries) == 0 {
		return nil
	}

	if len(subqueries) == 1 {
		return dao.db.Raw("(?) ORDER BY updated_at DESC", subqueries[0])
	}

	// Build UNION ALL query
	queryParts := make([]string, len(subqueries))
	queryArgs := make([]interface{}, len(subqueries))

	for i := range subqueries {
		queryParts[i] = "(?)"
		queryArgs[i] = subqueries[i]
	}

	unionSQL := fmt.Sprintf("%s ORDER BY updated_at DESC", joinWithUnionAll(queryParts))
	return dao.db.Raw(unionSQL, queryArgs...)
}

func joinWithUnionAll(parts []string) string {
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " UNION ALL " + parts[i]
	}
	return result
}
