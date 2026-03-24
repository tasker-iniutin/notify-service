package usecase

import (
	"context"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type MarkRead struct {
	repo d.NotifyRepo
}

func NewMarkRead(repo d.NotifyRepo) *MarkRead {
	return &MarkRead{repo: repo}
}

func (uc *MarkRead) Exec(ctx context.Context, userID d.UserID, notificationID d.NotificationID) (d.Notification, error) {
	if userID == 0 || notificationID == "" {
		return d.Notification{}, d.ErrValidation
	}

	n, found, err := uc.repo.Get(ctx, notificationID)
	if err != nil {
		return d.Notification{}, err
	}
	if !found || n.UserID != userID {
		return d.Notification{}, d.ErrNotFound
	}

	read := true
	updated, found, err := uc.repo.Update(ctx, d.NotificationUpdateRequest{
		ID:   notificationID,
		Read: &read,
	})
	if err != nil {
		return d.Notification{}, err
	}
	if !found {
		return d.Notification{}, d.ErrNotFound
	}

	return updated, nil
}
