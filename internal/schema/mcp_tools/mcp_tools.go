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

package mcp_tools

import (
	"github.com/apache/answer/internal/schema"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	MCPToolsList = []mcp.Tool{
		NewQuestionsTool(),
		NewAnswersTool(),
		NewCommentsTool(),
		NewTagsTool(),
		NewTagDetailTool(),
		NewUserTool(),
		NewCreateQuestionTool(),
		NewUpdateQuestionTool(),
		NewDeleteQuestionTool(),
		NewCreateAnswerTool(),
		NewUpdateAnswerTool(),
		NewDeleteAnswerTool(),
		NewCreateCommentTool(),
		NewUpdateCommentTool(),
		NewDeleteCommentTool(),
		NewVoteTool(),
		NewNotificationsTool(),
		NewUnreadCountTool(),
		NewMarkNotificationReadTool(),
		NewMarkAllNotificationsReadTool(),
		// New tools
		NewCollectionSwitchTool(),
		NewPersonalCollectionsTool(),
		NewFollowTool(),
		NewAcceptAnswerTool(),
		NewGetQuestionDetailTool(),
		NewGetAnswerDetailTool(),
		NewReportContentTool(),
		NewForumSectionsTool(),
		NewUpdateSectionVisibilityTool(),
		NewInviteSectionUsersTool(),
		NewRemoveSectionUsersTool(),
		NewGetSectionPermissionsTool(),
	}
)

func NewQuestionsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_questions",
		mcp.WithDescription("Searching for questions that already existed in the system. After the search, you can use the get_answers_by_question_id tool to get answers for the questions. Use get_forum_sections to discover available sections first."),
		mcp.WithString(schema.MCPSearchCondKeyword,
			mcp.Description("Keyword to search for questions. Multiple keywords separated by spaces"),
		),
		mcp.WithString(schema.MCPSearchCondUsername,
			mcp.Description("Search for questions that contain only those created by the specified user"),
		),
		mcp.WithString(schema.MCPSearchCondTag,
			mcp.Description("Filter by tag (semicolon separated for multiple tags)"),
		),
		mcp.WithString(schema.MCPSearchCondScore,
			mcp.Description("Minimum score that the question must have"),
		),
		mcp.WithString(schema.MCPSearchCondSection,
			mcp.Description("Filter by forum section slug name. Use get_forum_sections to discover available sections."),
		),
	)
	return listFilesTool
}

func NewAnswersTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_answers_by_question_id",
		mcp.WithDescription("Search for all answers corresponding to the question ID. The question ID is provided by get_questions tool."),
		mcp.WithString(schema.MCPSearchCondQuestionID,
			mcp.Description("The ID of the question to which the answer belongs. The question ID is provided by get_questions tool."),
		),
	)
	return listFilesTool
}

func NewCommentsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_comments",
		mcp.WithDescription("Searching for comments that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondObjectID,
			mcp.Description("Queries comments on an object, either a question or an answer. object_id is the id of the object."),
		),
	)
	return listFilesTool
}

func NewTagsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_tags",
		mcp.WithDescription("Searching for tags that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondTagName,
			mcp.Description("Tag name"),
		),
	)
	return listFilesTool
}

func NewTagDetailTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_tag_detail",
		mcp.WithDescription("Get detailed information about a specific tag"),
		mcp.WithString(schema.MCPSearchCondTagName,
			mcp.Description("Tag name"),
		),
	)
	return listFilesTool
}

func NewUserTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_user",
		mcp.WithDescription("Searching for users that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondUsername,
			mcp.Description("Username"),
		),
	)
	return listFilesTool
}

// Write operation tools

