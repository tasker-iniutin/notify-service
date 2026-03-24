package usecase

import (
	"context"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type CreateNotification struct {
	repo d.NotifyRepo
}

func NewCreateNotification(repo d.NotifyRepo) *CreateNotification {
	return &CreateNotification{repo: repo}
}

func (uc *CreateNotification) Exec(ctx context.Context, req d.NotificationCreateRequest) (d.Notification, error) {
	if req.UserID == 0 || req.TaskID == 0 || req.Type == "" || req.Title == "" {
		return d.Notification{}, d.ErrValidation
	}
	return uc.repo.Create(ctx, req)
}
