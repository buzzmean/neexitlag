import importlib.util
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
DEPLOY = ROOT / 'deploy/vps'
spec = importlib.util.spec_from_file_location('relay_env', DEPLOY / 'validate-env.py')
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class RelayLifecycleTests(unittest.TestCase):
    def valid(self):
        return (DEPLOY / 'example.env').read_text().replace('REPLACE_TARGET_IPV4', '2.56.177.4')

    def validate(self, text):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'relay.env'
            path.write_text(text)
            return module.validate(path)

    def test_safe_environment_and_crlf_input(self):
        self.assertEqual(self.validate(self.valid())['RELAY_ID'], 'VPS-01')
        self.assertEqual(self.validate(self.valid().replace('\n', '\r\n'))['RELAY_LISTEN'], '0.0.0.0:4433')
        for old, new in [('127.0.0.1:9090', '0.0.0.0:9090'),
                         ('https://127.0.0.1:8443', 'http://127.0.0.1:8443'),
                         ('2.56.177.4/32', '0.0.0.0/0'),
                         ('VPS-01', '$(touch /tmp/unsafe)'),
                         ('0.0.0.0:4433', '0.0.0.0:443')]:
            with self.subTest(new=new), self.assertRaises(ValueError):
                self.validate(self.valid().replace(old, new))
        with self.assertRaises(ValueError):
            self.validate(self.valid() + '\nRELAY_ID=duplicate\n')
        with self.assertRaises(ValueError):
            self.validate(self.valid() + '\nNEEXITLAG_ADMIN_TOKEN=not-supported\n')

    def test_all_shell_scripts_parse(self):
        for path in DEPLOY.glob('*.sh'):
            subprocess.run(['bash', '-n', str(path)], check=True, capture_output=True)

    def test_systemd_hardening_and_no_open_proxy_start(self):
        # Exercise the actual start wrapper with a stub executable in a private
        # path: quoted settings must reach argv without shell interpretation.
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            binary = base / 'relay'
            binary.write_text('#!/usr/bin/env python3\nimport json,sys\nprint(json.dumps(sys.argv[1:]))\n')
            binary.chmod(0o755)
            start = base / 'relay-start.sh'
            start.write_text((DEPLOY / 'relay-start.sh').read_text().replace('/opt/neexitlag/bin/relay', str(binary)))
            import os, json
            env = dict(os.environ, RELAY_ALLOW_CIDRS='2.56.177.4/32', RELAY_ID='VPS-01')
            output = subprocess.check_output(['bash', str(start)], env=env)
            args = json.loads(output)
            self.assertEqual(args[args.index('-relay-id') + 1], 'VPS-01')
            self.assertIn('-ticket-keys', args)
            self.assertIn('-ca', args)
            self.assertNotIn('-allow-http-development', args)
            env.pop('RELAY_ALLOW_CIDRS')
            env.pop('VPS_PUBLIC_IP', None)
            rejected = subprocess.run(['bash', str(start)], env=env, capture_output=True)
            self.assertNotEqual(rejected.returncode, 0)
            self.assertIn(b'Explicit relay destination', rejected.stderr)
        unit = (DEPLOY / 'neexitlag-relay.service').read_text()
        for directive in ['User=neexitlag-relay', 'NoNewPrivileges=yes', 'ProtectSystem=strict',
                          'StandardOutput=journal', 'Restart=on-failure', 'TimeoutStopSec=10']:
            self.assertIn(directive, unit)

    def test_atomic_update_health_rollback_and_preserving_uninstall(self):
        # Execute the actual shell flow in isolated paths. Only service/health/
        # timing are stubbed; real ELF staging, rename, rollback and cleanup run.
        import os, shutil
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            scripts = base / 'scripts'
            scripts.mkdir()
            replacements = {p: str(base / p.lstrip('/')) for p in
                            ['/etc/neexitlag', '/opt/neexitlag', '/var/lib/neexitlag', '/etc/systemd/system']}
            for name in ['update.sh', 'uninstall.sh', 'relay-start.sh', 'neexitlag-relay.service', 'validate-env.py']:
                text = (DEPLOY / name).read_text()
                for old, new in replacements.items():
                    text = text.replace(old, new)
                text = text.replace('[[ $EUID == 0', '[[ 1 == 1')
                (scripts / name).write_text(text)
            binary = base / 'opt/neexitlag/bin/relay'
            binary.parent.mkdir(parents=True)
            shutil.copy('/bin/true', binary)
            unit_dir = base / 'etc/systemd/system'
            unit_dir.mkdir(parents=True)
            credentials = base / 'etc/neexitlag/relay'
            credentials.mkdir(parents=True)
            trust = credentials / 'trust.json'
            trust.write_text('fixture-public-trust')
            state = base / 'var/lib/neexitlag'
            state.mkdir(parents=True)
            replay = state / 'replay.json'
            replay.write_text('preserve-replay-fixture')
            stubs = base / 'stubs'
            stubs.mkdir()
            for name, body in {
                'systemctl': 'exit 0',
                'sleep': 'exit 0',
                'id': 'exit 1',
                'curl': 'if cmp -s "$TEST_RELAY_BINARY" /bin/false; then exit 1; fi; printf "neexitlag_relay_health 1\\n"'
            }.items():
                stub = stubs / name
                stub.write_text('#!/bin/sh\n' + body + '\n')
                stub.chmod(0o755)
            env = dict(os.environ, PATH=str(stubs) + ':' + os.environ['PATH'], TEST_RELAY_BINARY=str(binary))
            def invoke(name, *args):
                return subprocess.run(['bash', str(scripts / name), *args], env=env, capture_output=True)
            for _ in range(2):
                updated = invoke('update.sh', '/bin/true')
                self.assertEqual(updated.returncode, 0, updated.stderr.decode())
            failed = invoke('update.sh', '/bin/false')
            self.assertNotEqual(failed.returncode, 0)
            self.assertEqual(binary.read_bytes(), Path('/bin/true').read_bytes())
            self.assertEqual(trust.read_text(), 'fixture-public-trust')
            self.assertEqual(replay.read_text(), 'preserve-replay-fixture')
            invalid = base / 'not-an-elf'
            invalid.write_text('not a binary')
            self.assertNotEqual(invoke('update.sh', str(invalid)).returncode, 0)
            self.assertEqual(binary.read_bytes(), Path('/bin/true').read_bytes())
            self.assertEqual(invoke('uninstall.sh').returncode, 0)
            self.assertFalse(binary.exists())
            self.assertTrue(trust.exists())
            self.assertTrue(replay.exists())
            self.assertEqual(invoke('uninstall.sh', '--purge-relay-data').returncode, 0)
            self.assertFalse(credentials.exists())
            self.assertFalse(replay.exists())


if __name__ == '__main__':
    unittest.main()
