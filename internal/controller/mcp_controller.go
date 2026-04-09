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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/answer/pkg/converter"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/collection"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/feature_toggle"
	"github.com/apache/answer/internal/service/follow"
	"github.com/apache/answer/internal/service/notification"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/report"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommonser "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/segmentfault/pacman/log"
)

type MCPController struct {
	searchService        *content.SearchService
	siteInfoService      siteinfo_common.SiteInfoCommonService
	tagCommonService     *tagcommonser.TagCommonService
	questioncommon       *questioncommon.QuestionCommon
	commentRepo          comment.CommentRepo
	userCommon           *usercommon.UserCommon
	answerRepo           answercommon.AnswerRepo
	featureToggleSvc     *feature_toggle.FeatureToggleService
	questionService      *content.QuestionService
	answerService        *content.AnswerService
	commentService       *comment.CommentService
	voteService          *content.VoteService
	notificationService  *notification.NotificationService
	collectionService    *collection.CollectionService
	followService        *follow.FollowService
	reportService        *report.ReportService
}

// NewMCPController new site info controller.
func NewMCPController(
	searchService *content.SearchService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	tagCommonService *tagcommonser.TagCommonService,
	questioncommon *questioncommon.QuestionCommon,
	commentRepo comment.CommentRepo,
	userCommon *usercommon.UserCommon,
	answerRepo answercommon.AnswerRepo,
	featureToggleSvc *feature_toggle.FeatureToggleService,
	questionService *content.QuestionService,
	answerService *content.AnswerService,
	commentService *comment.CommentService,
	voteService *content.VoteService,
	notificationService *notification.NotificationService,
	collectionService *collection.CollectionService,
	followService *follow.FollowService,
	reportService *report.ReportService,
) *MCPController {
	return &MCPController{
		searchService:       searchService,
		siteInfoService:     siteInfoService,
		tagCommonService:    tagCommonService,
		questioncommon:      questioncommon,
		commentRepo:         commentRepo,
		userCommon:          userCommon,
		answerRepo:          answerRepo,
		featureToggleSvc:    featureToggleSvc,
		questionService:     questionService,
		answerService:       answerService,
		commentService:      commentService,
		voteService:         voteService,
		notificationService: notificationService,
		collectionService:   collectionService,
		followService:       followService,
		reportService:       reportService,
	}
}

func (c *MCPController) ensureMCPEnabled(ctx context.Context) error {
	if c.featureToggleSvc == nil {
		return nil
	}
	return c.featureToggleSvc.EnsureEnabled(ctx, feature_toggle.FeatureMCP)
}

// getUserIDFromContext extracts user ID from context (set by API key auth middleware)
func (c *MCPController) getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("mcp_user_id").(string); ok {
		return userID
	}
	return ""
}

// isReadOnlyScope checks if the API key has read-only scope
func (c *MCPController) isReadOnlyScope(ctx context.Context) bool {
	scope, _ := ctx.Value("mcp_api_key_scope").(string)
	return scope == "read-only"
}

// MCPForumSectionsHandler returns accessible forum sections
func (c *MCPController) MCPForumSectionsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		userID := c.getUserIDFromContext(ctx)
		resp, err := c.questionService.GetAccessibleForumSections(ctx, userID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get forum sections: %v", err)), nil
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// MCPUpdateSectionVisibilityHandler updates section visibility
func (c *MCPController) MCPUpdateSectionVisibilityHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}
		args := request.GetArguments()
		section, _ := args["section"].(string)
		visibility, _ := args["visibility"].(string)
		if section == "" || visibility == "" {
			return mcp.NewToolResultError("section and visibility are required"), nil
		}
		req := &schema.ForumSectionVisibilityReq{
			Section:    section,
			Visibility: visibility,
			UserID:     userID,
		}
		err := c.questionService.UpdateForumSectionVisibility(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update visibility: %v", err)), nil
		}
		return mcp.NewToolResultText("Section visibility updated successfully"), nil
	}
}

