FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/http_api ./src/cmd/http_api
RUN CGO_ENABLED=0 go build -o /app/migrator ./src/cmd/migrator

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates netcat-openbsd

COPY --from=builder /app/http_api /app/http_api
COPY --from=builder /app/migrator /app/migrator
COPY src/internal/infrastructure/data/migrations /app/src/internal/infrastructure/data/migrations

COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
