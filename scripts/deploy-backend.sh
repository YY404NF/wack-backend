#!/usr/bin/env bash
set -eu

ARTIFACT_DIR="${1:?artifact dir is required}"
DEPLOY_PATH="${2:?deploy path is required}"
SERVICE_NAME="${3:?service name is required}"
SERVICE_USER="${4:?service user is required}"
APP_PORT="${5:?app port is required}"
DB_PATH="${6:?db path is required}"
DATA_DIR="${7:?data dir is required}"
JWT_SECRET="${8:?jwt secret is required}"

SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
ENV_DIR="/etc/wack"
ENV_FILE="${ENV_DIR}/${SERVICE_NAME}.env"
RELEASES_DIR="${DEPLOY_PATH}/releases"
CURRENT_LINK="${DEPLOY_PATH}/current"
TIMESTAMP="$(date +%Y%m%d%H%M%S)"
NEW_RELEASE_DIR="${RELEASES_DIR}/${TIMESTAMP}"

if [ -z "${DEPLOY_PATH}" ] || [ "${DEPLOY_PATH}" = "/" ]; then
  echo "refuse to deploy to an empty or root deploy path" >&2
  exit 1
fi

if [ -z "${DATA_DIR}" ] || [ "${DATA_DIR}" = "/" ]; then
  echo "refuse to use an empty or root data dir" >&2
  exit 1
fi

run_root() {
  if [ "$(id -u)" -eq 0 ]; then
    "$@"
  else
    sudo "$@"
  fi
}

run_root mkdir -p "${RELEASES_DIR}" "${DATA_DIR}" "${ENV_DIR}" "${NEW_RELEASE_DIR}"
run_root install -m 0755 "${ARTIFACT_DIR}/wack-backend" "${NEW_RELEASE_DIR}/wack-backend"

sed \
  -e "s|__SERVICE_NAME__|${SERVICE_NAME}|g" \
  -e "s|__SERVICE_USER__|${SERVICE_USER}|g" \
  -e "s|__WORKING_DIRECTORY__|${CURRENT_LINK}|g" \
  -e "s|__EXEC_START__|${CURRENT_LINK}/wack-backend|g" \
  -e "s|__ENV_FILE__|${ENV_FILE}|g" \
  "${ARTIFACT_DIR}/wack-backend.service" | run_root tee "${SERVICE_FILE}" >/dev/null

cat <<EOF | run_root tee "${ENV_FILE}" >/dev/null
WACK_PORT=${APP_PORT}
WACK_DATA_DIR=${DATA_DIR}
WACK_DB_PATH=${DB_PATH}
WACK_JWT_SECRET=${JWT_SECRET}
EOF

run_root ln -sfn "${NEW_RELEASE_DIR}" "${CURRENT_LINK}"
run_root systemctl daemon-reload
run_root systemctl enable "${SERVICE_NAME}"
run_root systemctl restart "${SERVICE_NAME}"
run_root systemctl status "${SERVICE_NAME}" --no-pager
