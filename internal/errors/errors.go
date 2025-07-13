package apperrors

import (
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Публичные ошибки, которые будут показываться пользователю в ответе сервера
var (
	ErrSegmentAlreadyExists = errors.New("segment already exists")
	ErrSegmentNotFound      = errors.New("segment not found")
	ErrUserExists           = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrShardUnavailable     = errors.New("shard unavailable")
	ErrSegmentDistributed   = errors.New("segment distributed")
)

// errorToCode - Отображение публичных ошибок в коды ответа Grpc
var errorToCode = map[error]codes.Code{
	ErrSegmentAlreadyExists: codes.AlreadyExists,
	ErrShardUnavailable:     codes.Unavailable,
	ErrUserExists:           codes.AlreadyExists,
	ErrUserNotFound:         codes.NotFound,
	ErrSegmentNotFound:      codes.NotFound,
	ErrSegmentDistributed:   codes.AlreadyExists,
}

// isPublic - функция, проверяющая, является ли ошибка публичной
func isPublic(err error) bool {
	_, ok := errorToCode[err]
	return ok
}

/*
	Convert - функция, конвертирующая ошибку в ее окончательный вид для пользователя, и логирующая исходный вид

В случае непубличной ошибки вернет codes.Internal с сообщением internal server error
*/
func Convert(log *slog.Logger, err error) error {
	if err == nil {
		return nil
	}

	log.Error("ERROR", slog.String("ERROR", err.Error()))

	if isPublic(err) {
		return status.Error(errorToCode[err], err.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
