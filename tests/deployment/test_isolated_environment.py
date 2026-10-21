import importlib.util
import json
import os
from pathlib import Path
from types import SimpleNamespace
import unittest
from unittest import mock

ROOT=Path(__file__).resolve().parents[2]
def load():
 spec=importlib.util.spec_from_file_location('isolated_environment',ROOT/'tests/deployment/isolated_environment.py')
 m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m);return m

class IsolatedEnvironmentTests(unittest.TestCase):
 def test_ordinary_host_and_nonroot_are_rejected(self):
  m=load()
  with mock.patch.dict(os.environ,{},clear=True),mock.patch.object(m.os,'geteuid',return_value=1000):
   with self.assertRaisesRegex(RuntimeError,'DISPOSABLE_ROOT_REQUIRED'):m.require()
  with mock.patch.dict(os.environ,{},clear=True),mock.patch.object(m.os,'geteuid',return_value=0):
   with self.assertRaisesRegex(RuntimeError,'DISPOSABLE_ENVIRONMENT_REQUIRED'):m.require()
 def check_local(self,links=None,route=None,pid='systemd',marker='disposable-docker/v1',mode=0o100600,shared=True):
  m=load()
  def read(path,*args,**kw):
   if str(path)=='/proc/1/comm':return pid
   if str(path)=='/proc/self/mountinfo':return '100 1 0:0 / /run rw '+('shared:1 ' if shared else '')+'- tmpfs tmpfs rw'
   return marker
  with mock.patch.dict(os.environ,{'VELIQ_LOCAL_ACCEPTANCE':'disposable-docker'},clear=True),mock.patch.object(m.os,'geteuid',return_value=0),mock.patch.object(Path,'exists',return_value=True),mock.patch.object(Path,'is_file',return_value=True),mock.patch.object(Path,'is_symlink',return_value=False),mock.patch.object(Path,'stat',return_value=SimpleNamespace(st_uid=0,st_mode=mode)),mock.patch.object(Path,'read_text',read),mock.patch.object(m.subprocess,'check_output',side_effect=[json.dumps(links or [{'ifname':'lo'}]).encode(),json.dumps(route or []).encode()]):
   return m.require()
 def test_true_offline_systemd_container_is_accepted(self):self.assertEqual(self.check_local(),'disposable-docker')
 def test_network_interface_or_default_route_is_rejected(self):
  with self.assertRaisesRegex(RuntimeError,'OFFLINE_LOOPBACK_ONLY_REQUIRED'):self.check_local(links=[{'ifname':'eth0'}])
  with self.assertRaisesRegex(RuntimeError,'DEFAULT_ROUTE_FORBIDDEN'):self.check_local(route=[{'dst':'default'}])
 def test_private_run_mount_is_rejected_before_deployment(self):
  with self.assertRaisesRegex(RuntimeError,'RUN_MOUNT_PROPAGATION_REQUIRED'):self.check_local(shared=False)
 def test_fake_init_or_untrusted_marker_is_rejected(self):
  with self.assertRaisesRegex(RuntimeError,'REAL_SYSTEMD_REQUIRED'):self.check_local(pid='bash')
  with self.assertRaisesRegex(RuntimeError,'DISPOSABLE_MARKER_INVALID'):self.check_local(marker='untrusted')
  with self.assertRaisesRegex(RuntimeError,'DISPOSABLE_MARKER_INVALID'):self.check_local(mode=0o100666)
