# 1) Build stage
FROM golang:1.24 AS builder
WORKDIR /app

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o simple-service ./cmd/server

FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=builder /app/simple-service .
EXPOSE 8080

ENTRYPOINT ["/simple-service"]