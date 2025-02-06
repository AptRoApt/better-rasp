FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./better-rasp


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/better-rasp .
COPY --from=builder /app/static . 
COPY --from=builder /app/migrations .

EXPOSE 8080

CMD ["./better-rasp"]