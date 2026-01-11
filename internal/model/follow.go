package model

import "gorm.io/gorm"

type Follow struct {
	BaseModel
	FollowerID  int64 `gorm:"column:follower_id;not null;index:idx_follower_following,unique" json:"follower_id"`
	FollowingID int64 `gorm:"column:following_id;not null;index:idx_follower_following,unique" json:"following_id"`

	// Relationships
	Follower  *User `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE" json:"follower,omitempty"`
	Following *User `gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE" json:"following,omitempty"`
}

func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.FollowerID == f.FollowingID {
		return gorm.ErrInvalidData
	}
	return nil
}
