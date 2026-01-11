package model

import "github.com/ilhamosaurus/sns-platform/pkg/types"

type Notification struct {
	BaseModel
	UserID     int64                    `gorm:"column:user_id;not null;index:idx_user_read_created" json:"user_id"`
	ActorID    int64                    `gorm:"column:actor_id;not null;index" json:"actor_id"` // User who triggered the notification
	Type       types.NotificationType   `gorm:"column:type;size:50;not null;index" json:"type"` // follow, like, comment, mention
	TargetType types.NotificationTarget `gorm:"column:target_type;size:50" json:"target_type"`  // post, comment, user
	TargetID   int64                    `gorm:"column:target_id;index" json:"target_id"`
	Message    string                   `gorm:"column:message;type:text" json:"message"`
	IsRead     bool                     `gorm:"column:is_read;default:false;index:idx_user_read_created" json:"is_read"`

	// Relationships
	User  *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Actor *User `gorm:"foreignKey:ActorID;constraint:OnDelete:CASCADE" json:"actor,omitempty"`
}
