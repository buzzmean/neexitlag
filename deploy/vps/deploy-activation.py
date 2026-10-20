#!/usr/bin/env python3
"""Explicit existing-VPS activation release. Import is side-effect free.
Uses the proven fresh binary transaction; never restores DB security history.
"""
import argparse
import datetime as dt
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import sys
import subprocess
import tempfile

ROOT=Path('/var/backups/neexitlag-activation')
RECOVERY=Path('/etc/systemd/system/veliq-activation-recover.service')
MANAGED={
 'renew-server-cert.py':'/opt/neexitlag/maintenance/renew-server-cert.py',
 'fresh-pc':'/opt/neexitlag/maintenance/veliq-fresh-helper',
 'codes':'/opt/neexitlag/maintenance/veliq-codes',
 'veliq-certificate-renewal.service':'/etc/systemd/system/veliq-certificate-renewal.service',
 'veliq-certificate-renewal.timer':'/etc/systemd/system/veliq-certificate-renewal.timer',
 '92-veliq-activation.conf':'/etc/systemd/system/neexitlag-control-plane.service.d/92-veliq-activation.conf',
}

def module(name,path):
 s=importlib.util.spec_from_file_location(name,path);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m

def save(path,state):
 tmp=path.with_suffix('.tmp');tmp.write_text(json.dumps(state));tmp.chmod(0o600)
 with tmp.open('rb') as f:os.fsync(f.fileno())
 os.replace(tmp,path)
 fd=os.open(path.parent,os.O_DIRECTORY)
 try:os.fsync(fd)
 finally:os.close(fd)

def guard():
 if os.geteuid()!=0 or os.uname().nodename.split('.')[0].lower()=='veliq-swe-01':raise RuntimeError('ROOT_REQUIRED_OR_EXCLUDED_HOST')
 if 'VPS_PUBLIC_IP=2.56.177.4' not in Path('/etc/neexitlag/public/relay.env').read_text().splitlines():raise RuntimeError('VPS_IDENTITY_MISMATCH')

def restore(directory,automatic=False,boot=False):
 directory=Path(directory);state=json.loads((directory/'state.json').read_text())
 if automatic and state['phase']!='pending':return
 engine=module('activation_rollback_upgrade',directory/'fresh-vps-upgrade.py')
 leaf=module('activation_rollback_leaf',directory/'renew-server-cert.py')
 run=leaf.command
 if Path(MANAGED['veliq-certificate-renewal.timer']).exists():
  run(['systemctl','disable','veliq-certificate-renewal.timer']) if boot else run(['systemctl','disable','--now','veliq-certificate-renewal.timer'])
 # A renewal's independent watchdog must finish before restoring the outer leaf.
 if boot:
  if leaf.BACKUPS.exists():
   for p in leaf.BACKUPS.glob('*/transaction.json'):
    prior=json.loads(p.read_text())
    if prior['phase']=='pending':
     if (p.parent/'probe.id').exists():run(['/opt/neexitlag/maintenance/veliq-fresh-helper','upgrade-cleanup','--vps-ip','2.56.177.4','--journal',str(p.parent/'probe.id')])
     prior['phase']='rolled-back';save(p,prior)
 else:
  leaf.Renewal().recover_pending(leaf.BACKUPS)
  run(['systemctl','stop','neexitlag-control-plane','neexitlag-relay'])
 for path,meta in state['files'].items():
  dest=Path(path)
  if meta is None:
   if dest.exists():dest.unlink()
  else:
   data=(directory/'saved'/meta['name']).read_bytes()
   if hashlib.sha256(data).hexdigest()!=meta['sha256']:raise RuntimeError('BACKUP_HASH_MISMATCH')
   leaf.atomic(dest,data,meta)
 run(['systemctl','daemon-reload'])
 if state.get('binary_id'):
  with engine.transaction_lock():engine.LocalUpgrade(engine.e.Host()).rollback(state['binary_id'],boot=boot)
 elif not boot:
  run(['systemctl','start','neexitlag-control-plane']);run(['systemctl','start','neexitlag-relay'])
 if state['timer_active']:run(['systemctl','enable','veliq-certificate-renewal.timer'] if boot else ['systemctl','enable','--now','veliq-certificate-renewal.timer'])
 state['phase']='rolled-back';save(directory/'state.json',state)
 # No SQL restore: nonce, revocations, new devices and additive tables survive.