// MCPInviteSectionUsersHandler invites users to a section
func (c *MCPController) MCPInviteSectionUsersHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}
		args := request.GetArguments()
		section, _ := args["section"].(string)
		roleStr, _ := args["role"].(string)
		usersRaw, _ := args["users"].([]interface{})
		users := make([]string, 0, len(usersRaw))
		for _, u := range usersRaw {
			if s, ok := u.(string); ok {
				users = append(users, s)
			}
		}
		if section == "" || roleStr == "" || len(users) == 0 {
			return mcp.NewToolResultError("section, role, and users are required"), nil
		}
		req := &schema.ForumSectionInviteReq{
			Section:   section,
			Users:     users,
			Role:      roleStr,
			InviterID: userID,
			IsAdmin:   c.questionService.IsUserAdminOrModerator(ctx, userID),
		}
		resp, err := c.questionService.InviteForumSectionUsers(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to invite users: %v", err)), nil
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// MCPRemoveSectionUsersHandler removes users from a section
func (c *MCPController) MCPRemoveSectionUsersHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}
		args := request.GetArguments()
		section, _ := args["section"].(string)
		roleStr, _ := args["role"].(string)
		usersRaw, _ := args["users"].([]interface{})
		users := make([]string, 0, len(usersRaw))
		for _, u := range usersRaw {
			if s, ok := u.(string); ok {
				users = append(users, s)
			}
		}
		if section == "" || roleStr == "" || len(users) == 0 {
			return mcp.NewToolResultError("section, role, and users are required"), nil
		}
		req := &schema.ForumSectionRemoveReq{
			Section:   section,
			Users:     users,
			Role:      roleStr,
			RemoverID: userID,
			IsAdmin:   c.questionService.IsUserAdminOrModerator(ctx, userID),
		}
		err := c.questionService.RemoveForumSectionUsers(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove users: %v", err)), nil
		}
		return mcp.NewToolResultText("Users removed successfully"), nil
	}
}

