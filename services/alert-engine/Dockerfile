FROM golang:1.26 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o alert-engine ./services/alert-engine

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /build/alert-engine .
CMD ["./alert-engine"]
