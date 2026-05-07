FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app-manager ./main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /app-manager /usr/local/bin/app-manager
COPY config.example.yaml /etc/app-manager/config.yaml

EXPOSE 8080

ENTRYPOINT ["app-manager"]