// MCPGetSectionPermissionsHandler returns section member/moderator lists
func (c *MCPController) MCPGetSectionPermissionsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		args := request.GetArguments()
		section, _ := args["section"].(string)
		if section == "" {
			return mcp.NewToolResultError("section is required"), nil
		}
		req := &schema.ForumSectionPermissionQueryReq{
			Section: section,
			UserID:  userID,
		}
		resp, err := c.questionService.GetForumSectionPermissions(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get permissions: %v", err)), nil
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPQuestionsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchCond(request)
		userID := c.getUserIDFromContext(ctx)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		// If section filter specified, check access first and add as tag filter
		if cond.Section != "" {
			canAccess, aErr := c.questionService.CanAccessSectionBySlug(ctx, userID, cond.Section)
			if aErr != nil || !canAccess {
				return mcp.NewToolResultText("[]"), nil
			}
			// Add section as a tag filter for the search
			cond.Tags = append(cond.Tags, cond.Section)
		}

		searchResp, err := c.searchService.Search(ctx, &schema.SearchDTO{
			Query:  cond.ToQueryString() + " is:question",
			Page:   1,
			Size:   5,
			Order:  "newest",
			UserID: userID,
		})
		if err != nil {
			return nil, err
		}
		resp := make([]*schema.MCPSearchQuestionInfoResp, 0)
		for _, question := range searchResp.SearchResults {
			// Check section access for each question's tags
			accessible := true
			for _, tag := range question.Object.Tags {
				if tag.MainTagSlugName != "" {
					canAccess, err := c.questionService.CanAccessSectionBySlug(ctx, userID, tag.MainTagSlugName)
					if err != nil || !canAccess {
						accessible = false
						break
					}
				} else {
					canAccess, err := c.questionService.CanAccessSectionBySlug(ctx, userID, tag.SlugName)
					if err != nil || !canAccess {
						accessible = false
						break
					}
				}
			}
			if !accessible {
				continue
			}
			t := &schema.MCPSearchQuestionInfoResp{
				QuestionID: question.Object.QuestionID,
				Title:      question.Object.Title,
				Content:    question.Object.Excerpt,
				Link:       fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, question.Object.QuestionID),
			}
			resp = append(resp, t)
		}

		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPQuestionDetailHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchQuestionDetail(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		question, err := c.questioncommon.Info(ctx, cond.QuestionID, userID)
		if err != nil {
			log.Errorf("get question failed: %v", err)
			return mcp.NewToolResultText("No question found."), nil
		}

		// Check section access via question tags
		for _, tag := range question.Tags {
			slug := tag.SlugName
			if tag.MainTagSlugName != "" {
				slug = tag.MainTagSlugName
			}
			canAccess, aErr := c.questionService.CanAccessSectionBySlug(ctx, userID, slug)
			if aErr != nil || !canAccess {
				return mcp.NewToolResultText("No question found."), nil
			}
		}

		resp := &schema.MCPSearchQuestionInfoResp{
			QuestionID: question.ID,
			Title:      question.Title,
			Content:    question.Content,
			Link:       fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, question.ID),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

func (c *MCPController) MCPAnswersHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchAnswerCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		// Check section access on the question before returning answers
		userID := c.getUserIDFromContext(ctx)
		questionInfo, qErr := c.questioncommon.Info(ctx, cond.QuestionID, userID)
		if qErr != nil {
			return mcp.NewToolResultText("No answers found."), nil
		}
		for _, tag := range questionInfo.Tags {
			slug := tag.SlugName
			if tag.MainTagSlugName != "" {
				slug = tag.MainTagSlugName
			}
			canAccess, aErr := c.questionService.CanAccessSectionBySlug(ctx, userID, slug)
			if aErr != nil || !canAccess {
				return mcp.NewToolResultText("No answers found."), nil
			}
		}

		answerList, err := c.answerRepo.GetAnswerList(ctx, &entity.Answer{QuestionID: cond.QuestionID})
		if err != nil {
			log.Errorf("get answers failed: %v", err)
			return nil, err
		}
		resp := make([]*schema.MCPSearchAnswerInfoResp, 0)
		for _, answer := range answerList {
			t := &schema.MCPSearchAnswerInfoResp{
				QuestionID:    answer.QuestionID,
				AnswerID:      answer.ID,
				AnswerContent: answer.OriginalText,
				Link:          fmt.Sprintf("%s/questions/%s/answers/%s", siteGeneral.SiteUrl, answer.QuestionID, answer.ID),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPCommentsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchCommentCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		dto := &comment.CommentQuery{
			PageCond:  pager.PageCond{Page: 1, PageSize: 5},
			QueryCond: "newest",
			ObjectID:  cond.ObjectID,
		}
		commentList, total, err := c.commentRepo.GetCommentPage(ctx, dto)
		if err != nil {
			return nil, err
		}
		if total == 0 {
			return mcp.NewToolResultText("No comments found."), nil
		}

		resp := make([]*schema.MCPSearchCommentInfoResp, 0)
		for _, comment := range commentList {
			t := &schema.MCPSearchCommentInfoResp{
				CommentID: comment.ID,
				Content:   comment.OriginalText,
				ObjectID:  comment.ObjectID,
				Link:      fmt.Sprintf("%s/comments/%s", siteGeneral.SiteUrl, comment.ID),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPTagsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchTagCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		tags, total, err := c.tagCommonService.GetTagPage(ctx, 1, 10, &entity.Tag{DisplayName: cond.TagName}, "newest")
		if err != nil {
			log.Errorf("get tags failed: %v", err)
			return nil, err
		}

		if total == 0 {
			res := strings.Builder{}
			res.WriteString("No tags found.\n")
			return mcp.NewToolResultText(res.String()), nil
		}

		resp := make([]*schema.MCPSearchTagResp, 0)
		for _, tag := range tags {
			t := &schema.MCPSearchTagResp{
				TagName:     tag.SlugName,
				DisplayName: tag.DisplayName,
				Description: tag.OriginalText,
				Link:        fmt.Sprintf("%s/tags/%s", siteGeneral.SiteUrl, tag.SlugName),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPTagDetailsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchTagCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		tag, exist, err := c.tagCommonService.GetTagBySlugName(ctx, cond.TagName)
		if err != nil {
			log.Errorf("get tag failed: %v", err)
			return nil, err
		}
		if !exist {
			return mcp.NewToolResultText("Tag not found."), nil
		}

		resp := &schema.MCPSearchTagResp{
			TagName:     tag.SlugName,
			DisplayName: tag.DisplayName,
			Description: tag.OriginalText,
			Link:        fmt.Sprintf("%s/tags/%s", siteGeneral.SiteUrl, tag.SlugName),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

func (c *MCPController) MCPUserDetailsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchUserCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		user, exist, err := c.userCommon.GetUserBasicInfoByUserName(ctx, cond.Username)
		if err != nil {
			log.Errorf("get user failed: %v", err)
			return nil, err
		}
		if !exist {
			return mcp.NewToolResultText("User not found."), nil
		}

		resp := &schema.MCPSearchUserInfoResp{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			Link:        fmt.Sprintf("%s/users/%s", siteGeneral.SiteUrl, user.Username),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

// Write operation handlers

func (c *MCPController) MCPCreateQuestionHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		tagsInterface, _ := args["tags"].([]interface{})

		// Convert tags
		var tags []*schema.TagItem
		for _, t := range tagsInterface {
			if tagStr, ok := t.(string); ok {
				tags = append(tags, &schema.TagItem{SlugName: tagStr})
			}
		}

		section, _ := args["section"].(string)

		req := &schema.QuestionAdd{
			Title:   title,
			Content: content,
			HTML:    converter.Markdown2HTML(content),
			Tags:    tags,
			Section: section,
			UserID:  userID,
		}

		resp, err := c.questionService.AddQuestion(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create question: %v", err)), nil
		}

		result, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func (c *MCPController) MCPUpdateQuestionHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		questionID, _ := args["question_id"].(string)
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		editSummary, _ := args["edit_summary"].(string)
		tagsInterface, _ := args["tags"].([]interface{})

		var tags []*schema.TagItem
		for _, t := range tagsInterface {
			if tagStr, ok := t.(string); ok {
				tags = append(tags, &schema.TagItem{SlugName: tagStr})
			}
		}

		section, _ := args["section"].(string)

		req := &schema.QuestionUpdate{
			ID:          questionID,
			Title:       title,
			Content:     content,
			HTML:        converter.Markdown2HTML(content),
			Tags:        tags,
			Section:     section,
			EditSummary: editSummary,
			UserID:      userID,
		}

		resp, err := c.questionService.UpdateQuestion(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update question: %v", err)), nil
		}

		result, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func (c *MCPController) MCPDeleteQuestionHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		questionID, _ := args["question_id"].(string)

		req := &schema.RemoveQuestionReq{
			ID:     questionID,
			UserID: userID,
		}

		err := c.questionService.RemoveQuestion(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete question: %v", err)), nil
		}

		return mcp.NewToolResultText("Question deleted successfully"), nil
	}
}

func (c *MCPController) MCPCreateAnswerHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		questionID, _ := args["question_id"].(string)
		content, _ := args["content"].(string)

		// Check section access on the question
		canAccess, _ := c.questionService.CanAccessQuestionByID(ctx, userID, questionID)
		if !canAccess {
			return mcp.NewToolResultError("You do not have access to this question's section"), nil
		}

		req := &schema.AnswerAddReq{
			QuestionID: questionID,
			Content:    content,
			HTML:       converter.Markdown2HTML(content),
			UserID:     userID,
		}

		answerID, err := c.answerService.Insert(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create answer: %v", err)), nil
		}

		result := map[string]string{"answer_id": answerID}
		resultJSON, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(resultJSON)), nil
	}
}

func (c *MCPController) MCPUpdateAnswerHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		answerID, _ := args["answer_id"].(string)
		content, _ := args["content"].(string)
		editSummary, _ := args["edit_summary"].(string)

		req := &schema.AnswerUpdateReq{
			ID:          answerID,
			Content:     content,
			HTML:        converter.Markdown2HTML(content),
			EditSummary: editSummary,
			UserID:      userID,
		}

		_, err := c.answerService.Update(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update answer: %v", err)), nil
		}

		return mcp.NewToolResultText("Answer updated successfully"), nil
	}
}

func (c *MCPController) MCPDeleteAnswerHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		answerID, _ := args["answer_id"].(string)

		req := &schema.RemoveAnswerReq{
			ID:     answerID,
			UserID: userID,
		}

		err := c.answerService.RemoveAnswer(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete answer: %v", err)), nil
		}

		return mcp.NewToolResultText("Answer deleted successfully"), nil
	}
}

