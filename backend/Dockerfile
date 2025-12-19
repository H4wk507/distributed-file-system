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

RUN echo "appuser:x:10001:10001::/nonexistent:/sbin/nologin" > /etc/passwd.scratch && \
    echo "appgroup:x:10001:" > /etc/group.scratch

FROM scratch AS api

WORKDIR /app

COPY --from=builder /etc/passwd.scratch /etc/passwd
COPY --from=builder /etc/group.scratch /etc/group

COPY --from=builder /api /api

USER 10001:10001

EXPOSE 8080

ENTRYPOINT ["/api"]

FROM alpine:3.19 AS node

RUN addgroup -g 10001 appgroup && \
    adduser -u 10001 -G appgroup -D -h /app -s /sbin/nologin appuser

RUN apk add --no-cache netcat-openbsd

WORKDIR /app

COPY --from=builder /node /node

RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser:appgroup

EXPOSE 9000

ENTRYPOINT ["/node"]
