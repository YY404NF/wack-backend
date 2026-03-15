# wack-backend

## GitHub Secrets

在仓库 `Settings -> Secrets and variables -> Actions -> Repository secrets` 中添加以下字段。

### 必需字段

| Secret | 示例值 |
| --- | --- |
| `DEPLOY_HOST` | `8.159.159.150` |
| `DEPLOY_PORT` | `22` |
| `DEPLOY_USER` | `root` |
| `DEPLOY_SSH_KEY` | `-----BEGIN OPENSSH PRIVATE KEY----- ...` |
| `BACKEND_DEPLOY_PATH` | `/srv/wack-backend` |
| `BACKEND_SERVICE_NAME` | `wack-backend` |
| `BACKEND_SERVICE_USER` | `root` |
| `BACKEND_PORT` | `8080` |
| `WACK_DB_PATH` | `/root/wack_db/wack.db` |
| `WACK_JWT_SECRET` | `replace-with-a-random-secret` |
| `WACK_CORS_ALLOW_ORIGIN` | `http://8.159.159.150` |
