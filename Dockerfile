FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./src/cmd/server

FROM alpine:3.22

WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /bin/server /app/server
COPY --from=builder /app/src/migrations /app/src/migrations

EXPOSE 8080

USER appuser

CMD ["/app/server"]
