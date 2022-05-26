set -Eeuo pipefail
log() { echo "==> $*" 1>&2; }
err() { log "ERROR: $*"; return 1; }
die() { log "FATAL: $*"; exit 1; }
