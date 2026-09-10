FROM golang:1.26 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o stream-processor ./services/stream-processor

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /build/stream-processor .
CMD ["./stream-processor"]
