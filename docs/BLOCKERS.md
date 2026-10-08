# Phase 3 external blocker

The attached executor is Linux x86_64 and has no Windows runtime, Wine, MSVC/WDK
or /dev/kvm. No Windows runner credentials or attached Windows device are configured.
Command attempted: `cmd.exe /c ver`.
Exact error: `/bin/bash: line 4: cmd.exe: command not found`; exit code 127.

Required to continue Phase 3 and verify its Definition of Done:
a Windows 10/11 x64 VM or runner with administrator access, Go toolchain and
signed WinDivert binaries for the temporary backend. Production WFP backend also
requires Visual Studio Build Tools/WDK and a driver signing/test signing workflow.
The acceptance test must observe real Process A → RELAY and Process B → DIRECT
simultaneously, plus service-crash fail-open, reinjection safety and PID reuse.

A Go Windows cross-build only checks portable source compatibility. It cannot
validate process capture or substitute for this acceptance test. No Windows
interception backend is represented as implemented. Phases 4–10 remain outstanding.
