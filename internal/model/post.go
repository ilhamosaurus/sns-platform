package model

import "github.com/ilhamosaurus/sns-platform/pkg/types"

type Post struct {
	BaseModel
	UserID       int64           `gorm:"column:user_id;not null;index:idx_user_created" json:"user_id"`
	Content      string          `gorm:"type:text" json:"content"`
	MediaType    types.MediaType `gorm:"column:media_type;size:20;index" json:"media_type"` // image, video, text
	MediaURL     string          `gorm:"column:media_url;size:255" json:"media_url"`
	IsPublic     bool            `gorm:"column:is_public;default:true;index" json:"is_public"`
	ViewCount    int64           `gorm:"column:view_count;default:0" json:"view_count"`
	ShareCount   int64           `gorm:"column:share_count;default:0" json:"share_count"`
	LikeCount    int64           `gorm:"column:like_count;default:0" json:"like_count"`
	CommentCount int64           `gorm:"column:comment_count;default:0" json:"comment_count"`

	// Relationships
	User      *User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Comments  []*Comment  `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Reactions []*Reaction `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"reactions,omitempty"`
}
