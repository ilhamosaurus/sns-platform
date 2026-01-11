package repository

import (
	"github.com/ilhamosaurus/sns-platform/internal/model"
	"gorm.io/gorm"
)

type FollowRepository interface {
	Follow(followerID, followingID int64) error
	Unfollow(followerID, followingID int64) error
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

type followRepository struct {
	db *gorm.DB
}

func (r *followRepository) Follow(followerID, followingID int64) error {
	follow := &model.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return r.db.Create(follow).Error
}

func (r *followRepository) Unfollow(followerID, followingID int64) error {
	return r.db.Where("follower_id = ? AND following_id = ? AND deleted_at IS NULL", followerID, followingID).Delete(&model.Follow{}).Error
}
