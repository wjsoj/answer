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

package content

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/pkg/obj"
	myErrors "github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/net/context"
)

// getSectionMainTagIDBySlug resolves a section slug to the main tag ID.
// If the tag has a MainTagID, it returns that instead.
func (qs *QuestionService) getSectionMainTagIDBySlug(ctx context.Context, slug string) (string, error) {
	tagInfo, exist, err := qs.tagCommon.GetTagBySlugName(ctx, strings.ToLower(slug))
	if err != nil {
		return "", err
	}
	if !exist {
		return "", myErrors.BadRequest(reason.ForumSectionNotFound)
	}
	if tagInfo.MainTagID > 0 {
		return strconv.FormatInt(tagInfo.MainTagID, 10), nil
	}
	return tagInfo.ID, nil
}

// getSectionVisibilityByTagID gets visibility from meta, defaults to "public".
func (qs *QuestionService) getSectionVisibilityByTagID(ctx context.Context, tagID string) string {
	meta, exist, err := qs.metaService.GetMetaByObjectIdAndKeyRaw(ctx, tagID, entity.ForumSectionVisibility)
	if err != nil || !exist {
		return "public"
	}
	if meta.Value == "private" {
		return "private"
	}
	return "public"
}

// getSectionUserSetByKey reads a JSON array of user IDs from meta and returns a set.
func (qs *QuestionService) getSectionUserSetByKey(ctx context.Context, tagID, key string) (map[string]bool, error) {
	result := make(map[string]bool)
	meta, exist, err := qs.metaService.GetMetaByObjectIdAndKeyRaw(ctx, tagID, key)
	if err != nil || !exist {
		return result, nil
	}
	var userIDs []string
	if err := json.Unmarshal([]byte(meta.Value), &userIDs); err != nil {
		return result, nil
	}
	for _, id := range userIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			result[id] = true
		}
	}
	return result, nil
}

// isSectionModeratorByTagID checks if a user is a moderator of the section.
// Global admins are always considered moderators.
func (qs *QuestionService) isSectionModeratorByTagID(ctx context.Context, userID, tagID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	// Check global admin/moderator role
	userRole, err := qs.userRoleRelService.GetUserRole(ctx, userID)
	if err != nil {
		return false, err
	}
	if userRole == role.RoleAdminID || userRole == role.RoleModeratorID {
		return true, nil
	}
	// Check section-level moderator
	moderators, err := qs.getSectionUserSetByKey(ctx, tagID, entity.ForumSectionModerators)
	if err != nil {
		return false, err
	}
	return moderators[userID], nil
}

// canAccessSectionByTagID checks if a user can access a section.
// Public sections: everyone can access.
// Private sections: only admins, moderators, and members.
func (qs *QuestionService) canAccessSectionByTagID(ctx context.Context, userID, tagID string) (bool, error) {
	visibility := qs.getSectionVisibilityByTagID(ctx, tagID)
	if visibility == "public" {
		return true, nil
	}
	if userID == "" {
		return false, nil
	}
	// Check moderator (includes admin)
	isMod, err := qs.isSectionModeratorByTagID(ctx, userID, tagID)
	if err != nil {
		return false, err
	}
	if isMod {
		return true, nil
	}
	// Check member
	members, err := qs.getSectionUserSetByKey(ctx, tagID, entity.ForumSectionMembers)
	if err != nil {
		return false, err
	}
	return members[userID], nil
}

// IsUserAdminOrModerator checks if a user has global admin or moderator role.
func (qs *QuestionService) IsUserAdminOrModerator(ctx context.Context, userID string) bool {
	if userID == "" {
		return false
	}
	userRole, err := qs.userRoleRelService.GetUserRole(ctx, userID)
	if err != nil {
		return false
	}
	return userRole == role.RoleAdminID || userRole == role.RoleModeratorID
}

