package kafkahandler

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type UserService interface {
	CreateUser(id string) (string, error)
	DeleteUser(id string) (string, error)
}

type Handler struct {
	log     *slog.Logger
	userSvc UserService
}

func New(
	log *slog.Logger,
	userSvc UserService,
) *Handler {
	return &Handler{
		log:     log,
		userSvc: userSvc,
	}
}

func (h *Handler) HandleMessage(ctx context.Context, msg kafka.Message) error {
	h.log.Debug("message received",
		slog.String("topic", msg.Topic),
		slog.Int("partition", msg.Partition),
		slog.Int64("offset", msg.Offset))

	switch msg.Topic {
	case "create-user":
		return h.handleUserCreate(ctx, msg.Value)
	case "delete-user":
		return h.handleUserDelete(ctx, msg.Value)
	default:
		h.log.Warn("unhandled topic", slog.String("topic", msg.Topic))
		return nil
	}
}

func (h *Handler) handleUserCreate(ctx context.Context, data []byte) error {
	h.userSvc.CreateUser("KEK")
	//var event segmentation.Event
	//if err := json.Unmarshal(data, &event); err != nil {
	//	return fmt.Errorf("unmarshal segment event: %w", err)
	//}
	return nil
}

func (h *Handler) handleUserDelete(ctx context.Context, data []byte) error {
	h.userSvc.DeleteUser("KEK")
	return nil
}
