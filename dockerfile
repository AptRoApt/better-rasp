FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./better-rasp ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/better-rasp .

COPY --from=builder /app/web/static /app/static
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/web/index.html /app/index.html

EXPOSE 8080

CMD ["./better-rasp"]