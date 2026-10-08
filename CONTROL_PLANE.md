# Control plane (Phase 5 design)
Go HTTP API + PostgreSQL, outside packet path. Planned endpoints:
/api/v1/auth, /devices, /relays, /config, /apps, /telemetry, /version.
Device key identity, locally verified short-lived relay tickets, signed versioned
configs with expiry and Last Known Good cache. No endpoints are implemented yet.