func (c *MCPController) MCPCreateCommentHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		objectID, _ := args["object_id"].(string)
		content, _ := args["content"].(string)

		req := &schema.AddCommentReq{
			ObjectID:            objectID,
			OriginalText:        content,
			ParsedText:          converter.Markdown2HTML(content),
			UserID:              userID,
			MentionUsernameList: []string{},
		}

		resp, err := c.commentService.AddComment(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create comment: %v", err)), nil
		}

		result, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func (c *MCPController) MCPUpdateCommentHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		commentID, _ := args["comment_id"].(string)
		content, _ := args["content"].(string)

		req := &schema.UpdateCommentReq{
			CommentID:    commentID,
			OriginalText: content,
			ParsedText:   converter.Markdown2HTML(content),
			UserID:       userID,
		}

		_, err := c.commentService.UpdateComment(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update comment: %v", err)), nil
		}

		return mcp.NewToolResultText("Comment updated successfully"), nil
	}
}

func (c *MCPController) MCPDeleteCommentHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		commentID, _ := args["comment_id"].(string)

		req := &schema.RemoveCommentReq{
			CommentID: commentID,
			UserID:    userID,
		}

		err := c.commentService.RemoveComment(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete comment: %v", err)), nil
		}

		return mcp.NewToolResultText("Comment deleted successfully"), nil
	}
}

