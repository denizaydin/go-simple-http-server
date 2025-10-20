# syntax=docker/dockerfile:1

########################################
# 1️⃣ Build stage
########################################
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Sertifikalar ve git (mod indirimi için)
RUN apk add --no-cache ca-certificates git && update-ca-certificates

# Go mod dosyası (go.sum olmayabilir)
COPY go.mod ./
RUN [ -f go.mod ] && go mod download || true

# Kaynak kodu kopyala
COPY . .

# Statik, küçük binary oluştur
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags "-s -w" -o /src/go-simple-http-server ./go-simple-http-server.go

########################################
# 2️⃣ Runtime stage
########################################
FROM alpine:3.20

# Zaman ve sertifikalar (CA lazım)
RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

WORKDIR /app
COPY --from=builder /src/go-simple-http-server /app/go-simple-http-server

# Non-root kullanıcı oluştur
RUN adduser -D -H app && chown app:app /app/go-simple-http-server
USER app

EXPOSE 8080
ENTRYPOINT ["/app/go-simple-http-server"]