import contextlib
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest import mock

ROOT=Path(__file__).resolve().parents[2]
def load(name):
 s=importlib.util.spec_from_file_location(name,ROOT/'deploy/vps'/name);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m

class ActivationReleaseTests(unittest.TestCase):
 def test_artifact_hash_and_bounded_transport_never_trusts_manifest_claim(self):
  import base64
  transport=load('activation-release-ssh.py');raw=b'public-test-artifact'
  item={'data':base64.b64encode(raw).decode(),'sha256':hashlib.sha256(raw).hexdigest()}
  self.assertEqual(transport.decode(item),raw)
  item['sha256']='0'*64
  with self.assertRaisesRegex(RuntimeError,'ARTIFACT_HASH_OR_SIZE_MISMATCH'):transport.decode(item)
 def test_binary_force_is_explicit_and_retains_default_noop(self):
  upgrade=load('fresh-vps-upgrade.py');events=[]
  current={'upgrade_needed':False,'local_vps_address':True}
  up=SimpleNamespace(apply=lambda *a:{'id':'fixture'},prove_commit=lambda i:events.append(('prove',i)) or {'pass':True})
  with tempfile.TemporaryDirectory() as tmp:
   bundle=Path(tmp);(bundle/'deploy/vps').mkdir(parents=True)
   with mock.patch.object(upgrade,'report',return_value=current),mock.patch.object(upgrade.e,'Host'),mock.patch.object(upgrade,'LocalUpgrade',return_value=up),mock.patch.object(upgrade,'capabilities',return_value={}),mock.patch.object(upgrade.e,'digest',return_value='public-test-hash'),mock.patch.object(upgrade,'transaction_lock',side_effect=contextlib.nullcontext):
    self.assertEqual(upgrade.prepare(bundle,'helper','abc',''),current)
    with self.assertRaisesRegex(Exception,'EXPLICIT_UPGRADE_CONFIRMATION_REQUIRED'):upgrade.prepare(bundle,'helper','abc','',force_update=True)
    result=upgrade.prepare(bundle,'helper','abc','UPGRADE-VPS-2.56.177.4',force_update=True,before_commit=lambda x:events.append(('saved',x['id'])))
    self.assertTrue(result['pass']);self.assertEqual(events,[('saved','fixture'),('prove','fixture')])
 def test_failure_rollback_preserves_security_writes_and_additive_schema(self):
  release=load('deploy-activation.py');leaf=load('renew-server-cert.py')
  for boot in (False,True):
   with self.subTest(boot=boot),tempfile.TemporaryDirectory() as tmp:
    root=Path(tmp);backup=root/'backup';(backup/'saved').mkdir(parents=True)
    old=root/'old-leaf';old.write_bytes(b'new-leaf');added=root/'activation.conf';added.write_bytes(b'new-unit')
    db=root/'security-state';db.write_bytes(b'current-nonce-current-revocations-new-devices')
    (backup/'saved/0').write_bytes(b'old-leaf')
    state={'phase':'pending','binary_id':'fixture','timer_active':False,'files':{str(old):{'name':'0','mode':0o600,'uid':__import__('os').getuid(),'gid':__import__('os').getgid(),'sha256':hashlib.sha256(b'old-leaf').hexdigest()},str(added):None}}
    release.save(backup/'state.json',state);calls=[]
    fake=SimpleNamespace(e=SimpleNamespace(Host=lambda:None),LocalUpgrade=lambda h:SimpleNamespace(rollback=lambda i,**kw:calls.append(('rollback',i,kw))),transaction_lock=contextlib.nullcontext)
    renewal=leaf.Renewal(root/'config',lambda args:calls.append(args) or '')
    with mock.patch.object(release,'module',side_effect=[fake,leaf]),mock.patch.object(leaf,'command',side_effect=lambda args:calls.append(args) or ''),mock.patch.object(leaf,'BACKUPS',root/'leaf-backups'),mock.patch.object(leaf,'Renewal',return_value=renewal):release.restore(backup,boot=boot)
    self.assertEqual(old.read_bytes(),b'old-leaf');self.assertFalse(added.exists());self.assertEqual(db.read_bytes(),b'current-nonce-current-revocations-new-devices')
    self.assertEqual(json.loads((backup/'state.json').read_text())['phase'],'rolled-back')
    self.assertIn(('rollback','fixture',{'boot':boot}),calls)
    self.assertFalse(any('pg_restore' in str(c) for c in calls))
    if boot:self.assertFalse(any(isinstance(c,list) and c[:2]==['systemctl','start'] for c in calls))
 def test_bad_backup_refuses_restore(self):
  release=load('deploy-activation.py');leaf=load('renew-server-cert.py')
  with tempfile.TemporaryDirectory() as tmp:
   root=Path(tmp);(root/'saved').mkdir();dest=root/'leaf';dest.write_bytes(b'new');(root/'saved/0').write_bytes(b'corrupt')
   release.save(root/'state.json',{'phase':'pending','files':{str(dest):{'name':'0','sha256':'0'*64}},'timer_active':False})
   with mock.patch.object(release,'module',side_effect=[SimpleNamespace(),leaf]),mock.patch.object(leaf,'command',return_value=''),mock.patch.object(leaf,'BACKUPS',root/'none'):
    with self.assertRaisesRegex(RuntimeError,'BACKUP_HASH_MISMATCH'):release.restore(root)
   self.assertEqual(dest.read_bytes(),b'new')
