# Backend Deploy

## Required GitHub Secrets

- `DEPLOY_HOST`
- `DEPLOY_PORT`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`
- `BACKEND_DEPLOY_PATH`
- `BACKEND_SERVICE_NAME`
- `BACKEND_SERVICE_USER`
- `BACKEND_PORT`
- `WACK_DB_PATH`
- `WACK_DATA_DIR`
- `WACK_JWT_SECRET`

## Recommended values

- `BACKEND_DEPLOY_PATH=/srv/wack-backend`
- `BACKEND_SERVICE_NAME=wack-backend`
- `BACKEND_SERVICE_USER=root`
- `BACKEND_PORT=8080`
- `WACK_DB_PATH=/root/wack_db/wack.db`
- `WACK_DATA_DIR=/srv/wack-backend/data`

## What the workflow does

1. Runs `go test ./...`
2. Builds a Linux amd64 binary
3. Uploads a deployment bundle to the server
4. Writes a systemd service
5. Writes `/etc/wack/<service>.env`
6. Switches the `current` symlink to the new release
7. Restarts the service
