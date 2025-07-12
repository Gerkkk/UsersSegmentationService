package segmentation

import (
	"log/slog"
	"main/internal/domain/models"
)

type Segmentation struct {
	log   *slog.Logger
	repo  SegmentationRepository
	cache SegmentationCache
}

type SegmentationRepository interface {
	CreateSegment(id, description string) (string, error)
	DeleteSegment(id string) (string, error)
	UpdateSegment(id string, newSegment models.Segment) (string, error)
	GetUserSegments(id string) ([]models.Segment, error)
	GetSegmentInfo(id string) (models.SegmentInfo, error)
	DistributeSegment(id string, usersPercentage int) (string, error)
}

type SegmentationCache interface {
	SaveUserSegments(key models.User, val []models.Segment) error
	TryGetUserSegments(key models.User) ([]models.Segment, error)
	Invalidate() error
}

func NewSegmentation(log *slog.Logger, repo SegmentationRepository, cache SegmentationCache) *Segmentation {
	return &Segmentation{log: log, repo: repo, cache: cache}
}

func (s *Segmentation) CreateSegment(id, description string) (string, error) {
	s.cache.Invalidate()
	s.repo.CreateSegment(id, description)
	panic("service not implemented")
}

func (s *Segmentation) DeleteSegment(id string) (string, error) {
	s.repo.DeleteSegment(id)
	panic("service not implemented")
}

func (s *Segmentation) UpdateSegment(id string, newSegment models.Segment) (string, error) {
	s.repo.UpdateSegment(id, newSegment)
	panic("service not implemented")
}

func (s *Segmentation) GetUserSegments(id string) ([]models.Segment, error) {
	s.repo.GetUserSegments(id)
	panic("service not implemented")
}

func (s *Segmentation) GetSegmentInfo(id string) (models.SegmentInfo, error) {
	s.repo.GetSegmentInfo(id)
	panic("service not implemented")
}

func (s *Segmentation) DistributeSegment(id string, usersPercentage int) (string, error) {
	s.repo.DistributeSegment(id, usersPercentage)
	panic("service not implemented")
}
