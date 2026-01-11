package model

import (
	"github.com/ilhamosaurus/sns-platform/pkg/types"
	"gorm.io/gorm"
)

type Reaction struct {
	BaseModel
	UserID    int64              `gorm:"column:user_id;not null;index:idx_user_target" json:"user_id"`
	PostID    *int64             `gorm:"column:post_id;index:idx_user_target" json:"post_id"`
	CommentID *int64             `gorm:"column:comment_id;index:idx_user_target" json:"comment_id"`
	Type      types.ReactionType `gorm:"column:type;size:20;not null;index" json:"type"` // like, love, haha, wow, sad, angry

	// Relationships
	User    *User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post    *Post    `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Comment *Comment `gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE" json:"comment,omitempty"`
}

func (r *Reaction) BeforeCreate(tx *gorm.DB) error {
	if (r.PostID == nil && r.CommentID == nil) || (r.PostID != nil && r.CommentID != nil) {
		return gorm.ErrInvalidData
	}
	return nil
}
