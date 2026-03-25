package grpc

import (
	"context"
	"errors"
	"strconv"

	notifypb "github.com/tasker-iniutin/api-contracts/gen/go/proto/notify/v1alpha"
	authctx "github.com/tasker-iniutin/common/authctx"
	d "github.com/tasker-iniutin/notify-service/internal/domain"
	uc "github.com/tasker-iniutin/notify-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	notifypb.UnimplementedNotifyServiceServer
	create *uc.CreateNotification
	list   *uc.ListNotifications
	mark   *uc.MarkRead
	delete *uc.DeleteNotification
}

func NewServer(repo d.NotifyRepo) *Server {
	return &Server{
		create: uc.NewCreateNotification(repo),
		list:   uc.NewListNotifications(repo),
		mark:   uc.NewMarkRead(repo),
		delete: uc.NewDeleteNotification(repo),
	}
}

func (s *Server) CreateNotification(ctx context.Context, req *notifypb.CreateNotificationRequest) (*notifypb.CreateNotificationResponse, error) {
	userID, err := requireUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	taskID, err := parseTaskID(req.GetTaskId())
	if err != nil {
		return nil, err
	}
	nt, err := mapTypeFromPB(req.GetType())
	if err != nil {
		return nil, err
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	created, err := s.create.Exec(ctx, d.NotificationCreateRequest{
		UserID:         userID,
		TaskID:         taskID,
		Type:           nt,
		Title:          req.GetTitle(),
		Body:           req.GetBody(),
		IdempotencyKey: req.GetIdempotencyKey(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	return &notifypb.CreateNotificationResponse{Notification: mapToPB(created)}, nil
}

func (s *Server) ListNotifications(ctx context.Context, req *notifypb.ListNotificationsRequest) (*notifypb.ListNotificationsResponse, error) {
	userID, err := requireUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	res, err := s.list.Exec(ctx, d.NotificationListRequest{
		UserID:     userID,
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
		UnreadOnly: req.GetUnreadOnly(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	out := &notifypb.ListNotificationsResponse{
		Notifications: make([]*notifypb.Notification, 0, len(res.Notifications)),
		NextPageToken: res.NextPageToken,
	}
	for _, n := range res.Notifications {
		out.Notifications = append(out.Notifications, mapToPB(n))
	}
	return out, nil
}

func (s *Server) MarkRead(ctx context.Context, req *notifypb.MarkReadRequest) (*notifypb.MarkReadResponse, error) {
	userID, err := requireUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	if req.GetNotificationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	updated, err := s.mark.Exec(ctx, userID, d.NotificationID(req.GetNotificationId()))
	if err != nil {
		return nil, mapErr(err)
	}

	return &notifypb.MarkReadResponse{Notification: mapToPB(updated)}, nil
}

func (s *Server) DeleteNotification(ctx context.Context, req *notifypb.DeleteNotificationRequest) (*notifypb.DeleteNotificationResponse, error) {
	userID, err := requireUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	if req.GetNotificationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	if err := s.delete.Exec(ctx, userID, d.NotificationID(req.GetNotificationId())); err != nil {
		return nil, mapErr(err)
	}

	return &notifypb.DeleteNotificationResponse{}, nil
}

func parseTaskID(raw string) (d.TaskID, error) {
	if raw == "" {
		return 0, status.Error(codes.InvalidArgument, "task_id is required")
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "invalid task_id")
	}
	return d.TaskID(id), nil
}

func mapTypeFromPB(t notifypb.NotificationType) (d.NotificationType, error) {
	switch t {
	case notifypb.NotificationType_TASK_CREATED:
		return d.NotificationTypeTaskCreated, nil
	case notifypb.NotificationType_TASK_COMPLETED:
		return d.NotificationTypeTaskCompleted, nil
	case notifypb.NotificationType_TASK_DELETED:
		return d.NotificationTypeTaskDeleted, nil
	case notifypb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED:
		return "", status.Error(codes.InvalidArgument, "notification type is required")
	default:
		return "", status.Error(codes.InvalidArgument, "invalid notification type")
	}
}

func mapTypeToPB(t d.NotificationType) notifypb.NotificationType {
	switch t {
	case d.NotificationTypeTaskCreated:
		return notifypb.NotificationType_TASK_CREATED
	case d.NotificationTypeTaskCompleted:
		return notifypb.NotificationType_TASK_COMPLETED
	case d.NotificationTypeTaskDeleted:
		return notifypb.NotificationType_TASK_DELETED
	default:
		return notifypb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}

func mapToPB(n d.Notification) *notifypb.Notification {
	return &notifypb.Notification{
		Id:        string(n.ID),
		UserId:    strconv.FormatUint(uint64(n.UserID), 10),
		Type:      mapTypeToPB(n.Type),
		TaskId:    strconv.FormatUint(uint64(n.TaskID), 10),
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}
}

func requireUser(ctx context.Context, raw string) (d.UserID, error) {
	if raw == "" {
		return 0, status.Error(codes.InvalidArgument, "user_id is required")
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if authID, ok := authctx.UserID(ctx); ok && authID != id {
		return 0, status.Error(codes.PermissionDenied, "user_id does not match token")
	}
	return d.UserID(id), nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, d.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, d.ErrBadPagination):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, d.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