def apply(stage,commit,confirm):
 guard()
 if confirm!='ACTIVATE-CODES-2.56.177.4':raise RuntimeError('EXPLICIT_DEPLOYMENT_CONFIRMATION_REQUIRED')
 stage=Path(stage)
 engine=module('activation_upgrade',stage/'fresh-vps-upgrade.py');leaf=module('activation_leaf',stage/'renew-server-cert.py')
 report=engine.report(engine.e.Host());print('ACTIVATION_PREFLIGHT='+json.dumps(report),flush=True)
 if not engine.compatible(report['capabilities']['relay'],'relay') or not engine.compatible(report['capabilities']['control-plane'],'control-plane'):
  raise RuntimeError('CURRENT_PROTOCOL_3_BASELINE_REQUIRED')
 if report['systemd_version']<247:raise RuntimeError('SYSTEMD_LOADCREDENTIAL_REQUIRES_247')
 if not Path('/etc/neexitlag/echo/echo.token').is_file():raise RuntimeError('EXISTING_ECHO_CREDENTIAL_REQUIRED')
 # Read-only preflight verifies same CA/SAN/root key before any upgrade writes.
 leaf.Renewal().preflight(dt.datetime.now(dt.timezone.utc),'2.56.177.4')
 ROOT.mkdir(mode=0o700,parents=True,exist_ok=True)
 if ROOT.is_symlink() or ROOT.stat().st_mode&0o077:raise RuntimeError('UNSAFE_BACKUP_ROOT')
 for prior in ROOT.glob('*/state.json'):
  if json.loads(prior.read_text())['phase']=='pending':raise RuntimeError('PRIOR_ACTIVATION_DEPLOYMENT_PENDING')
 directory=ROOT/(dt.datetime.now(dt.timezone.utc).strftime('%Y%m%dT%H%M%SZ')+'-'+secrets.token_hex(4));directory.mkdir(mode=0o700)
 (directory/'saved').mkdir(mode=0o700)
 # Permanent full configuration backup. Binary transaction separately pg_dumps
 # and saves all binaries/units before replacing anything.
 shutil.copytree('/etc/neexitlag',directory/'config')
 for name in ('deploy-activation.py','fresh-vps-upgrade.py','renew-server-cert.py'):
  shutil.copy2(stage/name,directory/name)
 shutil.copy2(stage/'production-engine.py',directory/'production-engine.py')
 run=leaf.command
 active=engine.e.Host().systemctl('is-enabled','veliq-certificate-renewal.timer',check=False)=='enabled'
 state={'phase':'pending','binary_id':'','files':{},'timer_active':active}
 for index,path in enumerate([*MANAGED.values(),*[str(Path('/etc/neexitlag')/x) for x in leaf.FILES]]):
  p=Path(path)
  if p.is_symlink():raise RuntimeError('MANAGED_PATH_SYMLINK_REFUSED')
  meta=None
  if p.exists():
   st=p.stat();data=p.read_bytes();name=str(index);(directory/'saved'/name).write_bytes(data)
   meta={'name':name,'mode':st.st_mode&0o777,'uid':st.st_uid,'gid':st.st_gid,'sha256':hashlib.sha256(data).hexdigest()}
  state['files'][path]=meta
 save(directory/'state.json',state)
 recovery='[Unit]\nDescription=VELIQ activation release recovery\nAfter=postgresql.service\nBefore=neexitlag-control-plane.service neexitlag-relay.service\n[Service]\nType=oneshot\nExecStart=/usr/bin/python3 '+str(directory/'deploy-activation.py')+' rollback --backup '+str(directory)+' --automatic --boot\n[Install]\nWantedBy=multi-user.target\n'
 leaf.atomic(RECOVERY,recovery.encode(),{'mode':0o644,'uid':0,'gid':0})
 run(['systemctl','daemon-reload']);run(['systemctl','enable',RECOVERY.name])
 unit='veliq-activation-rollback-'+directory.name
 run(['systemd-run','--unit='+unit,'--on-active=480','--timer-property=AccuracySec=1s','/usr/bin/python3',str(directory/'deploy-activation.py'),'rollback','--backup',str(directory),'--automatic'])
 try:
  if active:run(['systemctl','disable','--now','veliq-certificate-renewal.timer'])
  if Path(MANAGED['veliq-certificate-renewal.service']).exists():run(['systemctl','stop','veliq-certificate-renewal.service'])
  leaf.Renewal().recover_pending(leaf.BACKUPS)
  def record(summary):
   state['binary_id']=summary['id'];save(directory/'state.json',state)
  # Force current release even when old binaries already support /3.
  result=engine.prepare(stage/'bundle',stage/'fresh-pc',commit,'UPGRADE-VPS-2.56.177.4',force_update=True,before_commit=record)
  print('ACTIVATION_BINARY_RESULT='+json.dumps(result),flush=True)
  run(['systemctl','disable','--now','veliq-certificate-renewal.timer']) if active else None
  for name,path in MANAGED.items():
   p=Path(path);p.parent.mkdir(mode=0o755,parents=True,exist_ok=True)
   leaf.atomic(p,(stage/name).read_bytes(),{'mode':0o700 if name in ('fresh-pc','codes','renew-server-cert.py') else 0o644,'uid':0,'gid':0})
  leaf.BACKUPS.mkdir(mode=0o700,parents=True,exist_ok=True)
  run(['systemctl','daemon-reload']);run(['systemctl','restart','neexitlag-control-plane'])
  # Leaf operation has its own durable backup and independent watchdog/proof.
  p=subprocess.run(['/usr/bin/python3',MANAGED['renew-server-cert.py'],'run','--force','--confirm','RENEW-SERVER-2.56.177.4'],timeout=165,check=False)
  if p.returncode:raise RuntimeError('LEAF_RENEWAL_FAILED')
  import ssl,urllib.request
  context=ssl.create_default_context(cafile='/etc/neexitlag/control/ca.pem')
  with urllib.request.urlopen('https://127.0.0.1:8443/api/v1/enroll/policy',context=context,timeout=5) as response:
   policy=json.loads(response.read(65537))
  if policy.get('enrollment_ready') is not True:raise RuntimeError('ACTIVATION_POLICY_NOT_READY')
  run(['systemctl','enable','--now','veliq-certificate-renewal.timer'])
  state['phase']='committed';save(directory/'state.json',state)
  run(['systemctl','stop',unit+'.timer'])
  print('ACTIVATION_DEPLOYMENT_COMMITTED backup='+str(directory),flush=True)
 except BaseException:
  restore(directory);raise

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('mode',choices=('apply','rollback'));p.add_argument('--stage');p.add_argument('--commit');p.add_argument('--confirm',default='');p.add_argument('--backup');p.add_argument('--automatic',action='store_true');p.add_argument('--boot',action='store_true');a=p.parse_args();os.umask(0o077);guard()
 if a.mode=='rollback':
  if not a.backup or not re.fullmatch(r'/var/backups/neexitlag-activation/\d{8}T\d{6}Z-[a-f0-9]{8}',a.backup):raise RuntimeError('INVALID_BACKUP_PATH')
  restore(a.backup,a.automatic,a.boot)
 else:
  if not a.stage or not re.fullmatch('[a-f0-9]{40}',a.commit or ''):raise RuntimeError('EXACT_RELEASE_COMMIT_REQUIRED')
  apply(a.stage,a.commit,a.confirm)

if __name__=='__main__':
 try:main()
 except BaseException as exc:
  if isinstance(exc,SystemExit):raise
  code=str(exc) if re.fullmatch('[A-Z0-9_]+',str(exc)) else 'SAFE_DEPLOYMENT_FAILED'
  print('ACTIVATION_DEPLOYMENT_FAILED code='+code,file=sys.stderr);sys.exit(1)
