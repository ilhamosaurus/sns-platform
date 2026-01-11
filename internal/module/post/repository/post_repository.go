package repository

import (
	"context"

	"github.com/ilhamosaurus/sns-platform/internal/model"
	"github.com/ilhamosaurus/sns-platform/pkg/types"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, id int64, updates map[string]any) error
	GetByID(ctx context.Context, id int64) (*model.Post, error)
	List(ctx context.Context, query map[string]any, page, pageSize int) ([]*model.Post, int64, error)
	Delete(ctx context.Context, id int64) error
	UpdatePostCount(ctx context.Context, id int64, action types.Action) error
}

type postRepository struct {
	db *gorm.DB
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) Update(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).Error
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, query map[string]any, page, pageSize int) ([]*model.Post, int64, error) {
	var (
		posts      []*model.Post
		totalCount int64
	)

	db := r.db.WithContext(ctx).Model(&model.Post{}).Where("deleted_at IS NULL")

	for key, value := range query {
		db = db.Where(key, value)
	}

	if err := db.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, totalCount, nil
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).Delete(&model.Post{}).Error
}

func (r *postRepository) UpdatePostCount(ctx context.Context, id int64, action types.Action) error {
	var column, expr string
	switch action {
	case types.ActionLiked:
		column = "like_count"
		expr = column + " + 1"
	case types.ActionUnliked:
		column = "like_count"
		expr = "GREATEST(" + column + " - 1, 0)"
	case types.ActionCommented:
		column = "comment_count"
		expr = column + " + 1"
	case types.ActionUncommented:
		column = "comment_count"
		expr = "GREATEST(" + column + " - 1, 0)"
	case types.ActionShared:
		column = "share_count"
		expr = column + " + 1"
	default:
		return nil
	}

	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ? AND deleted_at IS NULL", id).UpdateColumn(column, gorm.Expr(expr)).Error
}
