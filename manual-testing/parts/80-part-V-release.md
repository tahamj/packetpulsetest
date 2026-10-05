# Part V — Release Governance

How a release is judged ready, shipped, checked and — if it must be — rolled back.

## 5.1 The automated gate

A release candidate passes all of these, with **no skips**, before anything ships:

```bash
PACKETPULSE_TEST_URL=http://localhost:18080 PACKETPULSE_TEST_REQUIRE_SUPERUSER=1 ./packetpulsetest.sh
./packetpulsetest.sh load          # before a release that touches sweeps, sign-in or exports
```

`./packetpulsetest.sh` runs `tenancy assignment acl audit translation contract monitor unit client backup docs`. Coverage floors are a ratchet (`scripts/PacketPulseCoverageFloor.py`, `scripts/PacketPulseFlutterCoverageFloor.py`): raised when real coverage rises, never lowered to make a build pass.

## 5.2 What blocks a release

| Severity | Definition | Examples |
|---|---|---|
| **P0 — blocks** | Any 🔒 INVARIANT or 🛑 MUST NOT HAPPEN observed | One organisation sees another's data; a probe reaches loopback or metadata; a session survives being disabled; a secret is returned; the chain verifies after tampering; results lost after a sweep |
| **P1 — blocks unless waived in writing** | A documented flow cannot be completed | Sign-in impossible for a role; an export cannot be downloaded; a screen shows its empty state under a refusal |
| **P2 — fix next release** | Wrong but recoverable | A misaligned column at one width; an unclear message |

## 5.3 Go / No-Go checklist

| ID | Check | Owner |
|---|---|---|
| `REL-001` | `./packetpulsetest.sh` green with no skips, on the commit being shipped | QA |
| `REL-002` | Every row of *What changed* (front matter) manually verified: `AUTH-001`, `AUTH-009`, `CHK-001`, `DEV-002`, `DEV-005`, `V6-001`, `CSV-001`, `EXP-006`, `AUD-003` | QA |
| `REL-003` | Journeys `JRN-001`, `JRN-003`, `JRN-004`, `JRN-006` on a disposable stack | QA |
| `REL-004` | The live `.env` has `APP_ENV=production`, an `SMTP_FROM_0` / `SMTP_PASSWORD_0` pair, and none of `OTP_OUTBOX_FILE`, `LICENCE_PUBLIC_KEY`, `EXPORT_ALLOW_LOOPBACK=true`, `LDAP_ALLOW_LOOPBACK=true` | Release owner |
| `REL-005` | New migrations read for what they change on live data (a migration that revokes sessions, rewrites figures or changes a key is announced to users) | Backend |
| `REL-006` | Public documents (`packetpulseweb/index.html`, `docs/*.html` and their PDFs) promise nothing the release does not do | Product |

## 5.4 Deploy and verify

```bash
scripts/deploy/PacketPulseDeploy.sh --host mshop.rummaan53.com
```

The script backs the live database up before the new release boots and migrates it. It does not install the nginx vhost: when `scripts/deploy/packetpulse-nginx.conf` changes, copy it to `/etc/nginx/conf.d/packetpulse.conf` on the host, run `sudo certbot --nginx -d packetpulse.rummaan53.com --reinstall` to put the TLS lines back, then `sudo nginx -t` and reload. Afterwards:

| ID | Check | Expected |
|---|---|---|
| `REL-007` | `https://packetpulse.rummaan53.com/healthz` and `/readyz` | 200 |
| `REL-008` | The live database's newest migration | The newest file in `packetpulsego/pkg/common/dbclient/migrations/` |
| `REL-009` | A new route answers an anonymous caller | 401, not 404 (e.g. `GET /api/v1/export/target`) |
| `REL-010` | `/` and `/app/` | Different documents; the app loads and signs in (`AUTH-002`) |
| `REL-011` | The server log since boot | No `level=ERROR` |
| `REL-012` | `/proc/sys/net/ipv4/ping_group_range` on the host | `0 2147483647` — unprivileged ICMP, so sweeps report `icmp` rather than falling back to TCP |
| `REL-013` | `curl -sI https://packetpulse.rummaan53.com/app/flutter_bootstrap.js`, then the `main.<hash>.dart.js` it names | The bootstrap: `Cache-Control: no-cache` and no `max-age`. The bundle: `immutable`. 🛑 **Must NOT** cache anything else under `/app/` - a returning browser then keeps the old client. In October 2026 a seven-day cache kept browsers on a client from before two-step sign-in, which dropped everyone back at sign-in with no message |
| `REL-014` | `curl -sI` on `/`, `/app/` and `/assets/packetpulse-site.css` | `X-Frame-Options: SAMEORIGIN` and `X-Content-Type-Options: nosniff` on each. 🛑 **Must NOT** be missing: an `add_header` inside a location drops every header set on the server block |

## 5.5 Rollback

Releases live in `/opt/packetpulse/releases/<version>`; the last three are kept. To roll back the API, point `current` at the previous release and restart:

```bash
ssh mshop.rummaan53.com 'ls -1t /opt/packetpulse/releases'
ssh mshop.rummaan53.com 'sudo ln -sfn /opt/packetpulse/releases/<previous> /opt/packetpulse/current && sudo systemctl restart packetpulse'
```

> [!WARNING]
> ⚠️ **TRAP** — migrations run forward only. Rolling the binary back does not roll the schema back; an older binary runs against a newer schema, which is safe only while every migration since added rather than removed. If a migration must be undone, restore the pre-deploy backup taken by the deploy (`scripts/deploy/backup/`), whose restore drill is part of the `backup` suite.

---
