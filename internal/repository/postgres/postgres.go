package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"main/internal/domain/models"
)

type SegmentationStorage struct {
	shardsNum int
	dbShards  map[int]*sql.DB
}

func NewSegmentationStorage(numShards int, dsns []string) (*SegmentationStorage, error) {
	dbShards := make(map[int]*sql.DB)

	for i, dsn := range dsns {
		db, err := sql.Open("postgres", dsn)

		if err != nil {
			retError := fmt.Errorf("error connecting to database with DSN %s: %w", dsn, err)
			return nil, retError
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		if err := db.Ping(); err != nil {
			retError := fmt.Errorf("error pinging database with DSN %s: %w", dsn, err)
			return nil, retError
		}

		dbShards[i] = db
	}

	segStorage := &SegmentationStorage{shardsNum: numShards, dbShards: dbShards}

	return segStorage, nil
}

func (s *SegmentationStorage) CreateSegment(id, description string) (string, error) {
	panic("db not implemented yet")
}

func (s *SegmentationStorage) DeleteSegment(id string) (string, error) {
	panic("db not implemented yet")
}

func (s *SegmentationStorage) UpdateSegment(id string, newSegment models.Segment) (string, error) {
	panic("db not implemented yet")
}

func (s *SegmentationStorage) GetUserSegments(id string) ([]models.Segment, error) {
	panic("db not implemented yet")
}

func (s *SegmentationStorage) GetSegmentInfo(id string) (models.SegmentInfo, error) {
	panic("db not implemented yet")
}

func (s *SegmentationStorage) DistributeSegment(id string, usersPercentage int) (string, error) {
	panic("db not implemented yet")
}
