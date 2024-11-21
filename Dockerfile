FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

ENV CONFG_PATH=/path/to/config \
    GIN_MODE=release

CMD ["./main"]
