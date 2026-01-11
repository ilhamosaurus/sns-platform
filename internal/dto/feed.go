package dto

import "github.com/ilhamosaurus/sns-platform/internal/model"

type FeedPost struct {
	*model.Post
	Author       *model.User `json:"author"`
	HasUserLiked bool        `json:"has_user_liked"`
	HasUserSaved bool        `json:"has_user_saved"`
}

type PostDetail struct {
	*FeedPost
	Comments        []*CommentWithReplies `json:"comments"`
	ReactionSummary map[string]int64      `json:"reaction_summary"`
}

type CommentWithReplies struct {
	*model.Comment
	Author       *model.User           `json:"author"`
	HasUserLiked bool                  `json:"has_user_liked"`
	Replies      []*CommentWithReplies `json:"replies,omitempty"`
}
