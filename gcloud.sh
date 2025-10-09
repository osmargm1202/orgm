#!/bin/bash

# Script interactivo para crear una cuenta de servicio y descargar su JSON

read -p "Introduce el nombre de la cuenta de servicio (ej: cloud-run-invoker): " ACCOUNT_NAME

PROJECT_ID="orgm-797f1"
DISPLAY_NAME="Cuenta para Cloud Run"
EMAIL="$ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com"
JSON_FILE="${ACCOUNT_NAME}.json"

# 1. Crear la cuenta de servicio
# gcloud iam service-accounts create "$ACCOUNT_NAME" \
#     --display-name="$DISPLAY_NAME" \
#     --project="$PROJECT_ID"

# 2. Dar permisos para invocar todos los servicios de Cloud Run
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:$EMAIL" \
    --role="roles/run.developer"

# 3. Generar y descargar el archivo JSON de credenciales
gcloud iam service-accounts keys create "$JSON_FILE" \
    --iam-account="$EMAIL" \
    --project="$PROJECT_ID"

echo "Listo. Archivo de credenciales: $JSON_FILE"