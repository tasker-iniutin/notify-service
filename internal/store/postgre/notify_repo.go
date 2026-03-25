package postgre

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	d "github.com/tasker-iniutin/notify-service/internal/domain"
)

type notifyRepoImpl struct {
	db *pgxpool.Pool
}

func NewNotifyPostgreRepo(db *pgxpool.Pool) *notifyRepoImpl {
	return &notifyRepoImpl{db: db}
}

func (r *notifyRepoImpl) Create(ctx context.Context, n d.NotificationCreateRequest) (d.Notification, error) {
	if n.IdempotencyKey != "" {
		if existing, ok, err := r.getByIdempotency(ctx, n.UserID, n.IdempotencyKey); err != nil {
			return d.Notification{}, err
		} else if ok {
			return existing, nil
		}
	}

	const q = `
		INSERT INTO notifications (user_id, task_id, type, title, body, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, task_id, type, title, body, is_read, created_at, expires_at
	`
	var notification d.Notification
	err := r.db.QueryRow(
		ctx,
		q,
		n.UserID,
		n.TaskID, n.Type, n.Title, n.Body, n.IdempotencyKey).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.TaskID,
		&notification.Type,
		&notification.Title,
		&notification.Body,
		&notification.Read,
		&notification.CreatedAt,
		&notification.ExpiresAt,
	)
	if err != nil {
		if n.IdempotencyKey != "" && isUniqueViolation(err) {
			if existing, ok, err := r.getByIdempotency(ctx, n.UserID, n.IdempotencyKey); err == nil && ok {
				return existing, nil
			}
		}
		return d.Notification{}, fmt.Errorf("create notification: %w", err)
	}
	return notification, nil
}
func (r *notifyRepoImpl) Get(ctx context.Context, id d.NotificationID) (d.Notification, bool, error) {
	const q = `
		SELECT id, user_id, task_id, type, title, body, is_read, created_at, expires_at
		FROM notifications
		WHERE id = $1
	`

	var n d.Notification
	err := r.db.QueryRow(ctx, q, id).Scan(
		&n.ID,
		&n.UserID,
		&n.TaskID,
		&n.Type,
		&n.Title,
		&n.Body,
		&n.Read,
		&n.CreatedAt,
		&n.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Notification{}, false, nil
		}
		return d.Notification{}, false, fmt.Errorf("get notification: %w", err)
	}
	return n, true, nil
}
func (r *notifyRepoImpl) Update(ctx context.Context, n d.NotificationUpdateRequest) (d.Notification, bool, error) {
	const q = `
		UPDATE notifications
		SET title = COALESCE($2, title),
		    body = COALESCE($3, body),
		    is_read = COALESCE($4, is_read)
		WHERE id = $1
		RETURNING id, user_id, task_id, type, title, body, is_read, created_at, expires_at
	`

	var updated d.Notification
	err := r.db.QueryRow(ctx, q, n.ID, n.Title, n.Body, n.Read).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.TaskID,
		&updated.Type,
		&updated.Title,
		&updated.Body,
		&updated.Read,
		&updated.CreatedAt,
		&updated.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Notification{}, false, nil
		}
		return d.Notification{}, false, fmt.Errorf("update notification: %w", err)
	}

	return updated, true, nil
}
func (r *notifyRepoImpl) Delete(ctx context.Context, id d.NotificationID, uID d.UserID) (bool, error) {
	const query = `
		DELETE FROM notifications
		WHERE id = $1 AND user_id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, id, uID)
	if err != nil {
		return false, fmt.Errorf("delete task: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}
func (r *notifyRepoImpl) List(ctx context.Context, req d.NotificationListRequest) (d.NotificationListResult, error) {
	const q = `
		SELECT id, user_id, task_id, type, title, body, is_read, created_at, expires_at
		FROM notifications
		WHERE user_id = $1
		  AND ($2 = false OR is_read = false)
		ORDER BY id ASC
		LIMIT $3 OFFSET $4
	`

	offset := uint32(0)
	if req.PageToken != "" {
		if v, err := strconv.ParseUint(req.PageToken, 10, 32); err == nil {
			offset = uint32(v)
		}
	}

	rows, err := r.db.Query(ctx, q, req.UserID, req.UnreadOnly, req.PageSize, offset)
	if err != nil {
		return d.NotificationListResult{}, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := make([]d.Notification, 0)
	for rows.Next() {
		var n d.Notification
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.TaskID,
			&n.Type,
			&n.Title,
			&n.Body,
			&n.Read,
			&n.CreatedAt,
			&n.ExpiresAt,
		); err != nil {
			return d.NotificationListResult{}, fmt.Errorf("scan notification: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return d.NotificationListResult{}, fmt.Errorf("iterate notifications: %w", err)
	}

	next := ""
	if len(out) == int(req.PageSize) && req.PageSize > 0 {
		next = strconv.FormatUint(uint64(offset+req.PageSize), 10)
	}

	return d.NotificationListResult{
		Notifications: out,
		NextPageToken: next,
	}, nil
}

func (r *notifyRepoImpl) getByIdempotency(ctx context.Context, userID d.UserID, key string) (d.Notification, bool, error) {
	const q = `
		SELECT id, user_id, task_id, type, title, body, is_read, created_at, expires_at
		FROM notifications
		WHERE user_id = $1 AND idempotency_key = $2
	`

	var n d.Notification
	err := r.db.QueryRow(ctx, q, userID, key).Scan(
		&n.ID,
		&n.UserID,
		&n.TaskID,
		&n.Type,
		&n.Title,
		&n.Body,
		&n.Read,
		&n.CreatedAt,
		&n.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Notification{}, false, nil
		}
		return d.Notification{}, false, fmt.Errorf("get by idempotency: %w", err)
	}
	return n, true, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
