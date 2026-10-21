#!/usr/bin/env bash
# Test-only, no SSH, no host /etc mounts, no production identity or Internet at runtime.
set -euo pipefail
[[ ! -e /etc/neexitlag && ! -e /opt/neexitlag ]] || { echo 'DO_NOT_RUN_ON_A_VPS_INSTALLATION' >&2; exit 90; }
repo=$(git rev-parse --show-toplevel)
cd "$repo"
layout=${1:-fresh-credential}
[[ $layout == legacy || $layout == fresh-credential ]] || { echo 'Usage: run-local-systemd.sh [legacy|fresh-credential]' >&2; exit 2; }
[[ -z $(git status --porcelain --untracked-files=no) ]] || { echo 'CLEAN_TRACKED_CHECKOUT_REQUIRED' >&2; exit 3; }
commit=$(git rev-parse HEAD)
work=$(mktemp -d /var/tmp/veliq-local-work.XXXXXXXX)
fixture=/var/tmp/veliq-local-fixture-${work##*.}
container=veliq-acceptance-${work##*.}
image=veliq/systemd-acceptance:debian13
report=${VELIQ_LOCAL_REPORT:-$repo/.e2e/local-systemd-$layout.json}
mkdir -p "$(dirname "$report")"
chmod 0700 "$work"
docker_local(){ env -u DOCKER_HOST -u DOCKER_CONTEXT -u DOCKER_TLS -u DOCKER_TLS_VERIFY -u DOCKER_CERT_PATH docker --host=unix:///var/run/docker.sock "$@"; }
cleanup(){
 code=$?
 trap - EXIT
 if docker_local inspect "$container" >/dev/null 2>&1; then
  docker_local exec -w /src "$container" python3 tests/deployment/local/collect.py "$commit" "$layout" "$code" >"$report" 2>/dev/null || true
  docker_local rm -f "$container" >/dev/null
 fi
 rm -rf -- "$work" "$fixture"
 echo "LOCAL_ACCEPTANCE exit=$code layout=$layout commit=$commit report=$report"
 exit "$code"
}
trap cleanup EXIT
docker_local info --format '{{.ServerVersion}}' >/dev/null
# Public system CA bundle is available only to this build step, not stored in image.
docker_local build --secret id=proxy_ca,src=/etc/ssl/certs/ca-certificates.crt -t "$image" -f tests/deployment/local/Dockerfile tests/deployment/local >"$work/image.log" 2>&1
mkdir -p "$work/bin"
export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
for role in relay control-plane; do go build -trimpath -o "$work/bin/$role" "./services/$role/cmd/$role"; done
for tool in remote-echo enroll fresh-pc codes; do go build -trimpath -o "$work/bin/$tool" "./tools/$tool"; done
go build -trimpath -o "$work/bin/neexitlag-route-test-linux-amd64" ./tools/route-test
NEEXITLAG_ISOLATED_UPGRADE_BUNDLE=$fixture VELIQ_LOCAL_FIXTURE_EXPORT=loopback-only go test ./tools/vps-provision -run '^TestIsolatedUpgradeBundleExport$' -count=1 >"$work/fixture.log" 2>&1
git archive --format=tar HEAD >"$work/source.tar"
printf '%s\n' "$commit" >"$work/commit"
# A new network namespace with only lo. No host network, cgroup bind, Docker socket,
# host /etc, credentials or repository bind mounts enter the container.
docker_local run -d --name "$container" --label com.veliq.acceptance=1 --network none --privileged --cgroupns=private --tmpfs /run --tmpfs /run/lock --tmpfs /tmp -e VELIQ_LOCAL_ACCEPTANCE=disposable-docker -e NEEXITLAG_ISOLATED_CA_LAYOUT="$layout" -e NO_PROXY=127.0.0.1,localhost,2.56.177.4 -e no_proxy=127.0.0.1,localhost,2.56.177.4 "$image" >"$work/container.id"
for attempt in {1..30}; do
 if docker_local exec "$container" test -S /run/systemd/private; then break; fi
 sleep 1
done
docker_local cp "$work/source.tar" "$container:/tmp/source.tar"
docker_local cp "$work/commit" "$container:/acceptance-commit"
docker_local exec "$container" sh -c 'mkdir -p /src/dist/relay/bin /src/dist/relay/deploy/vps; tar -xf /tmp/source.tar -C /src; rm /tmp/source.tar'
docker_local cp "$work/bin/." "$container:/src/dist/relay/bin/"
docker_local cp "$fixture" "$container:/var/tmp/neexitlag-ci-bundle"
docker_local exec -w /src "$container" sh -c 'cp deploy/vps/production-upgrade.py dist/relay/deploy/vps/; python3 deploy/vps/upgrade-manifest.py dist/relay "$(cat /acceptance-commit)"'
docker_local exec -w /src "$container" sh -c 'bash tests/deployment/isolated-upgrade.sh >/var/tmp/veliq-local-run.log 2>&1'