func (c *MCPController) MCPVoteHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		objectID, _ := args["object_id"].(string)
		voteType, _ := args["vote_type"].(string)
		isCancel, _ := args["is_cancel"].(bool)

		req := &schema.VoteReq{
			ObjectID: objectID,
			UserID:   userID,
			IsCancel: isCancel,
		}

		var resp *schema.VoteResp
		var err error

		switch voteType {
		case "up":
			resp, err = c.voteService.VoteUp(ctx, req)
		case "down":
			resp, err = c.voteService.VoteDown(ctx, req)
		default:
			return mcp.NewToolResultError("Invalid vote_type. Must be 'up' or 'down'"), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to vote: %v", err)), nil
		}

		result, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// Notification handlers

func (c *MCPController) MCPNotificationsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := request.GetArguments()
		typeStr, _ := args[schema.MCPSearchCondNotificationType].(string)
		inboxType, _ := args[schema.MCPSearchCondInboxType].(string)

		if typeStr == "" {
			typeStr = "inbox"
		}

		page := 1
		pageSize := 10
		if p, ok := args[schema.MCPSearchCondPage].(float64); ok && int(p) > 0 {
			page = int(p)
		}
		if ps, ok := args[schema.MCPSearchCondPageSize].(float64); ok && int(ps) > 0 {
			pageSize = int(ps)
		}

		searchCond := &schema.NotificationSearch{
			Page:         page,
			PageSize:     pageSize,
			TypeStr:      typeStr,
			InboxTypeStr: inboxType,
			UserID:       userID,
		}

		pageModel, err := c.notificationService.GetNotificationPage(ctx, searchCond)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get notifications: %v", err)), nil
		}

		data, _ := json.Marshal(pageModel)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPUnreadCountHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		req := &schema.GetRedDot{
			UserID: userID,
		}

		redDot, err := c.notificationService.GetRedDot(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get unread count: %v", err)), nil
		}

		resp := map[string]int64{
			"inbox":       redDot.Inbox,
			"achievement": redDot.Achievement,
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPMarkNotificationReadHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		notificationID, _ := args[schema.MCPSearchCondNotificationID].(string)
		if notificationID == "" {
			return mcp.NewToolResultError("notification_id is required"), nil
		}

		err := c.notificationService.ClearIDUnRead(ctx, userID, notificationID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to mark notification as read: %v", err)), nil
		}

		return mcp.NewToolResultText("Notification marked as read"), nil
	}
}

func (c *MCPController) MCPMarkAllNotificationsReadHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		typeStr, _ := args[schema.MCPSearchCondNotificationType].(string)
		if typeStr == "" {
			return mcp.NewToolResultError("type is required (inbox or achievement)"), nil
		}

		err := c.notificationService.ClearUnRead(ctx, userID, typeStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to mark all notifications as read: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("All %s notifications marked as read", typeStr)), nil
	}
}

