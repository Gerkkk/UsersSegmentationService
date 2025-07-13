package apperrors

import (
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSegmentAlreadyExists = errors.New("segment already exists")
	ErrSegmentNotFound      = errors.New("segment not found")
	ErrUserExists           = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrShardUnavailable     = errors.New("shard unavailable")
	ErrSegmentDistributed   = errors.New("segment distributed")

	ErrShardRollbackFailed = errors.New("shard rollback failed")
	ErrShardCommitFailed   = errors.New("shard commit failed")
	ErrShardPrepareFailed  = errors.New("shard prepare failed")
)

var errorToCode = map[error]codes.Code{
	ErrSegmentAlreadyExists: codes.AlreadyExists,
	ErrShardUnavailable:     codes.Unavailable,
	ErrUserExists:           codes.AlreadyExists,
	ErrUserNotFound:         codes.NotFound,
	ErrSegmentNotFound:      codes.NotFound,
	ErrSegmentDistributed:   codes.AlreadyExists,
}

func isPublic(err error) bool {
	_, ok := errorToCode[err]
	return ok
}

func Convert(log *slog.Logger, err error) error {
	if err == nil {
		return nil
	}

	log.Error("ERROR: %v", err)

	if isPublic(err) {
		return status.Error(errorToCode[err], err.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
