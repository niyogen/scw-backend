#!/bin/bash
# ─────────────────────────────────────────────────────────────────
# Carfenter / Sewa Services Backend — Google Cloud Run Deployment
# ─────────────────────────────────────────────────────────────────
set -euo pipefail

# --- COLOR DEFS ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=================================================================${NC}"
echo -e "${BLUE}🚀 Starting Google Cloud Run Deployment for Carfenter Backend...${NC}"
echo -e "${BLUE}=================================================================${NC}"

# --- CONFIGURATION ---
SERVICE_NAME="carfenter-backend"
REGION="us-central1"
PORT="8080"
GOOGLE_CLIENT_ID="${GOOGLE_CLIENT_ID:?Set GOOGLE_CLIENT_ID in the environment}"

echo -e "\n${BLUE}🔍 Checking Google Cloud authentication...${NC}"
if ! gcloud auth print-access-token >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Not authenticated or session expired.${NC}"
    echo -e "${YELLOW}Please log in using your account (rangap@niyogen.com):${NC}"
    echo -e "    ${GREEN}gcloud auth login rangap@niyogen.com${NC}\n"
    exit 1
fi

ACTIVE_ACCOUNT=$(gcloud config get-value account 2>/dev/null || echo "Unknown")
ACTIVE_PROJECT=$(gcloud config get-value project 2>/dev/null || echo "")

echo -e "${GREEN}✅ Authenticated as: ${ACTIVE_ACCOUNT}${NC}"
if [ -n "$ACTIVE_PROJECT" ]; then
    echo -e "${GREEN}✅ Active Project: ${ACTIVE_PROJECT}${NC}"
fi

echo -e "\n${BLUE}📦 Deploying container to Cloud Run ($SERVICE_NAME in $REGION)...${NC}"

gcloud run deploy "$SERVICE_NAME" \
    --source . \
    --region "$REGION" \
    --port "$PORT" \
    --allow-unauthenticated \
    --set-cloudsql-instances "claw-501704:us-central1:carfenter-db" \
    --update-env-vars "GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID},ENVIRONMENT=production,DB_HOST=/cloudsql/claw-501704:us-central1:carfenter-db,DB_PORT=5432,DB_USER=postgres,DB_NAME=delivery_db,DB_SSLMODE=disable,JWT_EXPIRES_IN_HOURS=72,GCS_BUCKET_NAME=carfenter-media-1003415657356,APP_BASE_URL=https://carfenter-backend-1003415657356.us-central1.run.app,SMTP_HOST=smtp.gmail.com,SMTP_PORT=587,SMTP_USER=itranga@gmail.com,SMTP_FROM_NAME=Sewa Home Services,SMTP_FROM_EMAIL=itranga@gmail.com" \
    --set-secrets "GOOGLE_CLIENT_SECRET=google-client-secret:latest,DB_PASSWORD=db-password:latest,JWT_SECRET=jwt-secret:latest,SMTP_PASSWORD=smtp-password:latest"

echo -e "\n${GREEN}=================================================================${NC}"
echo -e "${GREEN}🎉 Carfenter Backend deployed successfully to Google Cloud Run!${NC}"
echo -e "${GREEN}=================================================================${NC}"
