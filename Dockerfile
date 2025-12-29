FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /manager ./cmd/manager

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates wget && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=builder /manager /manager
RUN chmod +x /manager
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/manager"]

