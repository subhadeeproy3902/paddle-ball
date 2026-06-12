# Releases, GitHub Actions & Tokens

Everything about how this repo builds, tests, and ships — and which tokens are
involved (and which you can ignore).

---

## 1. The workflow at a glance

One workflow file, [.github/workflows/release.yml](.github/workflows/release.yml), with two jobs:

| Job | Runs when | Does |
|---|---|---|
| **CI** | every push to `main` and every pull request | `go build` · `go vet` · `go test` |
| **GoReleaser** | only on `v*` tags (e.g. `v1.0.1`) | cross-compiles binaries, builds packages + a Docker image, and publishes a GitHub Release |

They are mutually exclusive — the CI job is skipped on tags, the release job is
skipped on branch pushes — so a normal `git push` only ever runs the quick CI
checks, and only a version tag triggers a real release.

```
git push origin main      →  CI job (build/vet/test)        ✅ fast check
git push origin v1.2.3    →  GoReleaser job (full release)   📦 binaries + release
```

---

## 2. Tokens — what you actually need

There are two tokens in play. **You only have to create one of them, and only if
you want Homebrew.**

### `GITHUB_TOKEN` — automatic, nothing to do
GitHub Actions injects a fresh `GITHUB_TOKEN` into every workflow run. The
workflow grants it `contents: write` (to create the Release) and
`packages: write` (to push the Docker image to GitHub Container Registry). This
is why releases work out of the box — **no personal token required.**

> You do **not** need the personal `ghp_…` token you may have created for
> pushing from your laptop. If you ever pasted one somewhere public, revoke it
> at <https://github.com/settings/tokens>.

### `HOMEBREW_TOKEN` — optional, only for the Homebrew tap
GoReleaser can auto-update a separate Homebrew tap repo
(`homebrew-paddle-ball`) on each release. Pushing to *another* repo needs a
token with permission to do so — the automatic `GITHUB_TOKEN` can't reach
outside this repo. That's what `HOMEBREW_TOKEN` is for.

Right now the formula step is set to `skip_upload: true` in
[.goreleaser.yaml](.goreleaser.yaml), so **releases succeed without it.** To turn
Homebrew on, see §5.

---

## 3. Cut a release

```bash
git tag v1.0.1          # pick the next version (semver)
git push origin v1.0.1  # this push triggers the GoReleaser job
```

Watch it run:

```bash
gh run watch            # live status of the latest run
gh release view v1.0.1  # the published release once it finishes
```

To undo a bad tag before/after a release:

```bash
git tag -d v1.0.1               # delete locally
git push --delete origin v1.0.1 # delete on GitHub
gh release delete v1.0.1        # delete the release if one was made
```

---

## 4. What a release publishes

Each `v*` tag produces, attached to the GitHub Release:

- **Binaries** for Linux, macOS, and Windows × amd64 + arm64 (no Windows/arm64),
  as `paddle-ball_<os>_<arch>.tar.gz` (`.zip` on Windows).
- **Linux packages** — `.deb`, `.rpm`, and `.apk`.
- **`checksums.txt`** — SHA-256 of every artifact.
- **A Docker image** pushed to `ghcr.io/subhadeeproy3902/paddle-ball:<tag>` and `:latest`.
- **A Homebrew formula** (built as an artifact; pushed to the tap only once §5 is done).

---

## 5. Install paths, per platform

| Platform | Command |
|---|---|
| macOS / Linux | `curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/paddle-ball/main/install.sh \| bash` |
| Any Go env | `go install github.com/subhadeeproy3902/paddle-ball@latest` |
| Docker | `docker run --rm -it ghcr.io/subhadeeproy3902/paddle-ball:latest` |
| Debian/Ubuntu | download the `.deb` from Releases → `sudo dpkg -i paddle-ball_*.deb` |
| Fedora/RHEL | download the `.rpm` → `sudo rpm -i paddle-ball_*.rpm` |
| Alpine | download the `.apk` → `sudo apk add --allow-untrusted paddle-ball_*.apk` |
| Windows | download the `.zip` from Releases, unzip, add the folder to `PATH` (or `go install`) |

The `install.sh` script auto-detects OS + arch and pulls the matching archive
from the latest release.

---

## 6. Homebrew — enabled

The `homebrew-paddle-ball` tap and the `HOMEBREW_TOKEN` secret are configured,
and `brews[].skip_upload` is `false`, so each `v*` tag pushes an updated
`Formula/paddle-ball.rb` to the tap. Users install with:

```bash
brew install subhadeeproy3902/paddle-ball/paddle-ball
```

If a release ever fails on the Homebrew step, it means the `HOMEBREW_TOKEN`
secret is missing or its scope can't push to the tap repo — recreate it as a
classic PAT with `repo` scope (or a fine-grained token with Contents:
read/write on `homebrew-paddle-ball`).

## 7. Scoop & WinGet (Windows)

Windows users can already use the prebuilt `.zip`, `go install`, or Docker.
GoReleaser also **generates Scoop and WinGet manifests** on every release
(`scoops:` and `winget:` in [.goreleaser.yaml](.goreleaser.yaml)), currently
with `skip_upload: true` so they're build artifacts and never block a release.
To make `scoop install` / `winget install` work:

- **Scoop** — create a public **`scoop-paddle-ball`** bucket repo, ensure a
  token can push to it, and set `scoops[].skip_upload: false`. Then:
  `scoop bucket add paddle-ball https://github.com/subhadeeproy3902/scoop-paddle-ball; scoop install paddle-ball`.
- **WinGet** — point `winget[].repository` at your fork of
  `microsoft/winget-pkgs`, set `skip_upload: false`, and GoReleaser opens the
  catalog PR (subject to Microsoft's review).

---

## 8. Test the release pipeline locally (no tag, no publish)

```bash
go install github.com/goreleaser/goreleaser/v2@latest
goreleaser check                              # validate the config
goreleaser release --snapshot --clean --skip=publish,docker
# → artifacts appear in ./dist without touching GitHub
```

---

## Troubleshooting

| Problem | Fix |
|---|---|
| CI fails on `go test` | run `go test ./...` locally; the physics suite must pass |
| Release fails on Docker push | ensure the repo has Actions → Packages write permission (it's set in the workflow) |
| Release fails on Homebrew | you set `skip_upload: false` without a valid `HOMEBREW_TOKEN` + tap repo — see §6 |
| Release fails on snap | snap packaging was removed; re-add `snapcrafts:` only with a snapcraft install step |
| Game renders odd glyphs | use a UTF-8, 256-color terminal (Windows Terminal, not legacy cmd.exe) |

---

*Questions or bugs → <https://github.com/subhadeeproy3902/paddle-ball/issues>*
