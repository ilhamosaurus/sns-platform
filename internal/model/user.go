package model

type User struct {
	BaseModel
	Username      string `gorm:"column:username;uniqueIndex;size:50;not null" json:"username"`
	Email         string `gorm:"column:email;uniqueIndex;size:100;not null" json:"email"`
	PasswordHash  string `gorm:"column:password;size:255;not null" json:"-"`
	FullName      string `gorm:"column:full_name;size:100" json:"full_name"`
	Bio           string `gorm:"column:bio;type:text" json:"bio"`
	AvatarURL     string `gorm:"column:avatar_url;size:255" json:"avatar_url"`
	IsVerified    bool   `gorm:"column:is_verified;default:false;index" json:"is_verified"`
	IsPrivate     bool   `gorm:"column:is_private;default:false" json:"is_private"`
	FollwingCount int64  `gorm:"column:following_count;default:0" json:"following_count"`
	FollowerCount int64  `gorm:"column:follower_count;default:0" json:"follower_count"`
	PostCount     int64  `gorm:"column:post_count;default:0" json:"post_count"`

	// Relationships
	Posts            []*Post         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"posts,omitempty"`
	Comments         []*Comment      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
	Reactions        []*Reaction     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"reactions,omitempty"`
	Followers        []*Follow       `gorm:"foreignKey:FollowingID;constraint:OnDelete:CASCADE" json:"followers,omitempty"`
	Following        []*Follow       `gorm:"foreignKey:FollowerID;constraint:OnDelete:CASCADE" json:"following,omitempty"`
	SentMessages     []*Message      `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"sent_messages,omitempty"`
	ReceivedMessages []*Message      `gorm:"foreignKey:ReceiverID;constraint:OnDelete:CASCADE" json:"received_messages,omitempty"`
	Notifications    []*Notification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"notifications,omitempty"`
}
