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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/repo/limit"
	"github.com/apache/answer/pkg/encryption"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type RateLimitMiddleware struct {
	limitRepo *limit.LimitRepo
}

// NewRateLimitMiddleware new rate limit middleware
func NewRateLimitMiddleware(limitRepo *limit.LimitRepo) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limitRepo: limitRepo,
	}
}

// DuplicateRequestRejection detects and rejects duplicate requests
// It only works for the requests that post content. Such as add question, add answer, comment etc.
func (rm *RateLimitMiddleware) DuplicateRequestRejection(ctx *gin.Context, req any) (reject bool, key string) {
	userID := GetLoginUserIDFromContext(ctx)
	fullPath := ctx.FullPath()
	reqJson, _ := json.Marshal(req)
	key = encryption.MD5(fmt.Sprintf("%s:%s:%s", userID, fullPath, string(reqJson)))
	var err error
	reject, err = rm.limitRepo.CheckAndRecord(ctx, key)
	if err != nil {
		log.Errorf("check and record rate limit error: %s", err.Error())
		return false, key
	}
	if !reject {
		return false, key
	}
	log.Debugf("duplicate request: [%s] %s", fullPath, string(reqJson))
	handler.HandleResponse(ctx, errors.BadRequest(reason.DuplicateRequestError), nil)
	return true, key
}

// DuplicateRequestClear clear duplicate request record
func (rm *RateLimitMiddleware) DuplicateRequestClear(ctx *gin.Context, key string) {
	err := rm.limitRepo.ClearRecord(ctx, key)
	if err != nil {
		log.Errorf("clear rate limit error: %s", err.Error())
	}
}

// MCP Rate Limiter
type mcpRateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

var mcpLimiter = &mcpRateLimiter{
	requests: make(map[string][]time.Time),
	limit:    100, // 100 requests per minute
	window:   time.Minute,
}

func init() {
	// Cleanup old entries every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mcpLimiter.cleanup()
		}
	}()
}

func (rl *mcpRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

func (rl *mcpRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[key]

	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// RateLimitMCP rate limits MCP requests per API key
func RateLimitMCP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKeyID, exists := ctx.Get("mcp_api_key_id")
		key := ""
		if exists {
			key = fmt.Sprintf("mcp_%v", apiKeyID)
		} else {
			key = fmt.Sprintf("mcp_ip_%s", ctx.ClientIP())
		}

		if !mcpLimiter.allow(key) {
			handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
