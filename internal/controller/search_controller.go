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
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/action"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// SearchController tag controller
type SearchController struct {
	searchService   *content.SearchService
	actionService   *action.CaptchaService
	questionService *content.QuestionService
}

// NewSearchController new controller
func NewSearchController(
	searchService *content.SearchService,
	actionService *action.CaptchaService,
	questionService *content.QuestionService,
) *SearchController {
	return &SearchController{
		searchService:   searchService,
		actionService:   actionService,
		questionService: questionService,
	}
}

// Search godoc
// @Summary search object
// @Description search object
// @Tags Search
// @Produce json
// @Security ApiKeyAuth
// @Param q query string true "query string"
// @Param order query string true "order" Enums(newest,active,score,relevance)
// @Success 200 {object} handler.RespBody{data=schema.SearchResp}
// @Router /answer/api/v1/search [get]
func (sc *SearchController) Search(ctx *gin.Context) {
	dto := schema.SearchDTO{}

	if handler.BindAndCheck(ctx, &dto) {
		return
	}
	dto.UserID = middleware.GetLoginUserIDFromContext(ctx)
	unit := ctx.ClientIP()
	if dto.UserID != "" {
		unit = dto.UserID
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := sc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionSearch, unit, dto.CaptchaID, dto.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	if !isAdmin {
		sc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionSearch, unit)
	}
	resp, err := sc.searchService.Search(ctx, &dto)
	if err == nil && resp != nil {
		// Filter search results by section access
		filtered := make([]*schema.SearchResult, 0, len(resp.SearchResults))
		for _, result := range resp.SearchResults {
			accessible := true
			for _, tag := range result.Object.Tags {
				slug := tag.SlugName
				if tag.MainTagSlugName != "" {
					slug = tag.MainTagSlugName
				}
				canAccess, aErr := sc.questionService.CanAccessSectionBySlug(ctx, dto.UserID, slug)
				if aErr != nil || !canAccess {
					accessible = false
					break
				}
			}
			if accessible {
				filtered = append(filtered, result)
			}
		}
		resp.SearchResults = filtered
		resp.Total = int64(len(filtered))
	}
	handler.HandleResponse(ctx, err, resp)
}

// SearchDesc get search description
// @Summary get search description
// @Description get search description
// @Tags Search
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.SearchResp}
// @Router /answer/api/v1/search/desc [get]
func (sc *SearchController) SearchDesc(ctx *gin.Context) {
	var finder plugin.Search
	_ = plugin.CallSearch(func(search plugin.Search) error {
		finder = search
		return nil
	})
	resp := &schema.SearchDescResp{}
	if finder != nil {
		resp.Name = finder.Info().Name.Translate(ctx)
		resp.Icon = finder.Description().Icon
		resp.Link = finder.Description().Link
	}
	handler.HandleResponse(ctx, nil, resp)
}
