set -Eeuo pipefail
log() { echo "==> $*" 1>&2; }
die() { log "FATAL: $*"; exit 1; }
