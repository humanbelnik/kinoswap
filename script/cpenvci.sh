#!/bin/bash

# Create .env in CI from Github secrets

set -e

cat << EOF > .env
HTTP_PORT=${{ secrets.HTTP_PORT }}
HTTP_HOST=${{ secrets.HTTP_HOST }}

REDIS_PORT=${{ secrets.REDIS_PORT }}
REDIS_HOST=${{ secrets.REDIS_HOST }}
REDIS_PASSWORD=${{ secrets.REDIS_PASSWORD }}

DB_HOST=${{ secrets.DB_HOST }}
DB_PORT=${{ secrets.DB_PORT }}
DB_USER=${{ secrets.DB_USER }}
DB_PASSWORD=${{ secrets.DB_PASSWORD }}
DB_NAME=${{ secrets.DB_NAME }}
DB_SSLMODE=${{ secrets.DB_SSLMODE }}

AWS_ACCESS_KEY_ID=${{ secrets.AWS_ACCESS_KEY_ID }}
AWS_SECRET_ACCESS_KEY=${{ secrets.AWS_SECRET_ACCESS_KEY }}
AWS_REGION=${{ secrets.AWS_REGION }}
AWS_S3_FORCE_PATH_STYLE=${{ secrets.AWS_S3_FORCE_PATH_STYLE }}
AWS_S3_ENDPOINT=${{ secrets.AWS_S3_ENDPOINT }}

EMBEDDER_PORT=${{ secrets.EMBEDDER_PORT }}
EMBEDDER_HOST=${{ secrets.EMBEDDER_HOST }}

ADMIN_SECRET=${{ secrets.ADMIN_SECRET }}

PGADMIN_DEFAULT_EMAIL=${{ secrets.PGADMIN_DEFAULT_EMAIL }}
PGADMIN_DEFAULT_PASSWORD=${{ secrets.PGADMIN_DEFAULT_PASSWORD }}
EOF

