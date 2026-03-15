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

## Workflow 行为

后端仓库的 GitHub Actions 在 `push main` 后会执行以下步骤：

1. 运行 `go test ./...`
2. 构建 Linux amd64 后端二进制
3. 上传部署压缩包到服务器
4. 写入 `/etc/wack/<service>.env`
5. 安装或更新 systemd 服务文件
6. 切换 `current` 软链接到新 release
7. 重启后端服务
