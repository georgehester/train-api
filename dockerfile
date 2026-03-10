# Build Image
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src/ ./src/

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./src/

# Run Image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8000

ENTRYPOINT ["./server"]