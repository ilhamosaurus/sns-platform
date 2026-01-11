package model

import "time"

type ActivityFeed struct {
	BaseModel
	UserID      int64     `gorm:"column:user_id;not null;index:idx_user_created" json:"user_id"`
	PostID      int64     `gorm:"column:post_id;not null;index" json:"post_id"`
	AuthorID    int64     `gorm:"column:author_id;not null;index" json:"author_id"`
	PostCreated time.Time `gorm:"column:post_created;not null;index:idx_user_created" json:"post_created"`

	// Relationships
	User   *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Post   *Post `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"post,omitempty"`
	Author *User `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE" json:"author,omitempty"`
}

func (ActivityFeed) TableName() string {
	return "activity_feeds"
}
