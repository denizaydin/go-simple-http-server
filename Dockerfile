FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o go-simple-http-server go-simple-http-server.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-simple-http-server .
EXPOSE 8080
CMD ["./go-simple-http-server"]
