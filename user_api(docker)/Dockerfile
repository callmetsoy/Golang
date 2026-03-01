FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY .env .
EXPOSE 8080
ENTRYPOINT ["./main"]