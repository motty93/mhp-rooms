# Build stage
FROM golang:1.24-alpine AS builder

ARG GO_VERSION=1.24

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -a -installsuffix cgo \
    -o room-cleanup ./cmd/room-cleanup

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /build/room-cleanup .

RUN chown -R appuser:appgroup /app

USER appuser

# 非アクティブ判定のしきい値（時間）。Cloud Run Job の環境変数で上書き可能
ENV ROOM_INACTIVE_HOURS=48

CMD ["./room-cleanup"]
