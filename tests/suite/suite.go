package suite

import (
	"context"
	"main/internal/config"
	"net"
	"os"
	"strconv"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	segv1 "main/protos/gen/go/segmentation"
)

type Suite struct {
	*testing.T                          // Потребуется для вызова методов *testing.T внутри Suite
	Cfg        *config.Config           // Конфигурация приложения
	AuthClient segv1.SegmentationClient // Клиент для взаимодействия с gRPC-сервером
}

const (
	grpcHost = "localhost"
)

// New создает новую testsuite
//
// TODO: for pipeline tests we need to wait for app is ready
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath(configPath(), envPath())

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.Grpc.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: segv1.NewSegmentationClient(cc),
	}
}

func configPath() string {
	const key = "CONFIG_PATH"

	if v := os.Getenv(key); v != "" {
		return v
	}

	return "../configs/local_tests.yaml"
}

func envPath() string {
	const key = "ENV_PATH"

	if v := os.Getenv(key); v != "" {
		return v
	}

	return "../.env"
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.Grpc.Port))
}