// CanAccessSectionBySlug checks if a user can access a section by slug.
func (qs *QuestionService) CanAccessSectionBySlug(ctx context.Context, userID, slug string) (bool, error) {
	tagID, err := qs.getSectionMainTagIDBySlug(ctx, slug)
	if err != nil {
		return false, err
	}
	return qs.canAccessSectionByTagID(ctx, userID, tagID)
}

// CanAccessQuestionByID checks if a user can access a question based on its section tags.
// Returns true if the question is accessible, false if any of its tags belong to a private section
// the user cannot access. Fail-closed: returns false on errors.
func (qs *QuestionService) CanAccessQuestionByID(ctx context.Context, userID, questionID string) (bool, error) {
	tags, err := qs.tagCommon.GetObjectTag(ctx, questionID)
	if err != nil {
		return false, err
	}
	for _, tag := range tags {
		slug := tag.SlugName
		if tag.MainTagSlugName != "" {
			slug = tag.MainTagSlugName
		}
		canAccess, err := qs.CanAccessSectionBySlug(ctx, userID, slug)
		if err != nil {
			return false, err
		}
		if !canAccess {
			return false, nil
		}
	}
	return true, nil
}

// CanAccessObjectByID checks if a user can access an object (question, answer, or comment)
// based on section access control. Resolves the parent question and checks section access.
// For tags or unknown types, access is always granted.
func (qs *QuestionService) CanAccessObjectByID(ctx context.Context, userID, objectID string) (bool, error) {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return true, nil // unknown type, allow
	}
	switch objectTypeStr {
	case constant.QuestionObjectType:
		return qs.CanAccessQuestionByID(ctx, userID, objectID)
	case constant.AnswerObjectType:
		answer, exist, err := qs.answerRepo.GetAnswer(ctx, objectID)
		if err != nil || !exist {
			return false, err
		}
		return qs.CanAccessQuestionByID(ctx, userID, answer.QuestionID)
	case constant.CommentObjectType:
		// Comments don't directly encode the parent question; caller should check the parent object.
		return true, nil
	default:
		return true, nil
	}
}

// listSectionTags fetches all tags (sections are top-level tags without MainTagID).
func (qs *QuestionService) listSectionTags(ctx context.Context) ([]*entity.Tag, error) {
	var allTags []*entity.Tag
	page := 1
	pageSize := 100
	for {
		tags, _, err := qs.tagCommon.GetTagPage(ctx, page, pageSize, &entity.Tag{}, "")
		if err != nil {
			return nil, err
		}
		for _, t := range tags {
			if t.MainTagID == 0 && t.Status == entity.TagStatusAvailable {
				allTags = append(allTags, t)
			}
		}
		if len(tags) < pageSize {
			break
		}
		page++
		if page > 100 { // safety limit
			break
		}
	}
	return allTags, nil
}

// GetAccessibleForumSections returns sections the user can access.
func (qs *QuestionService) GetAccessibleForumSections(ctx context.Context, userID string) (*schema.ForumSectionPageResp, error) {
	tags, err := qs.listSectionTags(ctx)
	if err != nil {
		return nil, err
	}

	sections := make([]*schema.ForumSectionResp, 0)
	for _, t := range tags {
		visibility := qs.getSectionVisibilityByTagID(ctx, t.ID)

		canAccess := true
		if visibility == "private" {
			canAccess, err = qs.canAccessSectionByTagID(ctx, userID, t.ID)
			if err != nil {
				return nil, err
			}
		}
		if !canAccess {
			continue
		}

		canManage := false
		if userID != "" {
			canManage, err = qs.isSectionModeratorByTagID(ctx, userID, t.ID)
			if err != nil {
				return nil, err
			}
		}

		canPost := false
		if userID != "" {
			canPost, err = qs.canAccessSectionByTagID(ctx, userID, t.ID)
			if err != nil {
				return nil, err
			}
		}

		sections = append(sections, &schema.ForumSectionResp{
			TagID:       t.ID,
			SlugName:    t.SlugName,
			DisplayName: t.DisplayName,
			Visibility:  visibility,
			CanManage:   canManage,
			CanPost:     canPost,
			IsDefault:   false, // will be set below
		})
	}

	// Mark the first public section as default
	for _, s := range sections {
		if s.Visibility == "public" {
			s.IsDefault = true
			break
		}
	}

	return &schema.ForumSectionPageResp{
		List:  sections,
		Count: int64(len(sections)),
	}, nil
}

