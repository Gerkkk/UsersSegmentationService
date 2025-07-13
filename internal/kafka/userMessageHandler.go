package kafkahandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal/domain/events"
	"main/internal/domain/models"

	"github.com/segmentio/kafka-go"
)

type UserService interface {
	CreateUser(user models.User) (int, error)
	DeleteUser(id int) (int, error)
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
	var event events.NewUserEvent

	if err := json.Unmarshal(data, &event); err != nil {
		h.log.Error("failed to parse user create message", slog.String("error", err.Error()))
		return fmt.Errorf("unmarshal user create message: %w", err)
	}

	_, err := h.userSvc.CreateUser(models.User{Id: event.ID})
	if err != nil {
		h.log.Error("failed to create user", slog.String("error", err.Error()))
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (h *Handler) handleUserDelete(ctx context.Context, data []byte) error {
	var event events.DeleteUserEvent

	if err := json.Unmarshal(data, &event); err != nil {
		h.log.Error("failed to parse user delete message", slog.String("error", err.Error()))
		return fmt.Errorf("unmarshal user delete message: %w", err)
	}

	_, err := h.userSvc.DeleteUser(event.ID)
	if err != nil {
		h.log.Error("failed to delete user", slog.String("error", err.Error()))
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
