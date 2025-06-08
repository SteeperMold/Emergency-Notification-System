FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags "-s -w" -a -o apiservice ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/apiservice .
COPY .env .
COPY ./database/migrations ./database/migrations

EXPOSE 8080

ENTRYPOINT ["./apiservice"]
