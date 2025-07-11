package segmentation

import (
	"context"
	"google.golang.org/grpc"
	segv1 "main/protos/gen/go/segmentation"
)

type ServerApi struct {
	segv1.UnimplementedSegmentationServer
}

func Register(gRPC *grpc.Server) {
	segv1.RegisterSegmentationServer(gRPC, &ServerApi{})
}

func (s *ServerApi) CreateSegment(ctx context.Context, req *segv1.CreateSegmentRequest) (*segv1.CreateSegmentResponse, error) {
	panic("implement me")
}

func (s *ServerApi) DeleteSegment(ctx context.Context, req *segv1.DeleteSegmentRequest) (*segv1.DeleteSegmentResponse, error) {
	panic("implement me")
}

func (s *ServerApi) UpdateSegment(ctx context.Context, req *segv1.UpdateSegmentRequest) (*segv1.UpdateSegmentResponse, error) {
	panic("implement me")
}

func (s *ServerApi) GetUserSegments(ctx context.Context, req *segv1.GetUserSegmentsRequest) (*segv1.GetUserSegmentsResponse, error) {
	panic("implement me")
}

func (s *ServerApi) GetSegmentInfo(ctx context.Context, req *segv1.GetSegmentInfoRequest) (*segv1.GetSegmentInfoResponse, error) {
	panic("implement me")
}

func (s *ServerApi) DistributeSegment(ctx context.Context, req *segv1.DistributeSegmentRequest) (*segv1.DistributeSegmentResponse, error) {
	panic("implement me")
}
