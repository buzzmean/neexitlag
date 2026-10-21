#!/usr/bin/env python3
"""In-memory owner audit. No staging/locks/recovery/enroll/migrations or secrets.
A present-day snapshot cannot prove that every historical DB write was absent.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

import importlib.util as _diagnostic_import
try:
 import activation_diagnostics as diagnostics
except ModuleNotFoundError:
 _spec=_diagnostic_import.spec_from_file_location('activation_diagnostics',Path(__file__).with_name('activation-diagnostics.py'))
 diagnostics=_diagnostic_import.module_from_spec(_spec);_spec.loader.exec_module(diagnostics)

BASELINE='2c06089f61c7b7a27cdf575f5cea4b834992426d'

def inspect(fresh,engine,expected=BASELINE,root=Path('/'),host=None):
    root=Path(root)
    class ReadOnlyHost(engine.Host):
        def run(self,args,check=True,output=None,timeout=60):
            if output is not None:raise RuntimeError('READ_ONLY_OUTPUT_FILE_FORBIDDEN')
            program='psql' if args[:5]==['runuser','-u','postgres','--','psql'] else Path(args[0]).name
            diagnostics.step('READ_ONLY_COMMAND_'+(program.upper() if program in diagnostics.PROGRAMS else 'OTHER'))
            p=subprocess.run(args,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=timeout,check=False)
            if len(p.stdout)>2*engine.MAX_CONFIG:raise RuntimeError('READ_ONLY_RESPONSE_TOO_LARGE')
            if check and p.returncode:
                e=RuntimeError('READ_ONLY_COMMAND_FAILED');e.exit_code=p.returncode;e.command=program;e.command_stderr=p.stderr[:8192];raise e
            return p.returncode,p.stdout
        def sql(self,query,database='neexitlag'):
            # Server-enforced read-only transaction, no migrations/DDL/nonce proof.
            return self.run(['runuser','-u','postgres','--','psql','-X','-q','-A','-t','-v','ON_ERROR_STOP=1','-h','/var/run/postgresql','-p',str(self.db_port),'-d',database,'-c','BEGIN READ ONLY; '+query+'; COMMIT;'],timeout=5)[1].decode().strip()
        def configuration_owner(self,entry):return entry.st_uid==0
    h=host if host is not None else ReadOnlyHost(root)
    diagnostics.step('READ_ONLY_PREFLIGHT')
    report=fresh.report(h)
    diagnostics.step('READ_ONLY_BINARIES')
    binary=[]
    for role,service in [('control-plane','neexitlag-control-plane'),('relay','neexitlag-relay')]:
        pid=report['services'][service]['pid']
        installed=report['binary_sha256'][role]
        running=h.running_binary_hash(str(pid)) if pid else None
        binary.append({'role':role,'commit':report['capabilities'][role].get('commit'),'installed_sha256':installed,'running_sha256':running,'same_binary':running==installed})
    units=[]
    for name in ('neexitlag-control-plane','neexitlag-relay','neexitlag-echo','veliq-activation-recover','veliq-certificate-renewal'):
        service=name+'.service'
        drops=h.systemctl('show','--value','--property=DropInPaths',service,check=False).split()
        safe_drops=[Path(x).name if re.fullmatch('[A-Za-z0-9_.-]+',Path(x).name) else 'UNREVIEWED_PATH' for x in drops]
        units.append({'unit':service,'dropins':safe_drops,'active':h.systemctl('is-active',service,check=False),'enabled':h.systemctl('is-enabled',service,check=False),'load':h.systemctl('show','--value','--property=LoadState',service,check=False),'pid':h.systemctl('show','--value','--property=MainPID',service,check=False),'started':h.systemctl('show','--value','--property=ExecMainStartTimestamp',service,check=False)})
    _,timers=h.run(['systemctl','list-units','--all','--plain','--no-legend','--no-pager','--type=timer'],check=False)
    names=[line.split()[0] for line in timers.decode().splitlines() if line.split() and re.fullmatch(r'(veliq-(activation-rollback|leaf-rollback)-[0-9TZa-f-]+|veliq-certificate-renewal|neexitlag-upgrade-[0-9TZa-f-]+)\.timer',line.split()[0])]
    if len(names)>100:raise RuntimeError('READ_ONLY_TIMER_LIST_TOO_LARGE')
    timer_states=[{'unit':n,'active':h.systemctl('is-active',n,check=False),'enabled':h.systemctl('is-enabled',n,check=False)} for n in names]
    tables=int(h.sql("SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('activation_codes','activation_slots','activation_policy')",h.database))
    database={'schema_version':report['applied_schema_version'],'activation_tables':tables,'devices':int(h.sql('SELECT count(*) FROM devices',h.database)),'revoked':int(h.sql('SELECT count(*) FROM devices WHERE revoked',h.database)),'config_version':int(h.sql('SELECT coalesce(max(version),0) FROM config_versions',h.database)),'activation_migration_not_applied':report['applied_schema_version']==1 and tables==0,'historical_writes_unchanged':'NOT_PROVABLE_WITHOUT_BEFORE_SNAPSHOT'}
    diagnostics.step('READ_ONLY_BACKUPS')
    backups=[]
    for base in (root/'var/backups/neexitlag-activation',root/'var/backups/neexitlag-certificates'):
        if not base.exists():continue
        paths=sorted(base.iterdir())
        if len(paths)>200:raise RuntimeError('READ_ONLY_BACKUP_LIST_TOO_LARGE')
        for directory in paths:
            if directory.is_symlink() or not re.fullmatch(r'\d{8}T\d{6}Z-[a-f0-9]{8}',directory.name):continue
            state_path=directory/('state.json' if base.name.endswith('activation') else 'transaction.json')
            if state_path.exists() and state_path.stat().st_size>2*1024*1024:raise RuntimeError('READ_ONLY_BACKUP_METADATA_TOO_LARGE')
            state=json.loads(state_path.read_text()) if state_path.exists() else {}
            phase=state.get('phase','INCOMPLETE_BACKUP')
            if phase not in ('pending','prepared','committed','rolled-back','rolled_back','aborted','INCOMPLETE_BACKUP'):phase='UNKNOWN'
            item={'path':'/'+str(directory.relative_to(root)),'phase':phase,'binary_transaction_recorded':bool(state.get('binary_id'))}
            # File bytes/modes compared privately; never print secret-file hashes.
            same=True
            for absolute,meta in state.get('files',{}).items():
                # Only reviewed locations, never arbitrary backup-controlled paths.
                if not absolute.startswith(('/etc/neexitlag/','/opt/neexitlag/maintenance/','/etc/systemd/system/')):same=False;break
                live=root/absolute.lstrip('/')
                if meta is None:
                    if live.exists():same=False
                elif not live.is_file() or live.is_symlink() or hashlib.sha256(live.read_bytes()).hexdigest()!=meta.get('sha256'):same=False
            item['managed_files_match_backup']=same if state.get('files') else None
            backups.append(item)
    diagnostics.step('READ_ONLY_CA_METADATA')
    key=h.path('/etc/neexitlag/control/ca.key');runtime=h.path('/run/credentials/neexitlag-control-plane.service/veliq-device-ca-key')
    key_stat=key.stat()
    ca={'root_only':key_stat.st_uid==0 and key_stat.st_mode&0o777==0o600,'direct_service_readable':h.run(['runuser','-u','neexitlag-control','--','test','-r',str(key)],check=False)[0]==0,'runtime_copy_matches':runtime.is_file() and hashlib.sha256(runtime.read_bytes()).digest()==hashlib.sha256(key.read_bytes()).digest()}
    etc=root/'etc'
    older=sorted(etc.glob('neexitlag.bak-*')) if etc.exists() else []
    if len(older)>200:raise RuntimeError('READ_ONLY_BACKUP_LIST_TOO_LARGE')
    for directory in older:
        if directory.is_symlink() or not re.fullmatch(r'neexitlag.bak-\d{8}T\d{6}Z-[a-f0-9]{8,12}',directory.name):continue
        metadata=directory/'state.json'
        kind='binary_upgrade' if metadata.exists() else 'ca_rotation_or_legacy'
        if not metadata.exists():metadata=directory/'transaction.json'
        if metadata.exists() and metadata.stat().st_size>2*1024*1024:raise RuntimeError('READ_ONLY_BACKUP_METADATA_TOO_LARGE')
        value=json.loads(metadata.read_text()) if metadata.exists() else {}
        phase=value.get('phase','INCOMPLETE_BACKUP')
        if phase not in ('prepared','pending','pending-verification','committed','rolled_back','rolled-back','aborted','maintenance-compatibility','quiescing','migrating','installing-relay','installing-control-plane','INCOMPLETE_BACKUP'):phase='UNKNOWN'
        backups.append({'path':'/'+str(directory.relative_to(root)),'phase':phase,'backup_type':kind})
    baseline=all(x['commit']==expected and x['same_binary'] for x in binary) and database['schema_version']==1 and tables==0
    pending=any(x['phase'] in ('pending','pending-verification','maintenance-compatibility','quiescing','migrating','installing-relay','installing-control-plane') for x in backups)
    return {'read_only':True,'expected_baseline_commit':expected,'binary':binary,'services':report['services'],'units':units,'ports':report['ports'],'database':database,'ca_credential_dropin_exists':not report['ca_migration_required'],'ca_access':ca,'backups':backups,'timers':timer_states,'binaries_and_schema_at_baseline':baseline,'verdict':'PENDING_RECOVERY_REQUIRES_REVIEW' if pending else 'BASELINE_WITH_DEPLOYMENT_ARTIFACTS' if baseline and backups else 'BASELINE_CURRENT_STATE_ONLY' if baseline else 'BASELINE_DIFFERENT_REQUIRES_REVIEW','historical_no_changes_proven':False}

# Called by the small SSH console launcher, without writing this file on VPS.
def run(payload):
    diagnostics.step('READ_ONLY_REMOTE_IDENTITY')
    if os.geteuid()!=0 or os.uname().nodename.split('.')[0].lower()=='veliq-swe-01':raise RuntimeError('ROOT_REQUIRED_OR_EXCLUDED_HOST')
    diagnostics.require_vps_identity('/etc/neexitlag/public/relay.env')
    diagnostics.step('READ_ONLY_MODULES')
    import base64,types
    e=types.ModuleType('readonly_engine');e.__file__='<readonly-engine>'
    exec(compile(base64.b64decode(payload['engine'],validate=True),e.__file__,'exec'),e.__dict__)
    f=types.ModuleType('readonly_fresh');f.__file__='<readonly-fresh>';f.__dict__['_FRESH_ENGINE']=e
    exec(compile(base64.b64decode(payload['fresh'],validate=True),f.__file__,'exec'),f.__dict__)
    diagnostics.step('READ_ONLY_BASELINE')
    expected=payload.get('baseline',BASELINE)
    if not re.fullmatch('[a-f0-9]{40}',expected):raise RuntimeError('INVALID_BASELINE_COMMIT')
    return inspect(f,e,expected)
