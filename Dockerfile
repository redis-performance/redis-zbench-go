FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG GIT_SHA
ARG GIT_DIRTY
RUN GIT_SHA=${GIT_SHA:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")} && \
    GIT_DIRTY=${GIT_DIRTY:-$(git diff --no-ext-diff 2>/dev/null | wc -l || echo "0")} && \
    CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
        -ldflags="-w -s -X 'main.GitSHA1=${GIT_SHA}' -X 'main.GitDirty=${GIT_DIRTY}'" \
        -o redis-zbench-go .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/redis-zbench-go .

RUN chown appuser:appgroup /app/redis-zbench-go && \
    chmod +x /app/redis-zbench-go

USER appuser

ENTRYPOINT ["/app/redis-zbench-go"]
CMD ["--help"]
