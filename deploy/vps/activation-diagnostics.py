"""Bounded diagnostics: no exception text, arguments, source lines or locals escape."""
import errno
from pathlib import Path
import re
import shlex

STEP = 'ENTRY'
PROGRAMS = {'systemctl', 'systemd-run', 'runuser', 'openssl', 'pg_dump', 'pg_dumpall', 'psql', 'curl', 'ss', 'ip', 'tar', 'python3', 'fresh-pc', 'veliq-fresh-helper', 'relay', 'control-plane'}
CODES = {'ROOT_REQUIRED_OR_EXCLUDED_HOST', 'VPS_IDENTITY_MISMATCH', 'INVALID_BASELINE_COMMIT', 'READ_ONLY_OUTPUT_FILE_FORBIDDEN', 'READ_ONLY_RESPONSE_TOO_LARGE', 'READ_ONLY_COMMAND_FAILED', 'READ_ONLY_TIMER_LIST_TOO_LARGE', 'READ_ONLY_BACKUP_LIST_TOO_LARGE', 'READ_ONLY_BACKUP_METADATA_TOO_LARGE', 'BOUNDED_REQUEST_EXCEEDED', 'RELEASE_MANIFEST_INVALID', 'EXPLICIT_DEPLOYMENT_CONFIRMATION_REQUIRED', 'ARTIFACT_HASH_OR_SIZE_MISMATCH', 'ACTIVATION_RELEASE_FAILED_SEE_SAFE_CODE'}
MESSAGES = {'Unsupported EnvironmentFile syntax':'ENVIRONMENT_FILE_SYNTAX_UNSUPPORTED', 'Required EnvironmentFile setting absent/duplicated':'ENVIRONMENT_FILE_SETTING_ABSENT_OR_DUPLICATED', 'UNSUPPORTED_LOCAL_DATABASE':'UNSUPPORTED_LOCAL_DATABASE'}
FILES = {'activation-diagnostics.py', 'inspect-activation.py', 'fresh-vps-upgrade.py', 'production-upgrade.py', 'production-engine.py', 'activation-release-ssh.py', 'deploy-activation.py', '<readonly-engine>', '<readonly-fresh>', '<readonly-activation>', '<string>'}

def step(name):
    global STEP
    STEP = name

def safe_stderr(raw):
    text = raw.decode('utf-8', 'replace') if isinstance(raw, bytes) else str(raw)
    labels = [(r'Permission denied', 'PERMISSION_DENIED'), (r'No such file or directory', 'FILE_NOT_FOUND'), (r'not found', 'NOT_FOUND'), (r'Connection refused', 'CONNECTION_REFUSED'), (r'timed out', 'TIMED_OUT'), (r'certificate has expired', 'CERTIFICATE_EXPIRED'), (r'certificate verify failed', 'CERTIFICATE_VERIFY_FAILED'), (r'No space left on device', 'NO_SPACE_LEFT'), (r'Read-only file system', 'READ_ONLY_FILESYSTEM'), (r'relation .* does not exist', 'SQL_RELATION_MISSING'), (r'column .* does not exist', 'SQL_COLUMN_MISSING'), (r'authentication failed', 'SQL_AUTHENTICATION_FAILED'), (r'could not connect|connection .* failed', 'CONNECTION_FAILED'), (r'syntax error', 'SYNTAX_ERROR'), (r'must be owner|permission denied for', 'SQL_PERMISSION_DENIED')]
    found = [label for pattern, label in labels if re.search(pattern, text[:8192], re.I)]
    return '|'.join(found) if found else ('REDACTED_UNRECOGNIZED_STDERR' if text else 'EMPTY')

def failure(exc, stage=None):
    code = MESSAGES.get(str(exc), str(exc) if str(exc) in CODES else 'SAFE_OPERATION_FAILED')
    if isinstance(exc, OSError):
        code = {errno.ENOENT:'FILE_NOT_FOUND', errno.EACCES:'PERMISSION_DENIED', errno.ENOSPC:'NO_SPACE_LEFT'}.get(exc.errno, 'OS_OPERATION_FAILED')
    command = getattr(exc, 'command', '')
    exit_code = getattr(exc, 'exit_code', getattr(exc, 'returncode', None))
    frames = []
    tb = exc.__traceback__
    while tb is not None:
        frame = tb.tb_frame
        name = Path(frame.f_code.co_filename).name
        if name in FILES:
            frames.append({'file':name, 'line':tb.tb_lineno})
        tb = tb.tb_next
    return {'code':code, 'step':stage or STEP, 'exception':type(exc).__name__ if type(exc).__name__ in ('RuntimeError','UpgradeError','PermissionError','FileNotFoundError','ValueError','TimeoutExpired','OSError','CalledProcessError','SyntaxError','JSONDecodeError','UnicodeDecodeError') else 'OPERATION_ERROR', 'exit_code':exit_code if isinstance(exit_code,int) else None, 'command':command if command in PROGRAMS else 'NOT_REPORTED', 'stderr':safe_stderr(getattr(exc,'command_stderr',getattr(exc,'stderr',b''))), 'traceback':frames[-12:]}

def require_vps_identity(path, expected='2.56.177.4'):
    # EnvironmentFile syntax, never source/eval shell input. BOM/CRLF are text
    # encodings, not evidence of a different server. Duplicate values fail closed.
    values = []
    for line in Path(path).read_text(encoding='utf-8-sig').splitlines():
        if line.strip().startswith('VPS_PUBLIC_IP='):
            fields = shlex.split(line.strip().partition('=')[2], comments=False)
            if len(fields) != 1:
                raise RuntimeError('VPS_IDENTITY_MISMATCH')
            values.extend(fields)
    if values != [expected]:
        raise RuntimeError('VPS_IDENTITY_MISMATCH')
