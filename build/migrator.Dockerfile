FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o migrator ./cmd/migrator

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/migrator .
COPY configs ./configs
COPY internal/migrations ./internal/migrations
CMD ["./migrator"]