// Collection switch handler - add or remove question from personal collection

func (c *MCPController) MCPCollectionSwitchHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		objectID, _ := args[schema.MCPSearchCondObjectID].(string)
		if objectID == "" {
			return mcp.NewToolResultError("object_id is required"), nil
		}

		bookmark := true
		if b, ok := args[schema.MCPSearchCondBookmark].(bool); ok {
			bookmark = b
		}

		groupID := ""
		if g, ok := args[schema.MCPSearchCondGroupID].(string); ok {
			groupID = g
		}

		// Use default group if not provided
		if groupID == "" {
			groupID = "0"
		}

		req := &schema.CollectionSwitchReq{
			ObjectID: objectID,
			Bookmark: bookmark,
			GroupID:  groupID,
			UserID:   userID,
		}

		resp, err := c.collectionService.CollectionSwitch(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to switch collection: %v", err)), nil
		}

		result := &schema.MCPSwitchCollectionResp{
			ObjectCollectionCount: resp.ObjectCollectionCount,
			Success:               true,
			Message:              "Collection updated successfully",
		}
		if !bookmark {
			result.Message = "Removed from collection successfully"
		}

		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// Get personal collections handler

func (c *MCPController) MCPPersonalCollectionsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := request.GetArguments()
		page := 1
		pageSize := 10

		if p, ok := args[schema.MCPSearchCondPage].(float64); ok && int(p) > 0 {
			page = int(p)
		}
		if ps, ok := args[schema.MCPSearchCondPageSize].(float64); ok && int(ps) > 0 {
			pageSize = int(ps)
		}

		req := &schema.PersonalCollectionPageReq{
			Page:     page,
			PageSize: pageSize,
			UserID:   userID,
		}

		pageModel, err := c.questionService.PersonalCollectionPage(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get personal collections: %v", err)), nil
		}

		data, _ := json.Marshal(pageModel)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// Follow/Unfollow handler

func (c *MCPController) MCPFollowHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		objectID, _ := args[schema.MCPSearchCondObjectID].(string)
		if objectID == "" {
			return mcp.NewToolResultError("object_id is required"), nil
		}

		isCancel := false
		if ic, ok := args[schema.MCPSearchCondIsCancel].(bool); ok {
			isCancel = ic
		}

		dto := &schema.FollowDTO{
			ObjectID: objectID,
			IsCancel: isCancel,
			UserID:   userID,
		}

		resp, err := c.followService.Follow(ctx, dto)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to follow/unfollow: %v", err)), nil
		}

		result := &schema.MCPFollowResp{
			IsFollowed: resp.IsFollowed,
			Follows:    int64(resp.Follows),
		}
		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// Accept answer handler
// Deprecated: This handler is no longer used as the acceptance feature has been removed.

// func (c *MCPController) MCPAcceptAnswerHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 		if err := c.ensureMCPEnabled(ctx); err != nil {
// 			return nil, err
// 		}
//
// 		userID := c.getUserIDFromContext(ctx)
// 		if userID == "" {
// 			return mcp.NewToolResultError("Authentication required"), nil
// 		}
// 		if c.isReadOnlyScope(ctx) {
// 			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
// 		}
//
// 		args := request.GetArguments()
// 		questionID, _ := args[schema.MCPSearchCondQuestionID].(string)
// 		answerID, _ := args[schema.MCPSearchCondAnswerID].(string)
//
// 		if questionID == "" || answerID == "" {
// 			return mcp.NewToolResultError("question_id and answer_id are required"), nil
// 		}
//
// 		req := &schema.AcceptAnswerReq{
// 			QuestionID: questionID,
// 			AnswerID:   answerID,
// 			UserID:     userID,
// 		}
//
// 		err := c.answerService.AcceptAnswer(ctx, req)
// 		if err != nil {
// 			return mcp.NewToolResultError(fmt.Sprintf("Failed to accept answer: %v", err)), nil
// 		}
//
// 		return mcp.NewToolResultText("Answer accepted successfully"), nil
// 	}
// }

// Get question detail handler

