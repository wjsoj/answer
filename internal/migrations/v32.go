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

func addAPIKeyNameField(ctx context.Context, x *xorm.Engine) error {
	// Add Name column to api_key table
	type APIKey struct {
		Name string `xorm:"not null default '' VARCHAR(100) name"`
	}

	if err := x.Context(ctx).Sync(new(APIKey)); err != nil {
		return fmt.Errorf("sync api_key name field failed: %w", err)
	}

	// Add index on user_id for better query performance
	if err := x.Context(ctx).Sync(new(entity.APIKey)); err != nil {
		return fmt.Errorf("sync api_key table failed: %w", err)
	}

	return nil
}
