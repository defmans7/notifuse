#!/bin/bash

# Deploy PASETO v4 Key Generator to Google Cloud Functions

set -e

# Configuration
FUNCTION_NAME="keygen-function"
REGION="europe-west1"  # Belgium region
PROJECT_ID="notifusev3"
RUNTIME="go123"
ENTRY_POINT="KeygenFunction"

echo "🚀 Deploying PASETO v4 Key Generator to Google Cloud Functions..."

# Set the project
echo "📋 Setting project to $PROJECT_ID..."
gcloud config set project $PROJECT_ID

# Deploy the function (2nd gen for Go 1.23 support)
gcloud functions deploy $FUNCTION_NAME \
  --gen2 \
  --runtime $RUNTIME \
  --trigger-http \
  --entry-point $ENTRY_POINT \
  --allow-unauthenticated \
  --source . \
  --region $REGION \
  --project $PROJECT_ID \
  --memory 128Mi \
  --timeout 60s

echo "✅ Deployment complete!"
echo ""
echo "🌐 Function URL:"
gcloud functions describe $FUNCTION_NAME --region $REGION --gen2 --format="value(serviceConfig.uri)"
echo ""
echo "📝 To test the function:"
echo "  curl \$(gcloud functions describe $FUNCTION_NAME --region $REGION --gen2 --format=\"value(serviceConfig.uri)\")"