// UpdateForumSectionVisibility updates the visibility of a section.
func (qs *QuestionService) UpdateForumSectionVisibility(ctx context.Context, req *schema.ForumSectionVisibilityReq) error {
	tagID, err := qs.getSectionMainTagIDBySlug(ctx, req.Section)
	if err != nil {
		return err
	}

	isMod, err := qs.isSectionModeratorByTagID(ctx, req.UserID, tagID)
	if err != nil {
		return err
	}
	if !isMod {
		return myErrors.Forbidden(reason.ForumSectionPermissionDenied)
	}

	return qs.metaService.AddOrUpdateMetaByObjectIdAndKey(ctx, tagID, entity.ForumSectionVisibility,
		func(meta *entity.Meta, exist bool) (*entity.Meta, error) {
			if !exist {
				meta = &entity.Meta{ObjectID: tagID, Key: entity.ForumSectionVisibility}
			}
			meta.Value = req.Visibility
			return meta, nil
		})
}

// InviteForumSectionUsers adds users to a section role.
func (qs *QuestionService) InviteForumSectionUsers(ctx context.Context, req *schema.ForumSectionInviteReq) (*schema.ForumSectionInviteResp, error) {
	tagID, err := qs.getSectionMainTagIDBySlug(ctx, req.Section)
	if err != nil {
		return nil, err
	}

	// Only admins can grant moderator role
	if req.Role == "moderator" && !req.IsAdmin {
		return nil, myErrors.Forbidden(reason.ForumSectionPermissionDenied)
	}
	// Moderators can grant member role
	if req.Role == "member" {
		isMod, err := qs.isSectionModeratorByTagID(ctx, req.InviterID, tagID)
		if err != nil {
			return nil, err
		}
		if !isMod {
			return nil, myErrors.Forbidden(reason.ForumSectionPermissionDenied)
		}
	}

	// Resolve usernames to user IDs
	userIDs, skipped, err := qs.resolveUserIdentifiers(ctx, req.Users)
	if err != nil {
		return nil, err
	}

	// Always add to members
	if err := qs.addUsersToSectionKey(ctx, tagID, entity.ForumSectionMembers, userIDs); err != nil {
		return nil, err
	}

	// If role is moderator, also add to moderators
	if req.Role == "moderator" {
		if err := qs.addUsersToSectionKey(ctx, tagID, entity.ForumSectionModerators, userIDs); err != nil {
			return nil, err
		}
	}

	return &schema.ForumSectionInviteResp{SkippedUsers: skipped}, nil
}

// RemoveForumSectionUsers removes users from a section role.
func (qs *QuestionService) RemoveForumSectionUsers(ctx context.Context, req *schema.ForumSectionRemoveReq) error {
	tagID, err := qs.getSectionMainTagIDBySlug(ctx, req.Section)
	if err != nil {
		return err
	}

	// Only admins can remove moderators
	if req.Role == "moderator" && !req.IsAdmin {
		return myErrors.Forbidden(reason.ForumSectionPermissionDenied)
	}
	// Moderators can remove members
	if req.Role == "member" {
		isMod, err := qs.isSectionModeratorByTagID(ctx, req.RemoverID, tagID)
		if err != nil {
			return err
		}
		if !isMod {
			return myErrors.Forbidden(reason.ForumSectionPermissionDenied)
		}
	}

	userIDs, _, err := qs.resolveUserIdentifiers(ctx, req.Users)
	if err != nil {
		return err
	}

	return qs.removeUsersFromSectionKey(ctx, tagID, req.Role, userIDs)
}

