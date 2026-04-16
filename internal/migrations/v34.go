/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package migrations

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addQuestionSectionTagID(ctx context.Context, x *xorm.Engine) error {
	type Question struct {
		SectionTagID string `xorm:"not null default '' BIGINT(20) INDEX section_tag_id"`
	}

	if err := x.Context(ctx).Sync(new(Question)); err != nil {
		return fmt.Errorf("sync question section_tag_id column: %w", err)
	}

	// Migrate existing data: for questions that have a section tag in tag_rel,
	// copy the section tag ID to the new section_tag_id column.
	// A section tag is identified by having a forum.section.visibility meta entry.
	_, err := x.Context(ctx).Exec(`
		UPDATE question SET section_tag_id = (
			SELECT tr.tag_id FROM tag_rel tr
			INNER JOIN meta m ON m.object_id = tr.tag_id AND m.key = ?
			WHERE tr.object_id = question.id AND tr.status = 1
			LIMIT 1
		) WHERE EXISTS (
			SELECT 1 FROM tag_rel tr
			INNER JOIN meta m ON m.object_id = tr.tag_id AND m.key = ?
			WHERE tr.object_id = question.id AND tr.status = 1
		)
	`, entity.ForumSectionVisibility, entity.ForumSectionVisibility)
	if err != nil {
		return fmt.Errorf("migrate section tag data: %w", err)
	}

	// Remove section tags from tag_rel — sections are no longer stored as tags.
	_, err = x.Context(ctx).Exec(`
		DELETE FROM tag_rel WHERE tag_id IN (
			SELECT object_id FROM meta WHERE key = ?
		)
	`, entity.ForumSectionVisibility)
	if err != nil {
		return fmt.Errorf("remove section tags from tag_rel: %w", err)
	}

	return nil
}
