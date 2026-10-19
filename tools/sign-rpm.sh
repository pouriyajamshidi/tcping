#!/usr/bin/env bash
#
# sign-rpm.sh - sign the rpm packages of a release with the key that signs the
# Linux package repository.
#
# The release workflow runs this before it uploads the packages, so the rpm on
# the release page is the exact file the repository describes. That way the
# repository can send dnf to the release page for the package itself, and
# people who download the rpm by hand get a signed one too.
#
# Only docker is needed to run this, like build-linux-repo.sh.

set -euo pipefail

usage() {
	cat <<'EOF'
Usage: sign-rpm.sh <dir>

  dir           holds the .rpm files to sign. They are signed in place

Environment:
  GPG_KEY       path to the armored gpg private key to sign with, required. It
                must have no passphrase, since nothing can type one during a
                release
EOF
}

if [ $# -ne 1 ]; then
	usage >&2
	exit 1
fi

RPM_DIR=$(realpath "$1")

: "${GPG_KEY:?set GPG_KEY to the armored gpg private key}"
GPG_KEY=$(realpath "$GPG_KEY")

echo "[+] Signing the rpm packages in $RPM_DIR"
docker run --rm --interactive \
	--volume "$RPM_DIR:/rpm" \
	--volume "$GPG_KEY:/key/gpg.asc:ro" \
	--env "OWNER=$(id -u):$(id -g)" \
	debian:stable bash -s <<'EOF'
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends rpm gnupg >/dev/null

export GNUPGHOME=/gnupg
mkdir -m 700 "$GNUPGHOME"
gpg --batch --quiet --import /key/gpg.asc
key_id=$(gpg --list-secret-keys --with-colons | awk -F: '/^fpr:/ {print $10; exit}')

cd /rpm
rpmsign --define "_gpg_name $key_id" --addsign ./*.rpm >/dev/null

chown "$OWNER" ./*.rpm
EOF

echo "[+] Packages signed"
