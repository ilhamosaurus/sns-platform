package model

type Comment struct {
	BaseModel
	PostID       int64  `gorm:"column:post_id;not null;index:idx_post_created" json:"post_id"`
	UserID       int64  `gorm:"column:user_id;not null;index" json:"user_id"`
	ParentID     *int64 `gorm:"column:parent_id;index" json:"parent_id"` // For nested comments/replies
	Content      string `gorm:"column:content;type:text;not null" json:"content"`
	LikesCount   int64  `gorm:"column:likes_count;default:0" json:"likes_count"`
	RepliesCount int64  `gorm:"column:replies_count;default:0" json:"replies_count"`

	// Relationships
	Post      *Post       `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	User      *User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Parent    *Comment    `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"parent,omitempty"`
	Replies   []*Comment  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"replies,omitempty"`
	Reactions []*Reaction `gorm:"foreignKey:CommentID;constraint:OnDelete:CASCADE" json:"reactions,omitempty"`
}
