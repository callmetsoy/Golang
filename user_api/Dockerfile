FROM golang:1.25-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./cmd/api/main.go
CMD ["./main"]