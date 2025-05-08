# Docker Deployment for Notifuse

This document provides instructions for deploying the Notifuse application using Docker.

## Prerequisites

- Docker and Docker Compose installed
- Generate PASETO keys for authentication (see below)
- PostgreSQL database (included in docker-compose)

## Getting Started

### 1. Generate PASETO Keys

Before running the application, you need to generate PASETO keys for authentication. The application provides a keygen utility:

```bash
go run cmd/keygen/main.go
```

This will output base64-encoded private and public keys that you'll need to set in your environment variables.

### 2. Set Required Environment Variables

Create a `.env` file in the project root with the following required variables:

```
PASETO_PRIVATE_KEY=your_base64_encoded_private_key_from_keygen
PASETO_PUBLIC_KEY=your_base64_encoded_public_key_from_keygen
SECRET_KEY=your_secret_key_for_workspace_encryption
ROOT_EMAIL=admin@example.com
API_ENDPOINT=http://localhost:8080
```

### 3. Build and Run with Docker Compose

```bash
docker-compose up -d
```

This will:

- Build the React console frontend in production mode
- Build the Go API backend
- Start a PostgreSQL database

The application will be accessible at http://localhost:8080

### 4. Logs and Troubleshooting

View logs:

```bash
docker-compose logs -f api
```

### Environment Variables

#### Required Variables

- `PASETO_PRIVATE_KEY` - Base64 encoded private key
- `PASETO_PUBLIC_KEY` - Base64 encoded public key
- `SECRET_KEY` - Secret key for workspace settings encryption

#### Optional Variables with Defaults

```
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_PREFIX=notifuse
DB_NAME=notifuse_system
ENVIRONMENT=production
TRACING_ENABLED=false
```

See the docker-compose.yml file for additional configuration options.

## Advanced Configuration

### Database

The PostgreSQL database is configured with the following defaults:

- User: postgres
- Password: postgres
- Database: postgres
- Port: 5432

Data is persisted in a Docker volume named `postgres-data`.

### Frontend Console

The React frontend is built during the Docker image creation process:

1. Node.js is used to build the frontend in production mode
2. The built static files are placed in the `/app/console/dist` directory in the final image
3. This location matches exactly what's expected by the Go backend (`console/dist` path in ConsoleHandler)
4. The Go backend serves these static files when accessing the root URL

## Production Considerations

For production deployments:

1. Use a proper secrets management solution
2. Configure SSL for secure connections
3. Use dedicated database credentials
4. Consider using managed PostgreSQL services
5. Set `ENVIRONMENT=production`
