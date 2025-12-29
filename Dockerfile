# Build stage
FROM golang:1.24 as builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /manager ./cmd/manager

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /manager /manager

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/manager"]
