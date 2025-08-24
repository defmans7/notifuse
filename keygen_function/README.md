# PASETO v4 Key Generator - Google Cloud Function

This Google Cloud Function serves a web page that generates PASETO v4 asymmetric key pairs. It provides both a user-friendly web interface and a REST API for key generation.

## Features

- **Web Interface**: Beautiful, responsive HTML page for generating keys
- **REST API**: POST endpoint for programmatic key generation
- **Secure**: Uses the same PASETO v4 library as the main application
- **Copy to Clipboard**: Easy copying of generated keys
- **Security Warnings**: Clear warnings about private key handling

## Local Development

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Run locally using the Functions Framework:

   ```bash
   go run function.go
   ```

   Or install the Functions Framework CLI:

   ```bash
   go install github.com/GoogleCloudPlatform/functions-framework-go/funcframework/cmd/function@latest
   function --target KeygenFunction --source .
   ```

3. Open http://localhost:8080 in your browser

## Deployment

Deploy to Google Cloud Functions:

```bash
gcloud functions deploy keygen-function \
  --runtime go121 \
  --trigger-http \
  --entry-point KeygenFunction \
  --allow-unauthenticated \
  --source .
```

## API Usage

### GET /

Returns the HTML interface for key generation.

### POST /

Generates a new PASETO v4 key pair.

**Response:**

```json
{
  "privateKey": "base64-encoded-private-key",
  "publicKey": "base64-encoded-public-key"
}
```

**Example:**

```bash
curl -X POST https://your-function-url/
```

## Security Considerations

- **Private Key Security**: The private key should be kept secret and stored securely
- **HTTPS Only**: Always use HTTPS in production
- **Access Control**: Consider adding authentication for production use
- **Key Storage**: Never store private keys in client-side code or public repositories

## Dependencies

- `aidanwoods.dev/go-paseto` - PASETO implementation
- `github.com/GoogleCloudPlatform/functions-framework-go` - Google Cloud Functions framework
