"""Fail closed unless running on the disposable CI VM or our offline container."""
import json
import os
from pathlib import Path
import re
import subprocess

def require():
    if os.geteuid()!=0:raise RuntimeError('DISPOSABLE_ROOT_REQUIRED')
    github=(os.environ.get('GITHUB_ACTIONS')=='true' and os.environ.get('RUNNER_ENVIRONMENT')=='github-hosted' and os.environ.get('NEEXITLAG_ISOLATED_UPGRADE')=='github-disposable-runner')
    if github:
        if Path('/.dockerenv').exists():raise RuntimeError('GITHUB_HOSTED_VM_REQUIRED')
        return 'github-hosted'
    if os.environ.get('VELIQ_LOCAL_ACCEPTANCE')!='disposable-docker':raise RuntimeError('DISPOSABLE_ENVIRONMENT_REQUIRED')
    marker=Path('/etc/veliq-local-acceptance')
    if not Path('/.dockerenv').exists() or not marker.is_file() or marker.is_symlink():raise RuntimeError('DISPOSABLE_CONTAINER_REQUIRED')
    if marker.stat().st_uid!=0 or marker.stat().st_mode&0o777!=0o600 or marker.read_text().strip()!='disposable-docker/v1':raise RuntimeError('DISPOSABLE_MARKER_INVALID')
    if Path('/proc/1/comm').read_text().strip()!='systemd':raise RuntimeError('REAL_SYSTEMD_REQUIRED')
    mounts=Path('/proc/self/mountinfo').read_text().splitlines()
    if not any(line.split()[4]=='/run' and any(field.startswith('shared:') for field in line.split()[6:line.split().index('-')]) for line in mounts):raise RuntimeError('RUN_MOUNT_PROPAGATION_REQUIRED')
    links=json.loads(subprocess.check_output(['ip','-j','link','show']))
    if any(link['ifname']!='lo' for link in links):raise RuntimeError('OFFLINE_LOOPBACK_ONLY_REQUIRED')
    if json.loads(subprocess.check_output(['ip','-j','route','show','default'])):raise RuntimeError('DEFAULT_ROUTE_FORBIDDEN')
    return 'disposable-docker'

def commit():
    mode=require()
    value=Path('/acceptance-commit').read_text().strip() if mode=='disposable-docker' else subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
    if not re.fullmatch('[a-f0-9]{40}',value):raise RuntimeError('EXACT_FIXTURE_COMMIT_REQUIRED')
    return value

def wait_fixture_health(engine):
    require()
    host=engine.Host()
    host.relay_health='http://127.0.0.1:9090/healthz'
    return host.wait_health()


if __name__=='__main__':
    import sys
    if sys.argv[1:]==['--commit']:print(commit())
    elif not sys.argv[1:]:require()
    else:raise SystemExit(90)
