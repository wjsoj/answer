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

package schema

// ForumSectionResp response for a single forum section
type ForumSectionResp struct {
	TagID       string `json:"tag_id"`
	SlugName    string `json:"slug_name"`
	DisplayName string `json:"display_name"`
	Visibility  string `json:"visibility"`
	CanManage   bool   `json:"can_manage"`
}

// ForumSectionPageResp paginated list of forum sections
type ForumSectionPageResp struct {
	List  []*ForumSectionResp `json:"list"`
	Count int64               `json:"count"`
}

// ForumSectionVisibilityReq request to update section visibility
type ForumSectionVisibilityReq struct {
	Section    string `validate:"required,min=1,max=100" json:"section"`
	Visibility string `validate:"required,oneof=public private" json:"visibility"`
	UserID     string `json:"-"`
	IsAdmin    bool   `json:"-"`
}

// ForumSectionInviteReq request to invite users to a section
type ForumSectionInviteReq struct {
	Section   string   `validate:"required,min=1,max=100" json:"section"`
	Users     []string `validate:"required,min=1,max=100,dive,min=1,max=100" json:"users"`
	Role      string   `validate:"required,oneof=member moderator" json:"role"`
	InviterID string   `json:"-"`
	IsAdmin   bool     `json:"-"`
}

// ForumSectionInviteResp response for invite operation
type ForumSectionInviteResp struct {
	SkippedUsers []string `json:"skipped_users,omitempty"`
}

// ForumSectionPermissionQueryReq request to query section permissions
type ForumSectionPermissionQueryReq struct {
	Section string `validate:"required,min=1,max=100" form:"section"`
	UserID  string `json:"-"`
	IsAdmin bool   `json:"-"`
}

// ForumSectionPermissionResp response for section permissions
type ForumSectionPermissionResp struct {
	Members    []string `json:"members"`
	Moderators []string `json:"moderators"`
}

// ForumSectionRemoveReq request to remove users from a section role
type ForumSectionRemoveReq struct {
	Section   string   `validate:"required,min=1,max=100" json:"section"`
	Users     []string `validate:"required,min=1,max=100,dive,min=1,max=100" json:"users"`
	Role      string   `validate:"required,oneof=member moderator" json:"role"`
	RemoverID string   `json:"-"`
	IsAdmin   bool     `json:"-"`
}
