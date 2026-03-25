package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type fakeRepo struct {
	byID map[d.NotificationID]d.Notification
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byID: make(map[d.NotificationID]d.Notification)}
}

func (r *fakeRepo) Create(ctx context.Context, n d.NotificationCreateRequest) (d.Notification, error) {
	id := d.NotificationID("n1")
	if n.IdempotencyKey != "" {
		id = d.NotificationID(n.IdempotencyKey)
	}
	created := d.Notification{
		ID:        id,
		UserID:    n.UserID,
		TaskID:    n.TaskID,
		Type:      n.Type,
		Title:     n.Title,
		Body:      n.Body,
		Read:      false,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	r.byID[id] = created
	return created, nil
}

func (r *fakeRepo) Get(ctx context.Context, id d.NotificationID) (d.Notification, bool, error) {
	n, ok := r.byID[id]
	return n, ok, nil
}

func (r *fakeRepo) Update(ctx context.Context, n d.NotificationUpdateRequest) (d.Notification, bool, error) {
	existing, ok := r.byID[n.ID]
	if !ok {
		return d.Notification{}, false, nil
	}
	if n.Read != nil {
		existing.Read = *n.Read
	}
	if n.Title != nil {
		existing.Title = *n.Title
	}
	if n.Body != nil {
		existing.Body = *n.Body
	}
	r.byID[n.ID] = existing
	return existing, true, nil
}

func (r *fakeRepo) Delete(ctx context.Context, id d.NotificationID, uID d.UserID) (bool, error) {
	n, ok := r.byID[id]
	if !ok || n.UserID != uID {
		return false, nil
	}
	delete(r.byID, id)
	return true, nil
}

func (r *fakeRepo) List(ctx context.Context, req d.NotificationListRequest) (d.NotificationListResult, error) {
	if req.UserID == 0 {
		return d.NotificationListResult{}, errors.New("bad user")
	}
	return d.NotificationListResult{Notifications: []d.Notification{}}, nil
}

func TestCreateNotificationValidation(t *testing.T) {
	uc := NewCreateNotification(newFakeRepo())
	_, err := uc.Exec(context.Background(), d.NotificationCreateRequest{})
	if !errors.Is(err, d.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestListNotificationsValidation(t *testing.T) {
	uc := NewListNotifications(newFakeRepo())
	_, err := uc.Exec(context.Background(), d.NotificationListRequest{UserID: 0})
	if !errors.Is(err, d.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	_, err = uc.Exec(context.Background(), d.NotificationListRequest{
		UserID:    1,
		PageSize:  500,
		PageToken: "",
	})
	if !errors.Is(err, d.ErrBadPagination) {
		t.Fatalf("expected ErrBadPagination, got %v", err)
	}
	_, err = uc.Exec(context.Background(), d.NotificationListRequest{
		UserID:    1,
		PageSize:  10,
		PageToken: "bad",
	})
	if !errors.Is(err, d.ErrBadPagination) {
		t.Fatalf("expected ErrBadPagination, got %v", err)
	}
}

func TestMarkReadOwnership(t *testing.T) {
	repo := newFakeRepo()
	repo.byID["n1"] = d.Notification{
		ID:     "n1",
		UserID: 1,
		Read:   false,
	}
	uc := NewMarkRead(repo)

	_, err := uc.Exec(context.Background(), 2, "n1")
	if !errors.Is(err, d.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	updated, err := uc.Exec(context.Background(), 1, "n1")
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if !updated.Read {
		t.Fatal("expected notification to be marked read")
	}
}

func TestDeleteNotificationOwnership(t *testing.T) {
	repo := newFakeRepo()
	repo.byID["n1"] = d.Notification{ID: "n1", UserID: 1}
	uc := NewDeleteNotification(repo)

	if err := uc.Exec(context.Background(), 2, "n1"); !errors.Is(err, d.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := uc.Exec(context.Background(), 1, "n1"); err != nil {
		t.Fatalf("delete notification: %v", err)
	}
}
