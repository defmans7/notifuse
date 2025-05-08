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

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/app/server"] 