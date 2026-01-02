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

	"gorm.io/gorm"

	searchEntity "github.com/coze-dev/coze-studio/backend/domain/search/entity"
)

type SearchProjectsDAO struct {
	db *gorm.DB
}

func NewSearchProjectsDAO(db *gorm.DB) *SearchProjectsDAO {
	return &SearchProjectsDAO{db: db}
}

// SearchProjectRow represents a row in the unified search result
type SearchProjectRow struct {
	ID          int64
	Type        int32
	OwnerID     int64
	SpaceID     int64
	Name        string
	Description string
	IconURI     string
	CreatedAt   int64
	UpdatedAt   int64
}

// SearchProjects queries both single_agent_draft and app_draft tables and returns unified results
func (dao *SearchProjectsDAO) SearchProjects(ctx context.Context, req *searchEntity.SearchProjectsRequest) ([]*SearchProjectRow, error) {
	var results []*SearchProjectRow

	// Query bot drafts (type = 1)
	botQuery := dao.db.Table("single_agent_draft").
		Select("agent_id as id, 1 as type, creator_id as owner_id, space_id, name, description, icon_uri, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL", req.SpaceID)

	// Query app drafts (type = 2)
	appQuery := dao.db.Table("app_draft").
		Select("id, 2 as type, owner_id, space_id, name, description, icon_uri, created_at, updated_at").
		Where("space_id = ? AND deleted_at IS NULL", req.SpaceID)

	// UNION ALL to combine both queries and order by updated_at DESC
	unionQuery := dao.db.Raw("(?) UNION ALL (?) ORDER BY updated_at DESC", botQuery, appQuery)

	err := unionQuery.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}
