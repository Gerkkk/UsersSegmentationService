package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"main/internal/domain/models"
)

type SegmentationCache struct {
	log    *slog.Logger
	client *redis.Client
}

func NewSegmentationCache(log *slog.Logger, host, port, password, maxMemory, maxMemPolicy string) (*SegmentationCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	ping, err := client.Ping(context.Background()).Result()
	if err != nil || ping != "PONG" {
		return nil, err
	}

	err = client.ConfigSet(context.Background(), "maxmemory", maxMemory).Err()
	if err != nil {
		return nil, err
	}

	err = client.ConfigSet(context.Background(), "maxmemory-policy", maxMemPolicy).Err()
	if err != nil {
		return nil, err
	}

	return &SegmentationCache{client: client, log: log}, nil
}

func (sc *SegmentationCache) SaveUserSegments(key models.User, val []models.Segment) error {
	panic("implement cache pls")
}

func (sc *SegmentationCache) TryGetUserSegments(key models.User) ([]models.Segment, error) {
	panic("implement cache pls")
}

func (sc *SegmentationCache) Invalidate() error {
	panic("implement cache pls")
}
