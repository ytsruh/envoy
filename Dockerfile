FROM golang:1.24.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate ./server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o envoy-server ./cmd/server.go

FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/envoy-server .

EXPOSE 8080
ENTRYPOINT ["./envoy-server"]
