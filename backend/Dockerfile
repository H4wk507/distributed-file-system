FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=.,target=/app \
    CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api & \
    CGO_ENABLED=0 GOOS=linux go build -o /node ./cmd/node & \
    wait

FROM scratch AS api

WORKDIR /app

COPY --from=builder /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]

FROM alpine:3.19 AS node

RUN apk add --no-cache netcat-openbsd

WORKDIR /app

COPY --from=builder /node /node

EXPOSE 9000

ENTRYPOINT ["/node"]
