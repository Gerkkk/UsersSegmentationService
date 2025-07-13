FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/Segmentation

EXPOSE 9090

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY configs ./configs
CMD ["./server"]
