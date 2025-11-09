# Build stage
FROM golang:1.24-alpine AS builder

ARG GO_VERSION=1.24

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build optimized binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s' \
    -a -installsuffix cgo \
    -o ogp-renderer ./cmd/ogp-renderer

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary
COPY --from=builder /build/ogp-renderer .

# Copy font assets
COPY --from=builder /build/cmd/ogp-renderer/assets ./assets

# Copy internal modules (for palette)
COPY --from=builder /build/internal ./internal

RUN chown -R appuser:appgroup /app

USER appuser

# 環境変数のデフォルト値を設定
ENV FONT_PATH=/app/assets/fonts/NotoSansCJKjp-Bold.otf \
    ICON_IMAGE_PATH=/app/assets/images/icon.webp

CMD ["./ogp-renderer"]
