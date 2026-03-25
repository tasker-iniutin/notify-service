package postgre

import (
	"context"
	"os"
	"testing"

	d "github.com/tasker-iniutin/notify-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNotifyRepoCreateIdempotent(t *testing.T) {
	db := openNotifyTestDB(t)
	repo := NewNotifyPostgreRepo(db)

	first, err := repo.Create(context.Background(), d.NotificationCreateRequest{
		UserID:         1,
		TaskID:         10,
		Type:           d.NotificationTypeTaskCreated,
		Title:          "t1",
		Body:           "b1",
		IdempotencyKey: "k1",
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	second, err := repo.Create(context.Background(), d.NotificationCreateRequest{
		UserID:         1,
		TaskID:         10,
		Type:           d.NotificationTypeTaskCreated,
		Title:          "t1",
		Body:           "b1",
		IdempotencyKey: "k1",
	})
	if err != nil {
		t.Fatalf("create notification again: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same notification id, got %s and %s", first.ID, second.ID)
	}
}

func TestNotifyRepoListUnreadOnly(t *testing.T) {
	db := openNotifyTestDB(t)
	repo := NewNotifyPostgreRepo(db)

	created, err := repo.Create(context.Background(), d.NotificationCreateRequest{
		UserID: 1, TaskID: 1, Type: d.NotificationTypeTaskCreated, Title: "t", Body: "b",
	})
	if err != nil {
		t.Fatalf("create notification: %v", err)
	}

	read := true
	if _, _, err := repo.Update(context.Background(), d.NotificationUpdateRequest{
		ID:   created.ID,
		Read: &read,
	}); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	res, err := repo.List(context.Background(), d.NotificationListRequest{
		UserID:     1,
		PageSize:   10,
		PageToken:  "",
		UnreadOnly: true,
	})
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(res.Notifications) != 0 {
		t.Fatalf("expected 0 unread, got %d", len(res.Notifications))
	}
}

func openNotifyTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	setupNotifySchema(t, db)
	return db
}

func setupNotifySchema(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	const schema = `
		CREATE TABLE IF NOT EXISTS notifications (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			task_id BIGINT NOT NULL,
			type VARCHAR(64) NOT NULL,
			title TEXT NOT NULL,
			body TEXT NOT NULL DEFAULT '',
			idempotency_key TEXT,
			is_read BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			expires_at TIMESTAMP NOT NULL DEFAULT now()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS ux_notifications_user_idempotency
			ON notifications (user_id, idempotency_key)
			WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
		TRUNCATE TABLE notifications RESTART IDENTITY;
	`

	if _, err := db.Exec(context.Background(), schema); err != nil {
		t.Fatalf("setup notify schema: %v", err)
	}
}
