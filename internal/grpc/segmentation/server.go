package segmentationrpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"main/internal/domain/models"
	segv1 "main/protos/gen/go/segmentation"
	"strconv"
)

// ServerApi - Занимается валидацией входных данных запросов и отправляет ответы пользователю
type ServerApi struct {
	segv1.UnimplementedSegmentationServer
	segServ Segmentation
}

type Segmentation interface {
	CreateSegment(segment models.Segment) (string, error)
	DeleteSegment(id string) (string, error)
	UpdateSegment(id string, newSegment models.Segment) (string, error)
	GetUserSegments(id int) ([]models.Segment, error)
	GetSegmentInfo(id string) (models.SegmentInfo, error)
	DistributeSegment(id string, usersPercentage int) (string, error)
}

func Register(gRPC *grpc.Server, segmentation Segmentation) {
	segv1.RegisterSegmentationServer(gRPC, &ServerApi{segServ: segmentation})
}

func (s *ServerApi) CreateSegment(ctx context.Context, req *segv1.CreateSegmentRequest) (*segv1.CreateSegmentResponse, error) {
	id, err := s.segServ.CreateSegment(models.Segment{Id: req.Id, Description: req.Description})
	return &segv1.CreateSegmentResponse{Id: id}, err
}

func (s *ServerApi) DeleteSegment(ctx context.Context, req *segv1.DeleteSegmentRequest) (*segv1.DeleteSegmentResponse, error) {
	id, err := s.segServ.DeleteSegment(req.Id)
	if err != nil {
		return nil, err
	}

	return &segv1.DeleteSegmentResponse{Id: id}, nil
}

func (s *ServerApi) UpdateSegment(ctx context.Context, req *segv1.UpdateSegmentRequest) (*segv1.UpdateSegmentResponse, error) {
	var newId string
	if req.NewId == nil {
		newId = ""
	} else {
		newId = *req.NewId
	}

	var newDescription string
	if req.NewDescription == nil {
		newDescription = ""
	} else {
		newDescription = *req.NewDescription
	}

	id, err := s.segServ.UpdateSegment(req.Id, models.Segment{Id: newId, Description: newDescription})
	if err != nil {
		return nil, err
	}

	return &segv1.UpdateSegmentResponse{Id: id}, nil
}

func (s *ServerApi) GetUserSegments(ctx context.Context, req *segv1.GetUserSegmentsRequest) (*segv1.GetUserSegmentsResponse, error) {
	segs, err := s.segServ.GetUserSegments(int(req.Id))
	if err != nil {
		return nil, err
	}

	retCategs := make([]*segv1.CategoryInfo, 0)

	for _, seg := range segs {
		newCatInf := &segv1.CategoryInfo{}
		newCatInf.Id = seg.Id
		retCategs = append(retCategs, newCatInf)
	}

	return &segv1.GetUserSegmentsResponse{Categories: retCategs}, nil
}

func (s *ServerApi) GetSegmentInfo(ctx context.Context, req *segv1.GetSegmentInfoRequest) (*segv1.GetSegmentInfoResponse, error) {
	segInf, err := s.segServ.GetSegmentInfo(req.GetId())
	if err != nil {
		return nil, err
	}

	return &segv1.GetSegmentInfoResponse{Id: segInf.Id, Description: segInf.Description, UsersNum: segInf.UsersNum}, nil
}

func (s *ServerApi) DistributeSegment(ctx context.Context, req *segv1.DistributeSegmentRequest) (*segv1.DistributeSegmentResponse, error) {
	i, err := strconv.ParseInt(req.GetUsersPercentage(), 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid users percentage")
	}

	if i <= 0 || i > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid users percentage")
	}

	id, err := s.segServ.DistributeSegment(req.GetId(), int(i))
	if err != nil {
		return nil, err
	}

	return &segv1.DistributeSegmentResponse{Id: id}, nil
}
