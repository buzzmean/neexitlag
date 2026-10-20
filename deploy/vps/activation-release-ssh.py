#!/usr/bin/env python3
"""Host-verified SSH stdin artifact transport; no enrollment and no CA rotation."""
import base64
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

NAMES={'deploy-activation.py','fresh-vps-upgrade.py','production-engine.py','renew-server-cert.py','fresh-pc','codes','relay','control-plane','veliq-certificate-renewal.service','veliq-certificate-renewal.timer','92-veliq-activation.conf'}
def decode(item):
 raw=base64.b64decode(item['data'],validate=True)
 if len(raw)>32*1024*1024 or hashlib.sha256(raw).hexdigest()!=item['sha256']:raise RuntimeError('ARTIFACT_HASH_OR_SIZE_MISMATCH')
 return raw

def main():
 if os.geteuid()!=0 or os.uname().nodename.split('.')[0].lower()=='veliq-swe-01':raise RuntimeError('ROOT_REQUIRED_OR_EXCLUDED_HOST')
 if 'VPS_PUBLIC_IP=2.56.177.4' not in Path('/etc/neexitlag/public/relay.env').read_text().splitlines():raise RuntimeError('VPS_IDENTITY_MISMATCH')
 raw=sys.stdin.buffer.read(160*1024*1024+1)
 if len(raw)>160*1024*1024:raise RuntimeError('BOUNDED_REQUEST_EXCEEDED')
 p=json.loads(raw)
 if set(p['files'])!=NAMES or not re.fullmatch('[a-f0-9]{40}',p['commit']):raise RuntimeError('RELEASE_MANIFEST_INVALID')
 if p['confirm']!='ACTIVATE-CODES-2.56.177.4':raise RuntimeError('EXPLICIT_DEPLOYMENT_CONFIRMATION_REQUIRED')
 lock=os.open('/run/veliq-activation-release.lock',os.O_CREAT|os.O_RDWR|os.O_NOFOLLOW,0o600)
 with os.fdopen(lock,'a') as f:
  fcntl.flock(f,fcntl.LOCK_EX|fcntl.LOCK_NB)
  with tempfile.TemporaryDirectory(prefix='veliq-activation-',dir='/var/tmp') as tmp:
   stage=Path(tmp)
   for name,item in p['files'].items():
    dest=stage/name;dest.write_bytes(decode(item));dest.chmod(0o700)
   bundle=stage/'bundle';(bundle/'bin').mkdir(parents=True);(bundle/'deploy/vps').mkdir(parents=True)
   import shutil
   for role in ('relay','control-plane'):shutil.copy2(stage/role,bundle/'bin'/role)
   shutil.copy2(stage/'production-engine.py',bundle/'deploy/vps/production-upgrade.py')
   result=subprocess.run(['/usr/bin/python3',str(stage/'deploy-activation.py'),'apply','--stage',str(stage),'--commit',p['commit'],'--confirm',p['confirm']],check=False)
   if result.returncode:raise RuntimeError('ACTIVATION_RELEASE_FAILED_SEE_SAFE_CODE')

if __name__=='__main__':
 try:main()
 except BaseException as exc:
  if isinstance(exc,SystemExit):raise
  code=str(exc) if re.fullmatch('[A-Z0-9_]+',str(exc)) else 'SAFE_RELEASE_FAILED'
  print('ACTIVATION_RELEASE_FAILED code='+code,file=sys.stderr);sys.exit(1)
