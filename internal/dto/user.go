package dto

import "github.com/ilhamosaurus/sns-platform/internal/model"

type UserProfile struct {
	model.User
	FollowerCount  int64 `json:"follower_count"`
	FollowingCount int64 `json:"following_count"`
	PostCount      int64 `json:"post_count"`
	IsFollowing    bool  `json:"is_following"`
}