// GetForumSectionPermissions returns member and moderator lists for a section.
func (qs *QuestionService) GetForumSectionPermissions(ctx context.Context, req *schema.ForumSectionPermissionQueryReq) (*schema.ForumSectionPermissionResp, error) {
	tagID, err := qs.getSectionMainTagIDBySlug(ctx, req.Section)
	if err != nil {
		return nil, err
	}

	isMod, err := qs.isSectionModeratorByTagID(ctx, req.UserID, tagID)
	if err != nil {
		return nil, err
	}
	if !isMod {
		return nil, myErrors.Forbidden(reason.ForumSectionPermissionDenied)
	}

	moderatorNames, err := qs.getSectionUsernamesByKey(ctx, tagID, entity.ForumSectionModerators)
	if err != nil {
		return nil, err
	}

	memberNames, err := qs.getSectionUsernamesByKey(ctx, tagID, entity.ForumSectionMembers)
	if err != nil {
		return nil, err
	}

	// Remove moderators from members list to avoid duplication
	modSet := make(map[string]bool)
	for _, name := range moderatorNames {
		modSet[name] = true
	}
	membersOnly := make([]string, 0)
	for _, name := range memberNames {
		if !modSet[name] {
			membersOnly = append(membersOnly, name)
		}
	}

	return &schema.ForumSectionPermissionResp{
		Members:    membersOnly,
		Moderators: moderatorNames,
	}, nil
}

// getInaccessibleSectionTagIDs returns all tag IDs (including synonyms) of private sections
// the user cannot access. Used for DB-level pre-filtering in question queries.
func (qs *QuestionService) getInaccessibleSectionTagIDs(ctx context.Context, userID string) []string {
	tags, err := qs.listSectionTags(ctx)
	if err != nil {
		log.Errorf("list section tags for exclusion: %v", err)
		return nil
	}

	var excludeIDs []string
	for _, t := range tags {
		visibility := qs.getSectionVisibilityByTagID(ctx, t.ID)
		if visibility != "private" {
			continue
		}
		canAccess, err := qs.canAccessSectionByTagID(ctx, userID, t.ID)
		if err != nil {
			log.Errorf("check section access for exclusion: %v", err)
			// fail-closed: exclude on error
			canAccess = false
		}
		if !canAccess {
			excludeIDs = append(excludeIDs, t.ID)
			// Also exclude synonym tags
			synIDs, err := qs.tagCommon.GetTagIDsByMainTagID(ctx, t.ID)
			if err == nil {
				excludeIDs = append(excludeIDs, synIDs...)
			}
		}
	}
	return excludeIDs
}

// FilterQuestionsBySectionAccess filters questions based on section access.
// Returns only questions the user can access. Fail-closed: on error, the question is excluded.
func (qs *QuestionService) FilterQuestionsBySectionAccess(ctx context.Context, userID string, questions []*schema.QuestionPageResp) []*schema.QuestionPageResp {
	filtered := make([]*schema.QuestionPageResp, 0, len(questions))
	for _, q := range questions {
		accessible := true
		for _, t := range q.Tags {
			sectionTagID := t.ID
			if t.MainTagSlugName != "" {
				// Look up the main tag
				mainTag, exist, err := qs.tagCommon.GetTagBySlugName(ctx, t.MainTagSlugName)
				if err != nil || !exist {
					// fail-closed: exclude if we can't resolve the main tag
					accessible = false
					break
				}
				sectionTagID = mainTag.ID
			}
			canAccess, err := qs.canAccessSectionByTagID(ctx, userID, sectionTagID)
			if err != nil {
				log.Errorf("check section access: %v", err)
				// fail-closed: exclude on error
				accessible = false
				break
			}
			if !canAccess {
				accessible = false
				break
			}
		}
		if accessible {
			filtered = append(filtered, q)
		}
	}
	return filtered
}

