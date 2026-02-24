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
	"strconv"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/apikey"
	"github.com/gin-gonic/gin"
)

type APIKeyController struct {
	apiKeyService *apikey.APIKeyService
}

// NewAPIKeyController creates a new API key controller
func NewAPIKeyController(
	apiKeyService *apikey.APIKeyService,
) *APIKeyController {
	return &APIKeyController{
		apiKeyService: apiKeyService,
	}
}

// GetUserAPIKeys godoc
// @Summary Get user API keys
// @Description Get all API keys for the current user
// @Tags API Key
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=[]schema.GetUserAPIKeysResp}
// @Router /answer/api/v1/user/api-keys [get]
func (ac *APIKeyController) GetUserAPIKeys(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)

	resp, err := ac.apiKeyService.GetUserAPIKeyList(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// CreateUserAPIKey godoc
// @Summary Create user API key
// @Description Create a new API key for the current user
// @Tags API Key
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AddUserAPIKeyReq true "API key info"
// @Success 200 {object} handler.RespBody{data=schema.AddUserAPIKeyResp}
// @Router /answer/api/v1/user/api-keys [post]
func (ac *APIKeyController) CreateUserAPIKey(ctx *gin.Context) {
	req := &schema.AddUserAPIKeyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := ac.apiKeyService.AddUserAPIKey(ctx, userID, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateUserAPIKey godoc
// @Summary Update user API key
// @Description Update an API key for the current user
// @Tags API Key
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "API key ID"
// @Param data body schema.UpdateUserAPIKeyReq true "API key info"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/api-keys/{id} [put]
func (ac *APIKeyController) UpdateUserAPIKey(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	req := &schema.UpdateUserAPIKeyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = id

	userID := middleware.GetLoginUserIDFromContext(ctx)
	err = ac.apiKeyService.UpdateUserAPIKey(ctx, userID, req)
	handler.HandleResponse(ctx, err, nil)
}

// DeleteUserAPIKey godoc
// @Summary Delete user API key
// @Description Delete an API key for the current user
// @Tags API Key
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "API key ID"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/api-keys/{id} [delete]
func (ac *APIKeyController) DeleteUserAPIKey(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	userID := middleware.GetLoginUserIDFromContext(ctx)
	err = ac.apiKeyService.DeleteUserAPIKey(ctx, userID, id)
	handler.HandleResponse(ctx, err, nil)
}
