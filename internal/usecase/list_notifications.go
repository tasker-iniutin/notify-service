package usecase

import (
	"context"
	"math"
	"strconv"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type ListNotifications struct {
	repo d.NotifyRepo
}

func NewListNotifications(repo d.NotifyRepo) *ListNotifications {
	return &ListNotifications{repo: repo}
}

const (
	defaultLimit = uint32(50)
	maxLimit     = uint32(200)
)

func (uc *ListNotifications) Exec(ctx context.Context, req d.NotificationListRequest) (d.NotificationListResult, error) {
	if req.UserID == 0 {
		return d.NotificationListResult{}, d.ErrValidation
	}
	if req.PageSize == 0 {
		req.PageSize = defaultLimit
	}
	if req.PageSize > maxLimit {
		return d.NotificationListResult{}, d.ErrBadPagination
	}
	if _, err := parsePageToken(req.PageToken); err != nil {
		return d.NotificationListResult{}, d.ErrBadPagination
	}
	return uc.repo.List(ctx, req)
}

func parsePageToken(token string) (uint32, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.ParseUint(token, 10, 64)
	if err != nil || offset > math.MaxUint32 {
		return 0, d.ErrBadPagination
	}
	return uint32(offset), nil
}
