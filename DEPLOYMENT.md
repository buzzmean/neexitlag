# Deployment
No production deployment is approved or performed. Build relay with Go 1.25+.
Dedicated unprivileged Linux account, UDP ingress and restricted egress are required.
Use per-device certificates and pinned CA; never disable certificate verification.
Update signing, Windows driver signing, staged rollout and rollback are future work.