func (c *MCPController) MCPGetQuestionDetailHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		args := request.GetArguments()
		questionID, _ := args[schema.MCPSearchCondQuestionID].(string)
		if questionID == "" {
			return mcp.NewToolResultError("question_id is required"), nil
		}

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		// Get user ID from context if authenticated
		userID := c.getUserIDFromContext(ctx)

		question, err := c.questioncommon.Info(ctx, questionID, userID)
		if err != nil {
			log.Errorf("get question failed: %v", err)
			return mcp.NewToolResultText("No question found."), nil
		}

		// Check section access via question tags
		for _, tag := range question.Tags {
			slug := tag.SlugName
			if tag.MainTagSlugName != "" {
				slug = tag.MainTagSlugName
			}
			canAccess, aErr := c.questionService.CanAccessSectionBySlug(ctx, userID, slug)
			if aErr != nil || !canAccess {
				return mcp.NewToolResultText("No question found."), nil
			}
		}

		resp := map[string]interface{}{
			"question_id":       question.ID,
			"title":             question.Title,
			"content":           question.Content,
			"html":              question.HTML,
			"link":              fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, question.ID),
			"vote_count":       question.VoteCount,
			"answer_count":     question.AnswerCount,
			"view_count":       question.ViewCount,
			"status":           question.Status,
			"accepted_answer_id": question.AcceptedAnswerID,
			"create_time":       question.CreateTime,
			"update_time":       question.PostUpdateTime,
			"tags":             question.Tags,
			"user_info":        question.UserInfo,
			"collected":        question.Collected,
			"is_followed":      question.IsFollowed,
			"vote_status":      question.VoteStatus,
		}

		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// Get answer detail handler

func (c *MCPController) MCPGetAnswerDetailHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		args := request.GetArguments()
		answerID, _ := args[schema.MCPSearchCondAnswerID].(string)
		if answerID == "" {
			return mcp.NewToolResultError("answer_id is required"), nil
		}

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		answer, err := c.answerService.GetDetail(ctx, answerID)
		if err != nil {
			log.Errorf("get answer failed: %v", err)
			return mcp.NewToolResultText("No answer found."), nil
		}
		if answer == nil {
			return mcp.NewToolResultText("Answer not found."), nil
		}

		// Check section access on the parent question
		userID := c.getUserIDFromContext(ctx)
		canAccess, _ := c.questionService.CanAccessQuestionByID(ctx, userID, answer.QuestionID)
		if !canAccess {
			return mcp.NewToolResultText("Answer not found."), nil
		}

		// Get question title via questioncommon.Info
		questionInfo, err := c.questioncommon.Info(ctx, answer.QuestionID, userID)
		if err != nil {
			log.Errorf("get question info failed: %v", err)
		}

		resp := map[string]interface{}{
			"answer_id":      answer.ID,
			"question_id":    answer.QuestionID,
			"content":        answer.Content,
			"parsed_content": answer.HTML,
			"link":           fmt.Sprintf("%s/questions/%s/answers/%s", siteGeneral.SiteUrl, answer.QuestionID, answer.ID),
			"vote_count":    answer.VoteCount,
			"status":        answer.Status,
			"is_accepted":   answer.Accepted,
			"created_at":    answer.CreateTime,
			"updated_at":    answer.UpdateTime,
			"user_info":     answer.UserInfo,
		}

		if questionInfo != nil {
			resp["question_title"] = questionInfo.Title
		}

		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// Report content handler

func (c *MCPController) MCPReportContentHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}

		userID := c.getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}
		if c.isReadOnlyScope(ctx) {
			return mcp.NewToolResultError("This operation requires a non read-only API key"), nil
		}

		args := request.GetArguments()
		objectID, _ := args[schema.MCPSearchCondObjectID].(string)
		reportTypeVal, _ := args[schema.MCPSearchCondReportType].(float64)
		content, _ := args[schema.MCPSearchCondContent].(string)

		if objectID == "" || reportTypeVal == 0 {
			return mcp.NewToolResultError("object_id and report_type are required"), nil
		}

		req := &schema.AddReportReq{
			ObjectID:   objectID,
			ReportType: int(reportTypeVal),
			Content:    content,
			UserID:     userID,
		}

		err := c.reportService.AddReport(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to report content: %v", err)), nil
		}

		return mcp.NewToolResultText("Content reported successfully"), nil
	}
}
