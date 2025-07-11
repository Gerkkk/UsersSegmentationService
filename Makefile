PROTO_DIR = protos/proto
OUT_DIR = protos/gen/go

gen:
	mkdir -p $(OUT_DIR)
	protoc -I $(PROTO_DIR) $(PROTO_DIR)/segmentation/segmentationService.proto \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative