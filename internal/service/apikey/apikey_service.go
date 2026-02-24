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

package apikey

import (
	"context"
	"strings"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/token"
)

type APIKeyRepo interface {
	GetAPIKeyList(ctx context.Context) (keys []*entity.APIKey, err error)
	GetUserAPIKeyList(ctx context.Context, userID string) (keys []*entity.APIKey, err error)
	GetUserAPIKeyByID(ctx context.Context, id int, userID string) (key *entity.APIKey, exist bool, err error)
	GetAPIKey(ctx context.Context, apiKey string) (key *entity.APIKey, exist bool, err error)
	UpdateAPIKey(ctx context.Context, apiKey entity.APIKey) (err error)
	UpdateUserAPIKey(ctx context.Context, apiKey entity.APIKey, userID string) (err error)
	AddAPIKey(ctx context.Context, apiKey entity.APIKey) (err error)
	DeleteAPIKey(ctx context.Context, id int) (err error)
	DeleteUserAPIKey(ctx context.Context, id int, userID string) (err error)
	UpdateAPIKeyUsage(ctx context.Context, apiKeyID int) (err error)
}

type APIKeyService struct {
	apiKeyRepo APIKeyRepo
}

func NewAPIKeyService(
	apiKeyRepo APIKeyRepo,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

func (s *APIKeyService) GetAPIKeyList(ctx context.Context, req *schema.GetAPIKeyReq) (resp []*schema.GetAPIKeyResp, err error) {
	keys, err := s.apiKeyRepo.GetAPIKeyList(ctx)
	if err != nil {
		return nil, err
	}
	resp = make([]*schema.GetAPIKeyResp, 0)
	for _, key := range keys {
		// hide access key middle part, replace with *
		if len(key.AccessKey) < 10 {
			// If the access key is too short, do not mask it
			key.AccessKey = strings.Repeat("*", len(key.AccessKey))
		} else {
			key.AccessKey = key.AccessKey[:7] + strings.Repeat("*", 8) + key.AccessKey[len(key.AccessKey)-4:]
		}

		resp = append(resp, &schema.GetAPIKeyResp{
			ID:          key.ID,
			AccessKey:   key.AccessKey,
			Description: key.Description,
			Scope:       key.Scope,
			CreatedAt:   key.CreatedAt.Unix(),
			LastUsedAt:  key.LastUsedAt.Unix(),
		})
	}
	return resp, nil
}

func (s *APIKeyService) UpdateAPIKey(ctx context.Context, req *schema.UpdateAPIKeyReq) (err error) {
	apiKey := entity.APIKey{
		ID:          req.ID,
		Description: req.Description,
	}
	err = s.apiKeyRepo.UpdateAPIKey(ctx, apiKey)
	if err != nil {
		return err
	}
	return nil
}

func (s *APIKeyService) AddAPIKey(ctx context.Context, req *schema.AddAPIKeyReq) (resp *schema.AddAPIKeyResp, err error) {
	ak := "sk_" + strings.ReplaceAll(token.GenerateToken(), "-", "")
	apiKey := entity.APIKey{
		Description: req.Description,
		AccessKey:   ak,
		Scope:       req.Scope,
		LastUsedAt:  time.Now(),
		UserID:      req.UserID,
	}
	err = s.apiKeyRepo.AddAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	resp = &schema.AddAPIKeyResp{
		AccessKey: apiKey.AccessKey,
	}
	return resp, nil
}

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, req *schema.DeleteAPIKeyReq) (err error) {
	err = s.apiKeyRepo.DeleteAPIKey(ctx, req.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *APIKeyService) GetUserAPIKeyList(ctx context.Context, userID string) (resp []*schema.GetUserAPIKeysResp, err error) {
	keys, err := s.apiKeyRepo.GetUserAPIKeyList(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp = make([]*schema.GetUserAPIKeysResp, 0)
	for _, key := range keys {
		// hide access key middle part, replace with *
		maskedKey := key.AccessKey
		if len(key.AccessKey) < 10 {
			maskedKey = strings.Repeat("*", len(key.AccessKey))
		} else {
			maskedKey = key.AccessKey[:7] + strings.Repeat("*", 8) + key.AccessKey[len(key.AccessKey)-4:]
		}

		expiresAt := int64(0)
		if !key.ExpiresAt.IsZero() {
			expiresAt = key.ExpiresAt.Unix()
		}

		resp = append(resp, &schema.GetUserAPIKeysResp{
			ID:          key.ID,
			Name:        key.Name,
			AccessKey:   maskedKey,
			Description: key.Description,
			CreatedAt:   key.CreatedAt.Unix(),
			LastUsedAt:  key.LastUsedAt.Unix(),
			ExpiresAt:   expiresAt,
			UsageCount:  key.UsageCount,
		})
	}
	return resp, nil
}

func (s *APIKeyService) AddUserAPIKey(ctx context.Context, userID string, req *schema.AddUserAPIKeyReq) (resp *schema.AddUserAPIKeyResp, err error) {
	ak := "sk_" + strings.ReplaceAll(token.GenerateToken(), "-", "")
	apiKey := entity.APIKey{
		Name:        req.Name,
		Description: req.Description,
		AccessKey:   ak,
		Scope:       "read_write",
		LastUsedAt:  time.Now(),
		UserID:      userID,
		UsageCount:  0,
	}

	// Set expiration if specified
	if req.ExpiresIn > 0 {
		apiKey.ExpiresAt = time.Now().AddDate(0, 0, req.ExpiresIn)
	}

	err = s.apiKeyRepo.AddAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	resp = &schema.AddUserAPIKeyResp{
		AccessKey: apiKey.AccessKey,
	}
	return resp, nil
}

func (s *APIKeyService) UpdateUserAPIKey(ctx context.Context, userID string, req *schema.UpdateUserAPIKeyReq) (err error) {
	// Verify ownership
	_, exist, err := s.apiKeyRepo.GetUserAPIKeyByID(ctx, req.ID, userID)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}

	apiKey := entity.APIKey{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	}
	err = s.apiKeyRepo.UpdateUserAPIKey(ctx, apiKey, userID)
	if err != nil {
		return err
	}
	return nil
}

func (s *APIKeyService) DeleteUserAPIKey(ctx context.Context, userID string, id int) (err error) {
	err = s.apiKeyRepo.DeleteUserAPIKey(ctx, id, userID)
	if err != nil {
		return err
	}
	return nil
}
