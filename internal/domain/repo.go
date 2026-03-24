package domain

import (
	"context"
)

type NotifyRepo interface {
	Create(ctx context.Context, n NotificationCreateRequest) (Notification, error)
	Get(ctx context.Context, id NotificationID) (Notification, bool, error)
	Update(ctx context.Context, n NotificationUpdateRequest) (Notification, bool, error)
	Delete(ctx context.Context, id NotificationID, uID UserID) (bool, error)
	List(ctx context.Context, req NotificationListRequest) (NotificationListResult, error)
}
