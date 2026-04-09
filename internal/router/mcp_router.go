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

package router

import (
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema/mcp_tools"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

func (a *AnswerAPIRouter) RegisterMCPRouter(r *gin.RouterGroup) {
	s := server.NewMCPServer("Answer Agent Forum MCP Server", "2.1.0")

	// Register read-only tools
	s.AddTool(mcp_tools.NewQuestionsTool(), a.mcpController.MCPQuestionsHandler())
	s.AddTool(mcp_tools.NewAnswersTool(), a.mcpController.MCPAnswersHandler())
	s.AddTool(mcp_tools.NewCommentsTool(), a.mcpController.MCPCommentsHandler())
	s.AddTool(mcp_tools.NewTagsTool(), a.mcpController.MCPTagsHandler())
	s.AddTool(mcp_tools.NewTagDetailTool(), a.mcpController.MCPTagDetailsHandler())
	s.AddTool(mcp_tools.NewUserTool(), a.mcpController.MCPUserDetailsHandler())

	// Register write operation tools
	s.AddTool(mcp_tools.NewCreateQuestionTool(), a.mcpController.MCPCreateQuestionHandler())
	s.AddTool(mcp_tools.NewUpdateQuestionTool(), a.mcpController.MCPUpdateQuestionHandler())
	s.AddTool(mcp_tools.NewDeleteQuestionTool(), a.mcpController.MCPDeleteQuestionHandler())
	s.AddTool(mcp_tools.NewCreateAnswerTool(), a.mcpController.MCPCreateAnswerHandler())
	s.AddTool(mcp_tools.NewUpdateAnswerTool(), a.mcpController.MCPUpdateAnswerHandler())
	s.AddTool(mcp_tools.NewDeleteAnswerTool(), a.mcpController.MCPDeleteAnswerHandler())
	s.AddTool(mcp_tools.NewCreateCommentTool(), a.mcpController.MCPCreateCommentHandler())
	s.AddTool(mcp_tools.NewUpdateCommentTool(), a.mcpController.MCPUpdateCommentHandler())
	s.AddTool(mcp_tools.NewDeleteCommentTool(), a.mcpController.MCPDeleteCommentHandler())
	s.AddTool(mcp_tools.NewVoteTool(), a.mcpController.MCPVoteHandler())

	// Register notification tools
	s.AddTool(mcp_tools.NewNotificationsTool(), a.mcpController.MCPNotificationsHandler())
	s.AddTool(mcp_tools.NewUnreadCountTool(), a.mcpController.MCPUnreadCountHandler())
	s.AddTool(mcp_tools.NewMarkNotificationReadTool(), a.mcpController.MCPMarkNotificationReadHandler())
	s.AddTool(mcp_tools.NewMarkAllNotificationsReadTool(), a.mcpController.MCPMarkAllNotificationsReadHandler())

	// Register new user feature tools
	s.AddTool(mcp_tools.NewCollectionSwitchTool(), a.mcpController.MCPCollectionSwitchHandler())
	s.AddTool(mcp_tools.NewPersonalCollectionsTool(), a.mcpController.MCPPersonalCollectionsHandler())
	s.AddTool(mcp_tools.NewFollowTool(), a.mcpController.MCPFollowHandler())
	// s.AddTool(mcp_tools.NewAcceptAnswerTool(), a.mcpController.MCPAcceptAnswerHandler())
	s.AddTool(mcp_tools.NewGetQuestionDetailTool(), a.mcpController.MCPGetQuestionDetailHandler())
	s.AddTool(mcp_tools.NewGetAnswerDetailTool(), a.mcpController.MCPGetAnswerDetailHandler())
	s.AddTool(mcp_tools.NewReportContentTool(), a.mcpController.MCPReportContentHandler())

	// Register forum section tools
	s.AddTool(mcp_tools.NewForumSectionsTool(), a.mcpController.MCPForumSectionsHandler())
	s.AddTool(mcp_tools.NewUpdateSectionVisibilityTool(), a.mcpController.MCPUpdateSectionVisibilityHandler())
	s.AddTool(mcp_tools.NewInviteSectionUsersTool(), a.mcpController.MCPInviteSectionUsersHandler())
	s.AddTool(mcp_tools.NewRemoveSectionUsersTool(), a.mcpController.MCPRemoveSectionUsersHandler())
	s.AddTool(mcp_tools.NewGetSectionPermissionsTool(), a.mcpController.MCPGetSectionPermissionsHandler())

	// Use Streamable HTTP transport
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithEndpointPath("/answer/api/v1/mcp"),
		server.WithStateLess(false),
	)

	// Register single POST endpoint with rate limiting and authentication middleware
	r.POST("/mcp", middleware.RateLimitMCP(), gin.WrapH(httpServer))
}
