package types

import "strings"

type MediaType uint32

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeImage
	MediaTypeVideo
	MediaTypeText
)

func (mt MediaType) String() string {
	switch mt {
	case MediaTypeImage:
		return "image"
	case MediaTypeVideo:
		return "video"
	case MediaTypeText:
		return "text"
	default:
		return "unknown"
	}
}

func StringToMediaType(s string) MediaType {
	switch strings.ToLower(s) {
	case "image":
		return MediaTypeImage
	case "video":
		return MediaTypeVideo
	case "text":
		return MediaTypeText
	default:
		return MediaTypeUnknown
	}
}

type ReactionType uint32

const (
	ReactionTypeUnknown ReactionType = iota
	ReactionTypeLike
	ReactionTypeLove
	ReactionTypeHaha
	ReactionTypeWow
	ReactionTypeSad
	ReactionTypeAngry
)

func (rt ReactionType) String() string {
	switch rt {
	case ReactionTypeLike:
		return "like"
	case ReactionTypeLove:
		return "love"
	case ReactionTypeHaha:
		return "haha"
	case ReactionTypeWow:
		return "wow"
	case ReactionTypeSad:
		return "sad"
	case ReactionTypeAngry:
		return "angry"
	default:
		return "unknown"
	}
}

func StringToReactionType(s string) ReactionType {
	switch strings.ToLower(s) {
	case "like":
		return ReactionTypeLike
	case "love":
		return ReactionTypeLove
	case "haha":
		return ReactionTypeHaha
	case "wow":
		return ReactionTypeWow
	case "sad":
		return ReactionTypeSad
	case "angry":
		return ReactionTypeAngry
	default:
		return ReactionTypeUnknown
	}
}

type NotificationType uint32

const (
	NotificationTypeUnknown NotificationType = iota
	NotificationTypeFollow
	NotificationTypeLike
	NotificationTypeComment
	NotificationTypeMention
)

func (nt NotificationType) String() string {
	switch nt {
	case NotificationTypeFollow:
		return "follow"
	case NotificationTypeLike:
		return "like"
	case NotificationTypeComment:
		return "comment"
	case NotificationTypeMention:
		return "mention"
	default:
		return "unknown"
	}
}

func StringToNotificationType(s string) NotificationType {
	switch strings.ToLower(s) {
	case "follow":
		return NotificationTypeFollow
	case "like":
		return NotificationTypeLike
	case "comment":
		return NotificationTypeComment
	case "mention":
		return NotificationTypeMention
	default:
		return NotificationTypeUnknown
	}
}

type NotificationTarget uint32

const (
	NotificationTargetUnknown NotificationTarget = iota
	NotificationTargetPost
	NotificationTargetComment
	NotificationTargetUser
)

func (nt NotificationTarget) String() string {
	switch nt {
	case NotificationTargetPost:
		return "post"
	case NotificationTargetComment:
		return "comment"
	case NotificationTargetUser:
		return "user"
	default:
		return "unknown"
	}
}

func StringToNotificationTarget(s string) NotificationTarget {
	switch strings.ToLower(s) {
	case "post":
		return NotificationTargetPost
	case "comment":
		return NotificationTargetComment
	case "user":
		return NotificationTargetUser
	default:
		return NotificationTargetUnknown
	}
}

type Action uint32

const (
	ActionUnknown Action = iota
	ActionFollowing
	ActionUnfollowing
	ActionFollowed
	ActionUnfollowed
	ActionCreated
	ActionDeleted
	ActionLiked
	ActionUnliked
	ActionCommented
	ActionUncommented
	ActionShared
)

func (a Action) String() string {
	switch a {
	case ActionFollowing:
		return "following"
	case ActionUnfollowing:
		return "unfollowing"
	case ActionFollowed:
		return "followed"
	case ActionUnfollowed:
		return "unfollowed"
	case ActionCreated:
		return "created"
	case ActionDeleted:
		return "deleted"
	case ActionLiked:
		return "liked"
	case ActionUnliked:
		return "unliked"
	case ActionCommented:
		return "commented"
	case ActionUncommented:
		return "uncommented"
	case ActionShared:
		return "shared"
	default:
		return "unknown"
	}
}
