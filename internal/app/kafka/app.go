package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// MessageHandler - интерфейс обработчика сообщений кафки, который внедряется в кафка-потребитель
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg kafka.Message) error
}

// App - структура Kafka consumer-а, который используется приложением. Группа и набор топиков задаются в конфигах.
type App struct {
	log          *slog.Logger
	handler      MessageHandler
	topics       []string
	brokers      []string
	groupID      string
	shutdownLock sync.Mutex
	readers      []*kafka.Reader
}

// New - конструктор App
func New(
	log *slog.Logger,
	handler MessageHandler,
	brokers []string,
	topics []string,
	groupID string,
) *App {
	return &App{
		log:     log,
		handler: handler,
		brokers: brokers,
		topics:  topics,
		groupID: groupID,
	}
}

// MustRun - Запуск Kafka consumer-а. При ошибке паникует
func (a *App) MustRun(ctx context.Context) {
	const op = "kafkaapp.Run"
	log := a.log.With(slog.String("op", op))

	log.Info("starting Kafka consumers",
		slog.Any("topics", a.topics),
		slog.Any("brokers", a.brokers))

	for _, topic := range a.topics {
		go a.consumeTopic(ctx, topic)
	}

	<-ctx.Done()
	_ = a.Stop()
}

// consumeTopic - добавление топика для прослушивания consumer-ом
func (a *App) consumeTopic(ctx context.Context, topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: a.brokers,
		Topic:   topic,
		GroupID: a.groupID,
	})

	a.shutdownLock.Lock()
	a.readers = append(a.readers, r)
	a.shutdownLock.Unlock()

	defer func() {
		if err := r.Close(); err != nil {
			a.log.Error("failed to close Kafka reader",
				slog.String("topic", topic),
				slog.Any("error", err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			a.log.Info("stopping consumer", slog.String("topic", topic))
			return
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() == nil {
					a.log.Error("failed to read message",
						slog.String("topic", topic),
						slog.Any("error", err))
					time.Sleep(time.Second * 2)
					continue
				}
				return
			}
			go a.processMessage(ctx, m)
		}
	}
}

// processMessage - функция обработки сообщения consumer-ом. Передает сообщение в handler
func (a *App) processMessage(ctx context.Context, m kafka.Message) {
	if err := a.handler.HandleMessage(ctx, m); err != nil {
		a.log.Error("message handling failed",
			slog.String("topic", m.Topic),
			slog.Any("error", err))
	}
}

// Stop - stop kafka consumer-а
func (a *App) Stop() error {
	const op = "kafkaapp.Stop"
	log := a.log.With(slog.String("op", op))

	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	log.Info("stopping Kafka consumers")

	var errs []error
	for _, r := range a.readers {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
			log.Error("failed to close reader", slog.Any("error", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s: %w", op, errs[0])
	}
	return nil
}
