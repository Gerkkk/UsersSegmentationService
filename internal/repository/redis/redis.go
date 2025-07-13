package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"main/internal/domain/models"
	"time"
)

const (
	segPrefix = "userSegments"
	ttl       = 5 * time.Minute
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

func (sc *SegmentationCache) SaveUserSegments(key int, val []models.Segment) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal segments: %w", err)
	}

	res := sc.client.Set(context.Background(), fmt.Sprintf("%s:%d", segPrefix, key), data, ttl)
	if res.Err() != nil {
		return fmt.Errorf("Redis set failed: %s", res.Err().Error())
	}

	return nil
}

func (sc *SegmentationCache) TryGetUserSegments(key int) ([]models.Segment, error) {
	res := sc.client.Get(context.Background(), fmt.Sprintf("%s:%d", segPrefix, key))

	if err := res.Err(); err != nil {

		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, fmt.Errorf("Redis get failed: %s", res.Err().Error())
	}

	data, err := res.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes: %w", err)
	}

	var segments []models.Segment
	if err := json.Unmarshal(data, &segments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal segments: %w", err)
	}

	return segments, nil
}

func (sc *SegmentationCache) Invalidate() error {
	err := sc.client.FlushAll(context.Background()).Err()

	if err != nil {
		return err
	}

	return nil
}