// --- helper methods ---

// resolveUserIdentifiers converts usernames to user IDs, returning resolved IDs and skipped usernames.
func (qs *QuestionService) resolveUserIdentifiers(ctx context.Context, identifiers []string) ([]string, []string, error) {
	userIDs := make([]string, 0, len(identifiers))
	skipped := make([]string, 0)
	for _, ident := range identifiers {
		ident = strings.TrimSpace(ident)
		if ident == "" {
			continue
		}
		userInfo, exist, err := qs.userCommon.GetUserBasicInfoByUserName(ctx, ident)
		if err != nil {
			return nil, nil, err
		}
		if !exist {
			skipped = append(skipped, ident)
			continue
		}
		userIDs = append(userIDs, userInfo.ID)
	}
	return userIDs, skipped, nil
}

// addUsersToSectionKey adds user IDs to a meta key (members or moderators).
func (qs *QuestionService) addUsersToSectionKey(ctx context.Context, tagID, key string, newUserIDs []string) error {
	return qs.metaService.AddOrUpdateMetaByObjectIdAndKey(ctx, tagID, key,
		func(meta *entity.Meta, exist bool) (*entity.Meta, error) {
			if !exist {
				meta = &entity.Meta{ObjectID: tagID, Key: key}
			}

			existing := make(map[string]bool)
			if exist && meta.Value != "" {
				var ids []string
				if err := json.Unmarshal([]byte(meta.Value), &ids); err == nil {
					for _, id := range ids {
						if id = strings.TrimSpace(id); id != "" {
							existing[id] = true
						}
					}
				}
			}

			for _, id := range newUserIDs {
				existing[id] = true
			}

			merged := make([]string, 0, len(existing))
			for id := range existing {
				merged = append(merged, id)
			}
			sort.Strings(merged)

			data, err := json.Marshal(merged)
			if err != nil {
				return nil, err
			}
			meta.Value = string(data)
			return meta, nil
		})
}

// removeUsersFromSectionKey removes user IDs from a meta key.
func (qs *QuestionService) removeUsersFromSectionKey(ctx context.Context, tagID, roleKey string, removeIDs []string) error {
	key := entity.ForumSectionMembers
	if roleKey == "moderator" {
		key = entity.ForumSectionModerators
	}

	removeSet := make(map[string]bool)
	for _, id := range removeIDs {
		removeSet[id] = true
	}

	return qs.metaService.AddOrUpdateMetaByObjectIdAndKey(ctx, tagID, key,
		func(meta *entity.Meta, exist bool) (*entity.Meta, error) {
			if !exist {
				return meta, nil
			}
			var ids []string
			if err := json.Unmarshal([]byte(meta.Value), &ids); err != nil {
				return meta, nil
			}

			remaining := make([]string, 0)
			for _, id := range ids {
				if !removeSet[id] {
					remaining = append(remaining, id)
				}
			}

			data, err := json.Marshal(remaining)
			if err != nil {
				return nil, err
			}
			meta.Value = string(data)
			return meta, nil
		})
}

// getSectionUsernamesByKey gets usernames for a section role key.
func (qs *QuestionService) getSectionUsernamesByKey(ctx context.Context, tagID, key string) ([]string, error) {
	userSet, err := qs.getSectionUserSetByKey(ctx, tagID, key)
	if err != nil {
		return nil, err
	}

	usernames := make([]string, 0, len(userSet))
	for userID := range userSet {
		userInfo, exist, err := qs.userCommon.GetUserBasicInfoByID(ctx, userID)
		if err != nil || !exist {
			continue
		}
		usernames = append(usernames, userInfo.Username)
	}
	sort.Strings(usernames)
	return usernames, nil
}
