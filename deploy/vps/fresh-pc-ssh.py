#!/usr/bin/env python3
"""Executed only inside an interactive, host-verified owner SSH connection.
Stdout is a credential response, captured privately by bootstrap; never a log.
"""
import base64
import hashlib
import fcntl
import json
import os
import re
import ssl
from pathlib import Path
import subprocess
import sys
import tempfile


def main():
    if os.geteuid() != 0 or os.uname().nodename.split('.')[0].lower() == 'veliq-swe-01':
        raise RuntimeError('ROOT_REQUIRED_OR_EXCLUDED_HOST')
    raw = sys.stdin.buffer.read(64 * 1024 * 1024 + 1)
    if len(raw) > 64 * 1024 * 1024:
        raise RuntimeError('BOUNDED_REQUEST_EXCEEDED')
    p = json.loads(raw)
    if p['vps_ip'] != '2.56.177.4' or 'VPS_PUBLIC_IP=' + p['vps_ip'] not in Path('/etc/neexitlag/public/relay.env').read_text().splitlines():
        raise RuntimeError('EXISTING_VPS_IDENTITY_MISMATCH')
    if p['rotate'] and p['confirm'] != 'ROTATE-CA-2.56.177.4':
        raise RuntimeError('EXPLICIT_CA_ROTATION_CONFIRMATION_REQUIRED')
    for role in ('control-plane', 'relay'):
        out = subprocess.run(['/opt/neexitlag/bin/' + role, '-version-json'],
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=10, check=True)
        info = json.loads(out.stdout)
        if info.get('role') != role or info.get('max_config_bytes', 0) < 1048576 or info.get('full_revocations') is not True:
            raise RuntimeError('EXISTING_FULL_REVOCATION_SUPPORT_REQUIRED')
        if role == 'relay' and ('neexitlag/3' not in info.get('protocols', []) or info.get('lease_seconds') != 1800):
            raise RuntimeError('EXISTING_PROTOCOL_3_REQUIRED')
    helper = base64.b64decode(p['helper'], validate=True)
    if hashlib.sha256(helper).hexdigest() != p['helper_sha256']:
        raise RuntimeError('HELPER_HASH_MISMATCH')
    lock_dir = Path('/run/neexitlag-fresh')
    lock_dir.mkdir(mode=0o700, exist_ok=True)
    if lock_dir.is_symlink() or lock_dir.stat().st_uid != 0 or lock_dir.stat().st_mode & 0o077:
        raise RuntimeError('UNSAFE_LOCK_DIRECTORY')
    lock = open(lock_dir / 'enrollment.lock', 'a')
    os.chmod(lock.name, 0o600)
    fcntl.flock(lock, fcntl.LOCK_EX)
    with tempfile.TemporaryDirectory(prefix='veliq-fresh-', dir='/var/tmp') as tmp:
        tool = Path(tmp) / 'fresh-pc'
        tool.write_bytes(helper)
        tool.chmod(0o700)
        if p['rotate']:
            # Migration is opt-in only. Normal enrollment never touches TLS/systemd.
            migration = Path(tmp) / 'rotate.py'
            migration.write_bytes(base64.b64decode(p['migration'], validate=True))
            current = hashlib.sha256(ssl.PEM_cert_to_DER_cert(Path('/etc/neexitlag/control/ca.pem').read_text())).hexdigest()
            expected = p['expected_ca'] or current
            subprocess.run(['/usr/bin/python3', str(migration), 'apply', '--confirm', p['confirm'],
                            '--expected-current-ca-sha256', expected, '--helper', str(tool),
                            '--helper-sha256', p['helper_sha256']], check=True, stdout=sys.stderr)
        current = hashlib.sha256(ssl.PEM_cert_to_DER_cert(Path('/etc/neexitlag/control/ca.pem').read_text())).hexdigest()
        pin = p['enrollment_ca'] or current
        request = json.dumps(p['request']).encode()
        # Exactly one request, no key material arrives from the PC. Result stdout
        # is reserved for the SSH channel; no repr(request/response) in exceptions.
        subprocess.run([str(tool), 'serve', '--vps-ip', p['vps_ip'], '--ca-sha256', pin], input=request, check=True)

if __name__ == '__main__':
    try:
        main()
    except BaseException as exc:
        code = str(exc).split(':', 1)[0] if type(exc) is RuntimeError else ''
        if not re.fullmatch(r'[A-Z][A-Z0-9_]{1,95}', code):
            code = type(exc).__name__
        print('FRESH_SSH_FAILED: ' + code, file=sys.stderr)
        sys.exit(1)
