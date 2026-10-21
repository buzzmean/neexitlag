# Test-only Debian 13 userspace. No production files/identities enter the image.
FROM debian:trixie-slim
RUN --mount=type=secret,id=proxy_ca,required=false \
    set -eu; \
    if [ -s /run/secrets/proxy_ca ]; then \
      printf 'Acquire::https::CaInfo "/run/secrets/proxy_ca";\n' >/etc/apt/apt.conf.d/99-ci-proxy-ca; \
    fi; \
    apt-get update -qq; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      systemd systemd-sysv postgresql postgresql-client python3 openssl curl \
      iproute2 ca-certificates procps sudo git acl; \
    rm -f /etc/apt/apt.conf.d/99-ci-proxy-ca; \
    rm -rf /var/lib/apt/lists/*; \
    systemctl mask getty.target console-getty.service
RUN printf 'disposable-docker/v1\n' >/etc/veliq-local-acceptance && chmod 0600 /etc/veliq-local-acceptance
STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]