func NewCreateQuestionTool() mcp.Tool {
	return mcp.NewTool("create_question",
		mcp.WithDescription("Create a new question"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Question title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Question content in markdown format")),
		mcp.WithArray("tags", mcp.Description("Question tags (array of tag names)"), mcp.WithStringItems()),
	)
}

func NewUpdateQuestionTool() mcp.Tool {
	return mcp.NewTool("update_question",
		mcp.WithDescription("Update your own question"),
		mcp.WithString("question_id", mcp.Required(), mcp.Description("Question ID to update")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Updated question title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Updated question content in markdown format")),
		mcp.WithArray("tags", mcp.Description("Updated question tags (array of tag names)"), mcp.WithStringItems()),
		mcp.WithString("edit_summary", mcp.Description("Summary of changes made")),
	)
}

func NewDeleteQuestionTool() mcp.Tool {
	return mcp.NewTool("delete_question",
		mcp.WithDescription("Delete your own question"),
		mcp.WithString("question_id", mcp.Required(), mcp.Description("Question ID to delete")),
	)
}

func NewCreateAnswerTool() mcp.Tool {
	return mcp.NewTool("create_answer",
		mcp.WithDescription("Create a new answer to a question"),
		mcp.WithString("question_id", mcp.Required(), mcp.Description("Question ID to answer")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Answer content in markdown format")),
	)
}

func NewUpdateAnswerTool() mcp.Tool {
	return mcp.NewTool("update_answer",
		mcp.WithDescription("Update your own answer"),
		mcp.WithString("answer_id", mcp.Required(), mcp.Description("Answer ID to update")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Updated answer content in markdown format")),
		mcp.WithString("edit_summary", mcp.Description("Summary of changes made")),
	)
}

func NewDeleteAnswerTool() mcp.Tool {
	return mcp.NewTool("delete_answer",
		mcp.WithDescription("Delete your own answer"),
		mcp.WithString("answer_id", mcp.Required(), mcp.Description("Answer ID to delete")),
	)
}

func NewCreateCommentTool() mcp.Tool {
	return mcp.NewTool("create_comment",
		mcp.WithDescription("Create a comment on a question or answer"),
		mcp.WithString("object_id", mcp.Required(), mcp.Description("ID of the question or answer to comment on")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Comment content")),
	)
}

func NewUpdateCommentTool() mcp.Tool {
	return mcp.NewTool("update_comment",
		mcp.WithDescription("Update your own comment"),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("Comment ID to update")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Updated comment content")),
	)
}

func NewDeleteCommentTool() mcp.Tool {
	return mcp.NewTool("delete_comment",
		mcp.WithDescription("Delete your own comment"),
		mcp.WithString("comment_id", mcp.Required(), mcp.Description("Comment ID to delete")),
	)
}

func NewVoteTool() mcp.Tool {
	return mcp.NewTool("vote",
		mcp.WithDescription("Vote on a question or answer (upvote or downvote)"),
		mcp.WithString("object_id", mcp.Required(), mcp.Description("ID of the question or answer to vote on")),
		mcp.WithString("vote_type", mcp.Required(), mcp.Description("Vote type: 'up' for upvote, 'down' for downvote")),
		mcp.WithBoolean("is_cancel", mcp.Description("Set to true to cancel an existing vote")),
	)
}

func NewNotificationsTool() mcp.Tool {
	return mcp.NewTool("get_notifications",
		mcp.WithDescription("Get the current user's notification list. Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondNotificationType,
			mcp.Description("Notification type: 'inbox' or 'achievement'"),
		),
		mcp.WithString(schema.MCPSearchCondInboxType,
			mcp.Description("Inbox sub-type filter: 'all', 'posts', 'invites', or 'votes'. Only used when type is 'inbox'."),
		),
		mcp.WithNumber(schema.MCPSearchCondPage,
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber(schema.MCPSearchCondPageSize,
			mcp.Description("Page size (default 10)"),
		),
	)
}

func NewUnreadCountTool() mcp.Tool {
	return mcp.NewTool("get_unread_count",
		mcp.WithDescription("Get the current user's unread notification count (red dot). Returns counts for inbox and achievement notifications. Requires authentication via API key."),
	)
}

func NewMarkNotificationReadTool() mcp.Tool {
	return mcp.NewTool("mark_notification_read",
		mcp.WithDescription("Mark a specific notification as read. Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondNotificationID, mcp.Required(),
			mcp.Description("The ID of the notification to mark as read"),
		),
	)
}

func NewMarkAllNotificationsReadTool() mcp.Tool {
	return mcp.NewTool("mark_all_notifications_read",
		mcp.WithDescription("Mark all notifications of a given type as read. Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondNotificationType, mcp.Required(),
			mcp.Description("Notification type to clear: 'inbox' or 'achievement'"),
		),
	)
}

// New tool definitions

func NewCollectionSwitchTool() mcp.Tool {
	return mcp.NewTool("collection_switch",
		mcp.WithDescription("Add or remove a question from your personal collection (bookmark). Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondObjectID, mcp.Required(),
			mcp.Description("The ID of the question to bookmark or remove from collection"),
		),
		mcp.WithBoolean(schema.MCPSearchCondBookmark,
			mcp.Description("true to add to collection (bookmark), false to remove from collection (default: true)"),
		),
		mcp.WithString(schema.MCPSearchCondGroupID,
			mcp.Description("Collection group ID (optional, will use default group if not provided)"),
		),
	)
}

func NewPersonalCollectionsTool() mcp.Tool {
	return mcp.NewTool("get_personal_collections",
		mcp.WithDescription("Get the current user's collection (bookmarked questions). Requires authentication via API key."),
		mcp.WithNumber(schema.MCPSearchCondPage,
			mcp.Description("Page number (default 1)"),
		),
		mcp.WithNumber(schema.MCPSearchCondPageSize,
			mcp.Description("Page size (default 10)"),
		),
	)
}

func NewFollowTool() mcp.Tool {
	return mcp.NewTool("follow",
		mcp.WithDescription("Follow or unfollow a question, tag, or user. Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondObjectID, mcp.Required(),
			mcp.Description("The ID of the question, tag, or user to follow or unfollow"),
		),
		mcp.WithBoolean(schema.MCPSearchCondIsCancel,
			mcp.Description("true to unfollow, false to follow (default: false)"),
		),
	)
}

func NewAcceptAnswerTool() mcp.Tool {
	return mcp.NewTool("accept_answer",
		mcp.WithDescription("Accept an answer as the best answer for a question. Only the question author can accept an answer. Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondQuestionID, mcp.Required(),
			mcp.Description("The ID of the question"),
		),
		mcp.WithString(schema.MCPSearchCondAnswerID, mcp.Required(),
			mcp.Description("The ID of the answer to accept"),
		),
	)
}

func NewGetQuestionDetailTool() mcp.Tool {
	return mcp.NewTool("get_question_detail",
		mcp.WithDescription("Get detailed information about a specific question, including vote count, answer count, and other metadata."),
		mcp.WithString(schema.MCPSearchCondQuestionID, mcp.Required(),
			mcp.Description("The ID of the question"),
		),
	)
}

func NewGetAnswerDetailTool() mcp.Tool {
	return mcp.NewTool("get_answer_detail",
		mcp.WithDescription("Get detailed information about a specific answer."),
		mcp.WithString(schema.MCPSearchCondAnswerID, mcp.Required(),
			mcp.Description("The ID of the answer"),
		),
	)
}

func NewForumSectionsTool() mcp.Tool {
	return mcp.NewTool("get_forum_sections",
		mcp.WithDescription("List all forum sections (boards) that you have access to. Returns section names, slugs, visibility status, and whether you can manage each section. Use section slugs with get_questions to filter questions by section."),
	)
}

func NewUpdateSectionVisibilityTool() mcp.Tool {
	return mcp.NewTool("update_section_visibility",
		mcp.WithDescription("Update the visibility of a forum section. Requires section moderator or admin permissions. Requires authentication via API key."),
		mcp.WithString("section", mcp.Required(), mcp.Description("Section slug name")),
		mcp.WithString("visibility", mcp.Required(), mcp.Description("New visibility: 'public' or 'private'")),
	)
}

func NewInviteSectionUsersTool() mcp.Tool {
	return mcp.NewTool("invite_section_users",
		mcp.WithDescription("Invite users to a forum section as member or moderator. Moderators can invite members; only admins can invite moderators. Requires authentication via API key."),
		mcp.WithString("section", mcp.Required(), mcp.Description("Section slug name")),
		mcp.WithArray("users", mcp.Required(), mcp.Description("Usernames to invite"), mcp.WithStringItems()),
		mcp.WithString("role", mcp.Required(), mcp.Description("Role to assign: 'member' or 'moderator'")),
	)
}

func NewRemoveSectionUsersTool() mcp.Tool {
	return mcp.NewTool("remove_section_users",
		mcp.WithDescription("Remove users from a forum section role. Moderators can remove members; only admins can remove moderators. Requires authentication via API key."),
		mcp.WithString("section", mcp.Required(), mcp.Description("Section slug name")),
		mcp.WithArray("users", mcp.Required(), mcp.Description("Usernames to remove"), mcp.WithStringItems()),
		mcp.WithString("role", mcp.Required(), mcp.Description("Role to remove from: 'member' or 'moderator'")),
	)
}

func NewGetSectionPermissionsTool() mcp.Tool {
	return mcp.NewTool("get_section_permissions",
		mcp.WithDescription("Get the member and moderator lists of a forum section. Requires section moderator or admin permissions. Requires authentication via API key."),
		mcp.WithString("section", mcp.Required(), mcp.Description("Section slug name")),
	)
}

func NewReportContentTool() mcp.Tool {
	return mcp.NewTool("report_content",
		mcp.WithDescription("Report inappropriate content (question, answer, or comment). Requires authentication via API key."),
		mcp.WithString(schema.MCPSearchCondObjectID, mcp.Required(),
			mcp.Description("The ID of the content to report"),
		),
		mcp.WithNumber(schema.MCPSearchCondReportType, mcp.Required(),
			mcp.Description("Report type ID (internal config ID, e.g., 1 for spam, 2 for duplicate, etc.)"),
		),
		mcp.WithString(schema.MCPSearchCondContent,
			mcp.Description("Additional details about the report (optional)"),
		),
	)
}
