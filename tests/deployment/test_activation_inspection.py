import importlib.util
import json
import os
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest import mock
ROOT=Path(__file__).resolve().parents[2]
def load(name):
 s=importlib.util.spec_from_file_location(name,ROOT/'deploy/vps'/name);m=importlib.util.module_from_spec(s);s.loader.exec_module(m);return m

class ReadOnlyInspectionTests(unittest.TestCase):
 def test_baseline_is_not_misrepresented_as_proof_of_historical_no_changes(self):
  m=load('inspect-activation.py')
  with tempfile.TemporaryDirectory() as directory:
   root=Path(directory);calls=[]
   key=root/'etc/neexitlag/control/ca.key';key.parent.mkdir(parents=True);key.write_bytes(b'PRIVATE_FIXTURE_NEVER_PRINT');key.chmod(0o600)
   backup=root/'var/backups/neexitlag-activation/20261020T140000Z-aabbccdd';backup.mkdir(parents=True)
   (backup/'state.json').write_text(json.dumps({'phase':'rolled-back','files':{},'binary_id':''}))
   report={'services':{'neexitlag-control-plane':{'pid':150293},'neexitlag-relay':{'pid':150299}},'binary_sha256':{'relay':'public-binary','control-plane':'public-binary'},'capabilities':{'relay':{'commit':m.BASELINE},'control-plane':{'commit':m.BASELINE}},'applied_schema_version':1,'ports':{},'ca_migration_required':False}
   class Host:
    database='fixture'
    def path(self,s):return root/s.lstrip('/')
    def running_binary_hash(self,pid):return 'public-binary'
    def systemctl(self,*args,**kw):calls.append(args);return 'active'
    def run(self,args,**kw):calls.append(args);return (1,b'') if args[0]=='runuser' else (0,b'')
    def sql(self,q,d):calls.append(q);return '0' if 'information_schema' in q else '1'
   out=m.inspect(SimpleNamespace(report=lambda h:report),SimpleNamespace(Host=object),root=root,host=Host())
   text=json.dumps(out)
   self.assertNotIn('PRIVATE_FIXTURE',text);self.assertFalse(out['historical_no_changes_proven'])
   self.assertTrue(out['binaries_and_schema_at_baseline']);self.assertTrue(out['database']['activation_migration_not_applied'])
   self.assertEqual(out['verdict'],'BASELINE_WITH_DEPLOYMENT_ARTIFACTS')
   self.assertFalse(out['ca_access']['direct_service_readable'])
   for call in calls:
    if isinstance(call,tuple):self.assertNotIn(call[0],('start','stop','restart','enable','disable','systemd-run'))
    elif isinstance(call,list):self.assertNotIn(call[0],('pg_dump','pg_restore'))
    else:self.assertTrue(call.startswith('SELECT '))
 def test_queries_are_server_enforced_read_only_and_use_no_tempfiles(self):
  m=load('inspect-activation.py');engine=load('production-upgrade.py');seen=[]
  def report(h):
   h.database='fixture';h.db_port=5432
   value=h.sql('SHOW server_version_num','fixture');self.assertEqual(value,'170006')
   raise RuntimeError('STOP_FIXTURE_BEFORE_OTHER_INSPECTION')
  def process(args,**kw):
   seen.append(args);return SimpleNamespace(returncode=0,stdout=b'170006\n',stderr=b'')
  with mock.patch.object(m.subprocess,'run',side_effect=process),mock.patch.object(engine.tempfile,'TemporaryFile',side_effect=AssertionError('no remote files')):
   with self.assertRaisesRegex(RuntimeError,'STOP_FIXTURE'):m.inspect(SimpleNamespace(report=report),engine)
  self.assertTrue(seen[0][-1].startswith('BEGIN READ ONLY; '));self.assertTrue(seen[0][-1].endswith('; COMMIT;'))
