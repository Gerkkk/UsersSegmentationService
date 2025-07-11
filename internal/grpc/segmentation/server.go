package segmentationrpc

import (
	"context"
	"google.golang.org/grpc"
	"main/internal/domain/models"
	segv1 "main/protos/gen/go/segmentation"
)

type ServerApi struct {
	segv1.UnimplementedSegmentationServer
	segServ Segmentation
}

type Segmentation interface {
	CreateSegment(id, description string) (string, error)
	DeleteSegment(id string) (string, error)
	UpdateSegment(id string, newSegment models.Segment) (string, error)
	GetUserSegments(id string) ([]models.Segment, error)
	GetSegmentInfo(id string) (models.SegmentInfo, error)
	DistributeSegment(id string, usersPercentage int) (string, error)
}

func Register(gRPC *grpc.Server, segmentation Segmentation) {
	segv1.RegisterSegmentationServer(gRPC, &ServerApi{segServ: segmentation})
}

func (s *ServerApi) CreateSegment(ctx context.Context, req *segv1.CreateSegmentRequest) (*segv1.CreateSegmentResponse, error) {
	_, err := s.segServ.CreateSegment("1", "kek")
	if err != nil {
		return nil, err
	}
	panic("implement me")
}

func (s *ServerApi) DeleteSegment(ctx context.Context, req *segv1.DeleteSegmentRequest) (*segv1.DeleteSegmentResponse, error) {
	_, err := s.segServ.DeleteSegment("1")
	if err != nil {
		return nil, err
	}
	panic("implement me")
}

func (s *ServerApi) UpdateSegment(ctx context.Context, req *segv1.UpdateSegmentRequest) (*segv1.UpdateSegmentResponse, error) {
	_, err := s.segServ.UpdateSegment("1", models.Segment{Id: "1", Description: ""})
	if err != nil {
		return nil, err
	}
	panic("implement me")
}

func (s *ServerApi) GetUserSegments(ctx context.Context, req *segv1.GetUserSegmentsRequest) (*segv1.GetUserSegmentsResponse, error) {
	_, err := s.segServ.GetUserSegments("1")
	if err != nil {
		return nil, err
	}
	panic("implement me")
}

func (s *ServerApi) GetSegmentInfo(ctx context.Context, req *segv1.GetSegmentInfoRequest) (*segv1.GetSegmentInfoResponse, error) {
	_, err := s.segServ.GetSegmentInfo("1")
	if err != nil {
		return nil, err
	}
	panic("implement me")
}

func (s *ServerApi) DistributeSegment(ctx context.Context, req *segv1.DistributeSegmentRequest) (*segv1.DistributeSegmentResponse, error) {
	_, err := s.segServ.DistributeSegment("1", 10)
	if err != nil {
		return nil, err
	}
	panic("implement me")
}
