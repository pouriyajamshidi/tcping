#!/usr/bin/env bash
#
# build-linux-repo.sh - turn the release packages into an apt, dnf and apk
# repository that users add once and then install and upgrade from, the way
# brew and winget work on the other platforms.
#
# The result is a plain directory tree that GitHub Pages serves out of the
# pouriyajamshidi/packages repository. That repository is shared: it can hold
# more than one project, so this only replaces the packages of the project it
# is given and rebuilds the indexes over everything that is there. Pass it the
# tree as it stands today and it hands back the tree as it should be published.
#
# Only one version of a project is served at a time. Older ones stay
# downloadable on the GitHub release page.
#
# Only docker is needed to run this. Building the indexes needs tools that
# belong to the distributions themselves, so each one runs in a container of
# its own rather than being installed on the machine.

set -euo pipefail

usage() {
	cat <<'EOF'
Usage: build-linux-repo.sh <package-dir> <repo-dir>

  package-dir   holds the <project>-<arch>.deb, .rpm and .apk files of one
                release
  repo-dir      the repository tree to update. It may already hold packages,
                including other projects, and it may not exist yet

Environment:
  GPG_KEY       path to the armored gpg private key that signs the apt and dnf
                metadata and the rpm packages, required. It must have no
                passphrase, since nothing can type one during a release
  APK_KEY       path to the RSA private key that signs the apk index, required
  PROJECT       the project being published, which is also the prefix of its
                package files. Defaults to tcping
  REPO_NAME     names the things the whole repository shares: the keys, the
                .repo file for dnf and the repository id. Defaults to
                pouriyajamshidi
  REPO_URL      where the tree is going to be served from. Defaults to
                https://pouriyajamshidi.github.io/packages
EOF
}

