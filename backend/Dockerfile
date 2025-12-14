FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /node ./cmd/node

FROM alpine:3.19 AS api

WORKDIR /app

COPY --from=builder /api /app/api

EXPOSE 8080

CMD ["/app/api"]

FROM alpine:3.19 AS node

RUN apk add --no-cache netcat-openbsd

WORKDIR /app

COPY --from=builder /node /app/node

EXPOSE 9000

ENTRYPOINT ["/app/node"]
