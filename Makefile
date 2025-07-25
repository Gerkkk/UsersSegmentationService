PROTO_DIR = protos/proto
OUT_DIR = protos/gen/go

gen:
	mkdir -p $(OUT_DIR)
	protoc -I $(PROTO_DIR) $(PROTO_DIR)/segmentation/segmentationService.proto \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative

run-local:
	go run cmd/Segmentation/main.go --config=./configs/segmentation_config_local.yaml

run-migrator:
	go run cmd/migrator/main.go --config=./configs/segmentation_config_local.yaml

run-infra:
	docker-compose up -d

stop-infra:
	docker-compose down -v

run-docker:
	docker-compose -f docker-compose.infra.yml -f docker-compose.app.yml up --build

run-tests:
	go test ./...