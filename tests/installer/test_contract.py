import pathlib, unittest
ROOT=pathlib.Path(__file__).resolve().parents[2]
class InstallerContract(unittest.TestCase):
 def test_browser_processes_not_closed(self):
  iss=(ROOT/'installer/setup.iss').read_text()
  self.assertIn('CloseApplications=no',iss);self.assertIn('RestartApplications=no',iss)
  scripts='\n'.join(p.read_text(encoding='utf-8-sig') for p in (ROOT/'installer').glob('*.ps1'))
  for bad in ('Stop-Process','taskkill','chrome://quit','ssh ','scp ','tools/enroll','RotateCA'):self.assertNotIn(bad,scripts)
 def test_hash_ca_and_driver_pins(self):
  contract=(ROOT/'installer/package-contract.ps1').read_text()
  for expected in ('INSTALL_PACKAGE_HASH_MISMATCH','INSTALL_CA_PIN_MISMATCH','INSTALL_DRIVER_VERSION_UNAPPROVED','INSTALL_PACKAGE_REPARSE_POINT'):self.assertIn(expected,contract)
 def test_activation_only_preserves_owner_and_legacy_paths(self):
  script=(ROOT/'installer/install-package.ps1').read_text()
  for expected in ('ActivationOnly=$true','INSTALL_OWNER_MISMATCH','INSTALL_SAME_OWNER_ELEVATION_REQUIRED','update-service.ps1','stage-club-activation.ps1','INSTALL_SERVICE_HASH_MISMATCH'):self.assertIn(expected,script)
  self.assertNotIn('Remove-Item $installed',script)
 def test_no_secret_payload_and_no_reboot(self):
  build=(ROOT/'installer/build.ps1').read_text()
  self.assertIn('PUBLIC_TRUST_DIRECTORY_MUST_CONTAIN_ONLY_CA',build)
  self.assertNotIn('device.pem',build);self.assertNotIn('identity.json',build)
  self.assertIn('RestartIfNeededByRun=no',(ROOT/'installer/setup.iss').read_text())
