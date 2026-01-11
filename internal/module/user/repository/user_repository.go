package repository

import (
	"context"
	"fmt"

	"github.com/ilhamosaurus/sns-platform/internal/dto"
	"github.com/ilhamosaurus/sns-platform/internal/model"
	"github.com/ilhamosaurus/sns-platform/pkg/types"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context, query map[string]any, page, pageSize int) ([]*model.User, int64, error)
	Delete(ctx context.Context, id int64) error
	GetUserProfile(ctx context.Context, username string, viewerID int64) (*dto.UserProfile, error)
	UpdateFollowCount(ctx context.Context, username string, action types.Action) error
	UpdatePostCount(ctx context.Context, id int64, action types.Action) error
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, query map[string]any, page, pageSize int) ([]*model.User, int64, error) {
	var (
		users      []*model.User
		totalCount int64
	)

	db := r.db.WithContext(ctx).Model(&model.User{}).Where("deleted_at IS NULL")

	for key, value := range query {
		db = db.Where(key, value)
	}

	if err := db.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Delete(&model.User{}).Error
}

func (r *userRepository) GetUserProfile(ctx context.Context, username string, viewerID int64) (*dto.UserProfile, error) {
	var profile dto.UserProfile

	err := r.db.Table("users").
		Select(`
			users.*,
			CASE WHEN viewer_follows.id IS NOT NULL THEN true ELSE false END as is_following
		`).
		Joins(`LEFT JOIN follows viewer_follows ON users.id = viewer_follows.following_id 
			AND viewer_follows.follower_id = ? 
			AND viewer_follows.deleted_at IS NULL`, viewerID).
		Where("users.username = ? AND users.deleted_at IS NULL", username).
		First(&profile).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	return &profile, nil
}

func (r *userRepository) UpdateFollowCount(ctx context.Context, username string, action types.Action) error {
	var column, expr string
	switch action {
	case types.ActionFollowed:
		column = "follower_count"
		expr = "follower_count + ?"
	case types.ActionUnfollowed:
		column = "follower_count"
		expr = "follower_count - ?"
	case types.ActionFollowing:
		column = "following_count"
		expr = "following_count + ?"
	case types.ActionUnfollowing:
		column = "following_count"
		expr = "following_count - ?"
	default:
		return fmt.Errorf("invalid action type: %s", action.String())
	}
	return r.db.WithContext(ctx).Model(&model.User{}).Where(`username = ? AND deleted_at IS NULL`, username).UpdateColumn(column, gorm.Expr(expr, 1)).Error
}

func (r *userRepository) UpdatePostCount(ctx context.Context, id int64, action types.Action) error {
	var expr string
	switch action {
	case types.ActionCreated:
		expr = "post_count + ?"
	case types.ActionDeleted:
		expr = "post_count - ?"
	default:
		return fmt.Errorf("invalid action type: %s", action.String())
	}

	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", id).UpdateColumn("post_count", gorm.Expr(expr, 1)).Error
}
