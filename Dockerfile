FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o url-analyzer ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-analyzer .

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["./url-analyzer"]