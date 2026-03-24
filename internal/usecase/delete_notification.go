package usecase

import (
	"context"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type DeleteNotification struct {
	repo d.NotifyRepo
}

func NewDeleteNotification(repo d.NotifyRepo) *DeleteNotification {
	return &DeleteNotification{repo: repo}
}

func (uc *DeleteNotification) Exec(ctx context.Context, userID d.UserID, notificationID d.NotificationID) error {
	if userID == 0 || notificationID == "" {
		return ErrValidation
	}

	n, found, err := uc.repo.Get(ctx, notificationID)
	if err != nil {
		return err
	}
	if !found || n.UserID != userID {
		return ErrNotFound
	}

	deleted, err := uc.repo.Delete(ctx, notificationID, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}

	return nil
}