if [ $# -ne 2 ]; then
	usage >&2
	exit 1
fi

PACKAGE_DIR=$(realpath "$1")
REPO_DIR=$(realpath -m "$2")
PROJECT=${PROJECT:-tcping}
REPO_NAME=${REPO_NAME:-pouriyajamshidi}
REPO_URL=${REPO_URL:-https://pouriyajamshidi.github.io/packages}

: "${GPG_KEY:?set GPG_KEY to the armored gpg private key}"
: "${APK_KEY:?set APK_KEY to the alpine signing key}"
GPG_KEY=$(realpath "$GPG_KEY")
APK_KEY=$(realpath "$APK_KEY")

echo "[+] Updating the repository in $REPO_DIR"

# The containers below write as root. They hand the tree back when they are
# done, but a run that died halfway leaves files this user cannot delete.
if [ -d "$REPO_DIR" ] && [ "$(id -u)" != 0 ]; then
	docker run --rm --volume "$REPO_DIR:/repo" alpine:3.22 \
		chown -R "$(id -u):$(id -g)" /repo
fi

mkdir -p "$REPO_DIR"/{deb,rpm,apk/x86_64,apk/aarch64,keys}

# Out with this project's old packages, in with the new ones. Anything another
# project put here is left where it is.
rm -f "$REPO_DIR/deb/$PROJECT-"*.deb \
	"$REPO_DIR/rpm/$PROJECT-"*.rpm \
	"$REPO_DIR"/apk/*/"$PROJECT-"*.apk

cp "$PACKAGE_DIR/$PROJECT-"*.deb "$REPO_DIR/deb/"
cp "$PACKAGE_DIR/$PROJECT-"*.rpm "$REPO_DIR/rpm/"
# Alpine spells the architectures differently than the package file names do,
# and it wants one directory per architecture.
cp "$PACKAGE_DIR/$PROJECT-amd64.apk" "$REPO_DIR/apk/x86_64/"
cp "$PACKAGE_DIR/$PROJECT-arm64.apk" "$REPO_DIR/apk/aarch64/"

# ==================================================
# Debian and Fedora metadata
# ==================================================

# Both are built in one Debian container: it has apt-ftparchive, createrepo_c
# and rpmsign, and the key only has to be imported once.
echo "[+] Building the apt and dnf repositories"
docker run --rm --interactive \
	--volume "$REPO_DIR:/repo" \
	--volume "$GPG_KEY:/key/gpg.asc:ro" \
	--env "PROJECT=$PROJECT" \
	--env "REPO_NAME=$REPO_NAME" \
	--env "REPO_URL=$REPO_URL" \
	--env "OWNER=$(id -u):$(id -g)" \
	debian:stable bash -s <<'EOF'
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends apt-utils createrepo-c rpm gnupg >/dev/null

export GNUPGHOME=/gnupg
mkdir -m 700 "$GNUPGHOME"
gpg --batch --quiet --import /key/gpg.asc
key_id=$(gpg --list-secret-keys --with-colons | awk -F: '/^fpr:/ {print $10; exit}')

# The public halves the two package managers need. apt wants the binary form
# for signed-by, dnf wants the armored one for gpgkey.
gpg --export "$key_id" >"/repo/keys/$REPO_NAME.gpg"
gpg --armor --export "$key_id" >"/repo/keys/$REPO_NAME.asc"

cd /repo/deb
# A flat repository: one directory holding the packages and their index, which
# is all this needs. Every architecture and project shares the index and apt
# picks what it can use.
apt-ftparchive packages . >Packages
gzip -9 --keep --force Packages
apt-ftparchive \
	-o "APT::FTPArchive::Release::Origin=$REPO_NAME" \
	-o "APT::FTPArchive::Release::Label=$REPO_NAME" \
	-o APT::FTPArchive::Release::Architectures="amd64 arm64" \
	release . >Release
# InRelease is the signed Release that apt prefers, Release.gpg is the detached
# signature that older apt versions look for.
gpg --batch --yes --local-user "$key_id" --clearsign --output InRelease Release
gpg --batch --yes --local-user "$key_id" --detach-sign --armor --output Release.gpg Release

cd /repo/rpm
# Only this project's packages: another project's are signed by whoever
# published them.
rpmsign --define "_gpg_name $key_id" --addsign ./"$PROJECT"-*.rpm >/dev/null
createrepo_c --quiet .
gpg --batch --yes --local-user "$key_id" --detach-sign --armor \
	--output repodata/repomd.xml.asc repodata/repomd.xml

# Saves the user from writing this by hand, since dnf has no command that adds
# a repository from a URL and a key.
cat >"/repo/rpm/$REPO_NAME.repo" <<REPO
[$REPO_NAME]
name=$REPO_NAME
baseurl=$REPO_URL/rpm
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=$REPO_URL/keys/$REPO_NAME.asc
REPO

chown -R "$OWNER" /repo
EOF

# ==================================================
# Alpine metadata
# ==================================================

echo "[+] Building the apk repository"
docker run --rm --interactive \
	--volume "$REPO_DIR:/repo" \
	--volume "$APK_KEY:/key/$(basename "$APK_KEY"):ro" \
	--env "KEY=/key/$(basename "$APK_KEY")" \
	--env "KEY_NAME=$REPO_NAME.rsa.pub" \
	--env "OWNER=$(id -u):$(id -g)" \
	alpine:3.22 sh -s <<'EOF'
set -eu
apk add --no-cache abuild openssl >/dev/null

# The public half apk verifies the index with. It comes out of the private key,
# so there is nothing extra to keep around.
openssl rsa -in "$KEY" -pubout -out "/repo/keys/$KEY_NAME" 2>/dev/null

for dir in /repo/apk/*/; do
	cd "$dir"

	# apk downloads a package under the name and version the index gives it,
	# so the file has to be called that. The release names it after the
	# architecture instead, and the version in it is not spelled the way the
	# tag is, so it is read back out of the package.
	for pkg in ./*.apk; do
		name=$(tar -xzOf "$pkg" .PKGINFO | awk -F' = ' '/^pkgname/ {print $2}')
		version=$(tar -xzOf "$pkg" .PKGINFO | awk -F' = ' '/^pkgver/ {print $2}')
		[ "$pkg" = "./$name-$version.apk" ] || mv "$pkg" "$name-$version.apk"
	done

	# The packages are not signed themselves. The index carries their
	# checksums and the index is signed, which is what apk checks.
	apk index --output APKINDEX.tar.gz --allow-untrusted ./*.apk >/dev/null
	abuild-sign --private "$KEY" --public "$KEY_NAME" APKINDEX.tar.gz >/dev/null
done

chown -R "$OWNER" /repo
EOF

# ==================================================
# The page itself
# ==================================================

# GitHub Pages runs Jekyll unless it is told not to, and Jekyll drops files it
# does not recognize.
touch "$REPO_DIR/.nojekyll"

# The repository is public, so whoever lands on it gets the instructions
# instead of a directory listing. The list of packages comes from the index, so
# it stays right as projects are added.
held=$(awk '/^Package: / {print "- `" $2 "`"}' "$REPO_DIR/deb/Packages" | sort --unique)

cat >"$REPO_DIR/README.md" <<EOF
# Packages

An apt, dnf and apk repository, served at <$REPO_URL>.

It currently holds:

$held

Generated by the release workflows of the projects in it. Do not edit by hand.
Only the newest release of each project is served, and older ones stay on that
project's release page.

## Debian, Ubuntu and derivatives

\`\`\`bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL $REPO_URL/keys/$REPO_NAME.gpg -o /etc/apt/keyrings/$REPO_NAME.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/$REPO_NAME.gpg] $REPO_URL/deb ./" | sudo tee /etc/apt/sources.list.d/$REPO_NAME.list &&
  sudo apt update
\`\`\`

## Fedora, RHEL and derivatives

\`\`\`bash
sudo curl -fsSL $REPO_URL/rpm/$REPO_NAME.repo -o /etc/yum.repos.d/$REPO_NAME.repo
\`\`\`

## Alpine

\`\`\`bash
sudo curl -fsSL $REPO_URL/keys/$REPO_NAME.rsa.pub -o /etc/apk/keys/$REPO_NAME.rsa.pub &&
  echo "$REPO_URL/apk" | sudo tee -a /etc/apk/repositories &&
  sudo apk update
\`\`\`

Then install any package listed above the usual way, for example
\`sudo apt install tcping\`, \`sudo dnf install tcping\` or \`sudo apk add tcping\`.
EOF

echo "[+] Repository updated"
