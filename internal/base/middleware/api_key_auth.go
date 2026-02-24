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

package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// AuthAPIKey middleware to authenticate API key
func (am *AuthUserMiddleware) AuthAPIKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// Get API Key information including user ID
		apiKey, err := am.authService.GetAPIKeyWithUser(ctx, token)
		if err != nil || apiKey == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// Check if API key has expired
		if !apiKey.ExpiresAt.IsZero() && apiKey.ExpiresAt.Before(time.Now()) {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// Verify scope permissions - only applies to non-MCP routes (MCP uses POST for everything)
		isMCPRequest := strings.HasSuffix(ctx.Request.URL.Path, "/mcp")
		if !isMCPRequest && ctx.Request.Method != "GET" && apiKey.Scope == "read-only" {
			handler.HandleResponse(ctx, errors.Forbidden(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// Update usage count and last used time asynchronously
		go am.authService.UpdateAPIKeyUsage(ctx, apiKey.ID)

		// Inject user information into gin context
		ctx.Set("mcp_user_id", apiKey.UserID)
		ctx.Set("mcp_api_key_id", apiKey.ID)
		ctx.Set("mcp_api_key_scope", apiKey.Scope)

		// Also inject into request's standard context so mcp-go handlers can read it
		reqCtx := context.WithValue(ctx.Request.Context(), "mcp_user_id", apiKey.UserID)
		reqCtx = context.WithValue(reqCtx, "mcp_api_key_id", apiKey.ID)
		reqCtx = context.WithValue(reqCtx, "mcp_api_key_scope", apiKey.Scope)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()
	}
}
