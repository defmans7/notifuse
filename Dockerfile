# Stage 1: Build the React frontend
FROM node:20-alpine AS frontend-builder

# Set working directory for the frontend
WORKDIR /build/console

# Copy frontend package files
COPY console/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source code
COPY console/ ./

# Build frontend in production mode
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1.23-alpine AS backend-builder

# Set working directory
WORKDIR /build

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY cmd/ cmd/
COPY config/ config/
COPY internal/ internal/
COPY pkg/ pkg/

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /tmp/server ./cmd/api

# Stage 3: Create the runtime container
FROM alpine:latest

# Add necessary runtime packages
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Create application directory structure
WORKDIR /app
RUN mkdir -p /app/console/dist /app/data

# Copy the binary from the builder stage
COPY --from=backend-builder /tmp/server /app/server

# Copy the built console files
COPY --from=frontend-builder /build/console/dist/ /app/console/dist/

# Environment variables with defaults that can be overridden
ENV SERVER_PORT=8080 \
    SERVER_HOST=0.0.0.0 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_PREFIX=notifuse \
    DB_NAME=notifuse_system \
    ENVIRONMENT=production \
    TRACING_ENABLED=false \
    TRACING_SERVICE_NAME=notifuse-api \
    TRACING_SAMPLING_PROBABILITY=0.1 \
    TRACING_TRACE_EXPORTER=none \
    TRACING_METRICS_EXPORTER=none

# The following environment variables need to be provided at runtime:
# - PASETO_PRIVATE_KEY (Required)
# - PASETO_PUBLIC_KEY (Required)
# - SECRET_KEY (Required)
# - ROOT_EMAIL
# - API_ENDPOINT
# - WEBHOOK_ENDPOINT

# Expose the application port
EXPOSE ${SERVER_PORT}

# Run the application
CMD ["/app/server"] 