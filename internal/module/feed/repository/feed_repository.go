package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ilhamosaurus/sns-platform/internal/dto"
	"github.com/ilhamosaurus/sns-platform/pkg/types"
	"gorm.io/gorm"
)

type FeedRepository interface {
	// Define feed-related data access methods here
	GetUserFeed(ctx context.Context, userID int64, limit, offset int) ([]*dto.FeedPost, error)
	GetExploreFeed(ctx context.Context, userID int64, limit, offset int, timeRange time.Duration) ([]*dto.FeedPost, error)
	GetPostWithDetails(ctx context.Context, postID, userID int64) (*dto.PostDetail, error)
}

type feedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) FeedRepository {
	return &feedRepository{db: db}
}

// GetUserFeed retrieves the activity feed for a user (posts from followed users)
// This is an optimized query using the pre-computed ActivityFeed table
func (r *feedRepository) GetUserFeed(ctx context.Context, userID int64, limit, offset int) ([]*dto.FeedPost, error) {
	var feedPosts []*dto.FeedPost

	// Query using the denormalized activity_feeds table for better performance
	err := r.db.WithContext(ctx).Table("activity_feeds").
		Select(`
			posts.*,
			users.id as "author__id",
			users.username as "author__username",
			users.full_name as "author__full_name",
			users.avatar_url as "author__avatar_url",
			users.is_verified as "author__is_verified",
			CASE WHEN user_likes.id IS NOT NULL THEN true ELSE false END as has_user_liked
		`).
		Joins("INNER JOIN posts ON activity_feeds.post_id = posts.id AND posts.deleted_at IS NULL").
		Joins("INNER JOIN users ON posts.user_id = users.id AND users.deleted_at IS NULL").
		Joins(`LEFT JOIN reactions user_likes ON posts.id = user_likes.post_id 
			AND user_likes.user_id = ? 
			AND user_likes.type = 'like' 
			AND user_likes.deleted_at IS NULL`, userID).
		Where("activity_feeds.user_id = ? AND activity_feeds.deleted_at IS NULL", userID).
		Order("activity_feeds.post_created DESC").
		Limit(limit).
		Offset(offset).
		Scan(&feedPosts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user feed: %w", err)
	}

	return feedPosts, nil
}

// GetExploreFeed retrieves trending/popular posts for discovery
func (r *feedRepository) GetExploreFeed(ctx context.Context, userID int64, limit, offset int, timeRange time.Duration) ([]*dto.FeedPost, error) {
	var feedPosts []*dto.FeedPost

	cutoffTime := time.Now().Add(-timeRange)

	err := r.db.WithContext(ctx).Table("posts").
		Select(`
			posts.*,
			users.id as "author__id",
			users.username as "author__username",
			users.full_name as "author__full_name",
			users.avatar_url as "author__avatar_url",
			users.is_verified as "author__is_verified",
			CASE WHEN user_likes.id IS NOT NULL THEN true ELSE false END as has_user_liked,
			(COALESCE(like_counts.count, 0) * 3 + COALESCE(comment_counts.count, 0) * 5 + posts.share_count * 2) as engagement_score
		`).
		Joins("INNER JOIN users ON posts.user_id = users.id AND users.deleted_at IS NULL").
		Joins(`LEFT JOIN reactions user_likes ON posts.id = user_likes.post_id 
			AND user_likes.user_id = ? 
			AND user_likes.type = 'like' 
			AND user_likes.deleted_at IS NULL`, userID).
		Where("posts.is_public = ? AND posts.created_at >= ? AND posts.deleted_at IS NULL", true, cutoffTime).
		Order("engagement_score DESC, posts.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&feedPosts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch explore feed: %w", err)
	}

	return feedPosts, nil
}

func (r *feedRepository) GetPostWithDetails(ctx context.Context, postID, userID int64) (*dto.PostDetail, error) {
	var detail dto.PostDetail

	// Get post with basic stats
	err := r.db.WithContext(ctx).Table("posts").
		Select(`
			posts.*,
			users.id as "author__id",
			users.username as "author__username",
			users.full_name as "author__full_name",
			users.avatar_url as "author__avatar_url",
			users.is_verified as "author__is_verified",
			CASE WHEN user_likes.id IS NOT NULL THEN true ELSE false END as has_user_liked
		`).
		Joins("INNER JOIN users ON posts.user_id = users.id AND users.deleted_at IS NULL").
		Joins(`LEFT JOIN reactions user_likes ON posts.id = user_likes.post_id 
			AND user_likes.user_id = ? 
			AND user_likes.type = 'like' 
			AND user_likes.deleted_at IS NULL`, userID).
		Where("posts.id = ? AND posts.deleted_at IS NULL", postID).
		First(&detail).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	// Get reaction summary
	var reactions []struct {
		Type  types.ReactionType
		Count int64
	}
	r.db.Table("reactions").
		Select("type, COUNT(*) as count").
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Group("type").
		Scan(&reactions)

	detail.ReactionSummary = make(map[string]int64)
	for _, reaction := range reactions {
		detail.ReactionSummary[reaction.Type.String()] = reaction.Count
	}

	// Get comments with nested replies
	detail.Comments, err = r.getCommentsWithReplies(ctx, postID, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	return &detail, nil
}

// getCommentsWithReplies recursively fetches comments and their replies
func (r *feedRepository) getCommentsWithReplies(ctx context.Context, postID, userID int64, parentID *int64) ([]*dto.CommentWithReplies, error) {
	var comments []*dto.CommentWithReplies

	query := r.db.WithContext(ctx).Table("comments").
		Select(`
			comments.*,
			users.id as "author__id",
			users.username as "author__username",
			users.full_name as "author__full_name",
			users.avatar_url as "author__avatar_url",
			CASE WHEN user_likes.id IS NOT NULL THEN true ELSE false END as has_user_liked
		`).
		Joins("INNER JOIN users ON comments.user_id = users.id AND users.deleted_at IS NULL").
		Joins(`LEFT JOIN reactions user_likes ON comments.id = user_likes.comment_id 
			AND user_likes.user_id = ? 
			AND user_likes.type = 'like' 
			AND user_likes.deleted_at IS NULL`, userID).
		Where("comments.post_id = ? AND comments.deleted_at IS NULL", postID).
		Order("comments.created_at ASC")

	if parentID == nil {
		query = query.Where("comments.parent_id IS NULL")
	} else {
		query = query.Where("comments.parent_id = ?", *parentID)
	}

	if err := query.Scan(&comments).Error; err != nil {
		return nil, err
	}

	// Fetch replies for each comment
	for i := range comments {
		replies, err := r.getCommentsWithReplies(ctx, postID, userID, &comments[i].ID)
		if err != nil {
			return nil, err
		}
		comments[i].Replies = replies
	}

	return comments, nil
}
