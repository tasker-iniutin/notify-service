package domain

import "time"

type NotificationID string
type UserID uint64
type TaskID uint64

type NotificationType string

const (
	NotificationTypeTaskCreated   NotificationType = "task_created"
	NotificationTypeTaskCompleted NotificationType = "task_completed"
	NotificationTypeTaskDeleted   NotificationType = "task_deleted"
	NotificationTypeDeadlineDue   NotificationType = "deadline_due"
)

type Notification struct {
	ID        NotificationID
	UserID    UserID
	TaskID    TaskID
	Type      NotificationType
	Title     string
	Body      string
	Read      bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

type NotificationCreateRequest struct {
	UserID UserID
	TaskID TaskID
	Type   NotificationType
	Title  string
	Body   string
}

type NotificationUpdateRequest struct {
	ID    NotificationID
	Read  *bool
	Title *string
	Body  *string
}

type NotificationListRequest struct {
	UserID     UserID
	PageSize   uint32
	PageToken  string
	UnreadOnly bool
}

type NotificationListResult struct {
	Notifications []Notification
	NextPageToken string
}
