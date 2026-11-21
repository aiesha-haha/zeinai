# Build Stage (v2 - added students sync logic)
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Set Go Proxy for better resilience
ENV GOPROXY=https://proxy.golang.org,direct

# Install git and essential build tools
RUN apk add --no-cache git

# Copy mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Install Swag for documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest
# Generate Docs
RUN /go/bin/swag init -g cmd/server/main.go --output docs

# Optional tidy (should be tidy if already downloaded, but helps if new imports added)
RUN go mod tidy

# Install Goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0

# Build the main application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Build the cron job
RUN CGO_ENABLED=0 GOOS=linux go build -o cron_reset_tokens ./cmd/cron/reset_tokens.go

# Build the worker
RUN CGO_ENABLED=0 GOOS=linux go build -o kmj_worker ./cmd/worker/main.go

# Final Stage
FROM alpine:3.19

WORKDIR /app

# Install standard certificates
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/cron_reset_tokens .
COPY --from=builder /app/kmj_worker .
COPY --from=builder /app/locales ./locales

COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

# Copy goose binary
COPY --from=builder /go/bin/goose /bin/goose

# Expose port
EXPOSE 8000

# Run migrations and then start app
CMD ["sh", "-c", "goose -dir migrations postgres \"host=$DB_HOST user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSL_MODE\" up && ./main"]
