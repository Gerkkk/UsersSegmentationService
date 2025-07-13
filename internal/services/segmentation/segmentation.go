package segmentation

import (
	"log/slog"
	"main/internal/domain/models"
	apperrors "main/internal/errors"
)

// Segmentation - структура сервиса для управления сегментами
type Segmentation struct {
	log   *slog.Logger
	repo  SegmentationRepository
	cache SegmentationCache
}

type SegmentationRepository interface {
	CreateSegment(segment models.Segment) (string, error)
	DeleteSegment(id string) (string, error)
	UpdateSegment(id string, newSegment models.Segment) (string, error)
	GetUserSegments(id int) ([]models.Segment, error)
	GetSegmentInfo(id string) (models.SegmentInfo, error)
	DistributeSegment(id string, usersPercentage int) (string, error)
}

type SegmentationCache interface {
	SaveUserSegments(key int, val []models.Segment) error
	TryGetUserSegments(key int) ([]models.Segment, error)
	Invalidate() error
}

func NewSegmentation(log *slog.Logger, repo SegmentationRepository, cache SegmentationCache) *Segmentation {
	return &Segmentation{log: log, repo: repo, cache: cache}
}

// CreateSegment - создать сегмент с заданной структурой
func (s *Segmentation) CreateSegment(segment models.Segment) (string, error) {
	id, err := s.repo.CreateSegment(segment)

	if err != nil {
		err = apperrors.Convert(s.log, err)
		return "", err
	}

	return id, nil
}

// DeleteSegment - удалить сегмент по id
func (s *Segmentation) DeleteSegment(id string) (string, error) {
	id, err := s.repo.DeleteSegment(id)

	if err != nil {
		err = apperrors.Convert(s.log, err)
		return "", err
	}

	err = s.cache.Invalidate()

	if err != nil {
		s.log.Error("failed to invalidate cache segmentation", slog.String("error", err.Error()))
	}

	return id, nil
}

// UpdateSegment - исправить поля сегмента с id на поля newSegment
func (s *Segmentation) UpdateSegment(id string, newSegment models.Segment) (string, error) {
	id, err := s.repo.UpdateSegment(id, newSegment)

	if err != nil {
		err = apperrors.Convert(s.log, err)
		return "", err
	}

	err = s.cache.Invalidate()

	if err != nil {
		s.log.Error("failed to invalidate cache segmentation", slog.String("error", err.Error()))
	}

	return id, nil
}

// GetUserSegments - получить сегменты по id пользователя
func (s *Segmentation) GetUserSegments(id int) ([]models.Segment, error) {
	cachedSegments, err := s.cache.TryGetUserSegments(id)

	if err != nil {
		s.log.Error("failed to fetch cached segmentations", slog.String("error", err.Error()))
	}

	if cachedSegments != nil {
		return cachedSegments, nil
	}

	segments, err := s.repo.GetUserSegments(id)
	if err != nil {
		err = apperrors.Convert(s.log, err)
		return nil, err
	}

	err = s.cache.SaveUserSegments(id, segments)
	if err != nil {
		s.log.Error("failed to cache segmentation", slog.String("error", err.Error()))
	}

	return segments, nil
}

// GetSegmentInfo - Получить статистику сегмента по id
func (s *Segmentation) GetSegmentInfo(id string) (models.SegmentInfo, error) {
	res, err := s.repo.GetSegmentInfo(id)

	if err != nil {
		err = apperrors.Convert(s.log, err)
		return models.SegmentInfo{}, err
	}

	return res, nil
}

// DistributeSegment - рспространить сегмент id на заданный процент пользователей, если он еще не распространен
func (s *Segmentation) DistributeSegment(id string, usersPercentage int) (string, error) {
	id, err := s.repo.DistributeSegment(id, usersPercentage)

	if err != nil {
		err = apperrors.Convert(s.log, err)
		return "", err
	}

	err = s.cache.Invalidate()

	if err != nil {
		s.log.Error("failed to invalidate cache segmentation", slog.String("error", err.Error()))
	}

	return id, nil
}
