#!/usr/bin/env bash
set -euo pipefail

# Publish ponto-cli to AUR
# Requires: AUR_KEY environment variable

if [ -z "${GITHUB_REF_NAME:-}" ]; then
  echo "ERROR: GITHUB_REF_NAME is not set (must run from GitHub Actions release workflow)"
  exit 1
fi
VERSION="${GITHUB_REF_NAME#v}"
REPO="alextakitani/ponto-cli"

echo "Publishing ponto-cli $VERSION to AUR..."

# Get source tarball checksum
SOURCE_URL="https://github.com/$REPO/archive/v${VERSION}.tar.gz"
curl -fsSL "$SOURCE_URL" -o source.tar.gz
SHA256=$(sha256sum source.tar.gz | cut -d' ' -f1)
rm source.tar.gz

# Generate PKGBUILD
cat > PKGBUILD << EOF
# Maintainer: Alex Takitani
pkgname=ponto-cli
pkgver=$VERSION
pkgrel=1
pkgdesc="CLI for Ponto"
arch=('x86_64' 'aarch64')
url="https://github.com/$REPO"
license=('MIT')
depends=('glibc')
makedepends=('go')
provides=('ponto')
conflicts=('ponto' 'ponto-bin')
source=("\$pkgname-\$pkgver.tar.gz::https://github.com/$REPO/archive/v\$pkgver.tar.gz")
sha256sums=('$SHA256')
options=('!debug')

build() {
    cd "\$pkgname-\$pkgver"
    export CGO_CPPFLAGS="\${CPPFLAGS}"
    export CGO_CFLAGS="\${CFLAGS}"
    export CGO_CXXFLAGS="\${CXXFLAGS}"
    export CGO_LDFLAGS="\${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -mod=readonly -modcacherw"
    go build -ldflags "-s -w -X main.version=\${pkgver}" -o ponto ./cmd/ponto

    # Generate completions
    ./ponto completion bash > ponto.bash
    ./ponto completion zsh > ponto.zsh
    ./ponto completion fish > ponto.fish
}

package() {
    cd "\$pkgname-\$pkgver"
    install -Dm755 ponto "\$pkgdir/usr/bin/ponto"
    install -Dm644 MIT-LICENSE "\$pkgdir/usr/share/licenses/\$pkgname/MIT-LICENSE"
    install -Dm644 ponto.bash "\$pkgdir/usr/share/bash-completion/completions/ponto"
    install -Dm644 ponto.zsh "\$pkgdir/usr/share/zsh/site-functions/_ponto"
    install -Dm644 ponto.fish "\$pkgdir/usr/share/fish/vendor_completions.d/ponto.fish"
}
EOF

# Generate .SRCINFO
cat > .SRCINFO << EOF
pkgbase = ponto-cli
	pkgdesc = CLI for Ponto
	pkgver = $VERSION
	pkgrel = 1
	url = https://github.com/$REPO
	arch = x86_64
	arch = aarch64
	license = MIT
	makedepends = go
	depends = glibc
	provides = ponto
	conflicts = ponto
	conflicts = ponto-bin
	options = !debug
	source = ponto-cli-$VERSION.tar.gz::https://github.com/$REPO/archive/v$VERSION.tar.gz
	sha256sums = $SHA256

pkgname = ponto-cli
EOF

# Clone AUR repo and push
mkdir -p ~/.ssh
echo "$AUR_KEY" > ~/.ssh/aur
chmod 600 ~/.ssh/aur
cat >> ~/.ssh/config << SSHEOF
Host aur.archlinux.org
    IdentityFile ~/.ssh/aur
    User aur
    StrictHostKeyChecking accept-new
SSHEOF

git clone ssh://aur@aur.archlinux.org/ponto-cli.git aur-repo
cp PKGBUILD .SRCINFO aur-repo/
cd aur-repo
git config user.name "cli-release-bot"
git config user.email "cli-release-bot@users.noreply.github.com"
git add PKGBUILD .SRCINFO
if git diff --cached --quiet; then
  echo "AUR package already up to date for $VERSION"
else
  git commit -m "Update to $VERSION"
  git push
fi

echo "Published ponto-cli $VERSION to AUR"
