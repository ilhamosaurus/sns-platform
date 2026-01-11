package model

import "time"

type Message struct {
	BaseModel
	SenderID   int64      `gorm:"column:sender_id;not null;index:idx_sender_receiver" json:"sender_id"`
	ReceiverID int64      `gorm:"column:receiver_id;not null;index:idx_sender_receiver" json:"receiver_id"`
	Content    string     `gorm:"column:content;type:text;not null" json:"content"`
	MediaURL   string     `gorm:"column:media_url;size:255" json:"media_url"`
	IsRead     bool       `gorm:"column:is_read;default:false;index" json:"is_read"`
	ReadAt     *time.Time `gorm:"column:read_at" json:"read_at"`

	// Relationships
	Sender   *User `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sender,omitempty"`
	Receiver *User `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE" json:"receiver,omitempty"`
}
