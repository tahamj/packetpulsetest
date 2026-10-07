# Appendices

## Appendix A — Every API route

All routes are under `/api/v1`. **Access**: *public* (no session), *session* (signed in), *member* (signed in **and** in an organisation), *superuser* (the platform operator), *api key* (a Results API key with `read:results`). **Licence**: refused while the organisation's licence is suspended, revoked or expired. **Audited**: recorded by the audit middleware from `packetpulsego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go`; sign-in, sign-out and failed sign-ins are recorded by the user service itself (§2.5). `./packetpulsetest.sh docs` derives this table from the route handlers and fails if it drifts.

| Method | Path | Access | Capability | Licence | Audited |
|---|---|---|---|:---:|:---:|
| `POST` | `/apikey/add` | member | `apikey_manage` | required | ✅ |
| `GET` | `/apikey/list` | member | `apikey_manage` | — | — |
| `DELETE` | `/apikey/{credentialId}` | member | `apikey_manage` | — | ✅ |
| `GET` | `/auditlog/list` | session | `auditlog_view` | — | — |
| `GET` | `/auditlog/verify` | session | `auditlog_view` | — | — |
| `POST` | `/diagnostic/clientobservation` | member | `diagnostic_run` | required | ✅ |
| `GET` | `/diagnostic/list` | member | `diagnostic_view_own` or `diagnostic_view_all` | — | — |
| `POST` | `/diagnostic/submit` | member | `diagnostic_run` | required | ✅ |
| `GET` | `/diagnostic/tt/{ttNumber}` | member | `diagnostic_view_own` or `diagnostic_view_all` | — | — |
| `GET` | `/diagnostic/{requestId}` | member | `diagnostic_view_own` or `diagnostic_view_all` | — | — |
| `GET` | `/diagnostic/{requestId}/report.csv` | member | `diagnostic_view_own` or `diagnostic_view_all` and `report_export` | — | ✅ |
| `GET` | `/diagnostic/{requestId}/report.pdf` | member | `diagnostic_view_own` or `diagnostic_view_all` and `report_export` | — | ✅ |
| `POST` | `/dnssite/add` | member | `dns_site_manage` | required | ✅ |
| `POST` | `/dnssite/bulkimport` | member | `dns_site_manage` | required | ✅ |
| `POST` | `/dnssite/importfile` | member | `dns_site_manage` | required | ✅ |
| `GET` | `/dnssite/list` | member | `dns_site_view` | — | — |
| `DELETE` | `/dnssite/{dnsSiteId}` | member | `dns_site_manage` | — | ✅ |
| `PUT` | `/dnssite/{dnsSiteId}` | member | `dns_site_manage` | required | ✅ |
| `GET` | `/export/target` | member | `apikey_manage` | — | — |
| `PUT` | `/export/target` | member | `apikey_manage` | — | ✅ |
| `POST` | `/export/target/test` | member | `apikey_manage` | — | ✅ |
| `GET` | `/init` | public | — | — | — |
| `GET` | `/init/help/{screenId}` | public | — | — | — |
| `GET` | `/init/language/list` | public | — | — | — |
| `GET` | `/ldap/config` | member | `ldap_manage` | — | — |
| `PUT` | `/ldap/config` | member | `ldap_manage` | — | ✅ |
| `POST` | `/ldap/config/test` | member | `ldap_manage` | — | — |
| `POST` | `/ldap/groupmap/add` | member | `ldap_manage` | — | ✅ |
| `GET` | `/ldap/groupmap/list` | member | `ldap_manage` | — | — |
| `DELETE` | `/ldap/groupmap/{mapId}` | member | `ldap_manage` | — | ✅ |
| `GET` | `/licence/my` | member | `licence_view` | — | — |
| `GET` | `/monitor/alert/list` | member | `sla_view` | — | — |
| `POST` | `/monitor/channel/add` | member | `sla_manage` | required | ✅ |
| `GET` | `/monitor/channel/list` | member | `sla_view` | — | — |
| `DELETE` | `/monitor/channel/{channelId}` | member | `sla_manage` | — | ✅ |
| `POST` | `/monitor/channel/{channelId}/test` | member | `sla_manage` | required | ✅ |
| `POST` | `/monitor/maintenance/add` | member | `schedule_manage` | required | ✅ |
| `GET` | `/monitor/maintenance/list` | member | `sla_view` | — | — |
| `DELETE` | `/monitor/maintenance/{windowId}` | member | `schedule_manage` | required | ✅ |
| `PUT` | `/monitor/maintenance/{windowId}` | member | `schedule_manage` | required | ✅ |
| `POST` | `/monitor/schedule/add` | member | `schedule_manage` | required | ✅ |
| `GET` | `/monitor/schedule/list` | member | `schedule_manage` | — | — |
| `DELETE` | `/monitor/schedule/{scheduleId}` | member | `schedule_manage` | — | ✅ |
| `PUT` | `/monitor/schedule/{scheduleId}` | member | `schedule_manage` | required | ✅ |
| `POST` | `/monitor/schedule/{scheduleId}/run` | member | `schedule_manage` | required | ✅ |
| `POST` | `/monitor/sla/add` | member | `sla_manage` | required | ✅ |
| `GET` | `/monitor/sla/list` | member | `sla_view` | — | — |
| `DELETE` | `/monitor/sla/{slaPolicyId}` | member | `sla_manage` | — | ✅ |
| `PUT` | `/monitor/sla/{slaPolicyId}` | member | `sla_manage` | required | ✅ |
| `GET` | `/monitor/slareport` | member | `report_view` | — | — |
| `GET` | `/monitor/slareport.pdf` | member | `report_export` | — | — |
| `GET` | `/monitor/trend` | member | `report_view` | — | — |
| `GET` | `/organisation/brand` | member | — | — | — |
| `DELETE` | `/organisation/brand/{part}` | member | `staff_manage` | — | ✅ |
| `PUT` | `/organisation/brand/{part}` | member | `staff_manage` | — | ✅ |
| `GET` | `/organisation/settings` | member | — | — | — |
| `PUT` | `/organisation/settings` | member | `staff_manage` | — | ✅ |
| `GET` | `/ping/dashboard` | member | `report_view` | — | — |
| `POST` | `/ping/run` | member | `diagnostic_run` | required | — |
| `GET` | `/ping/run/{runId}` | member | `diagnostic_view_all` | — | — |
| `GET` | `/ping/run/{runId}/report.pdf` | member | `report_export` | — | — |
| `GET` | `/ping/runlist` | member | `diagnostic_view_all` | — | — |
| `GET` | `/platform/currency/list` | superuser | — | — | — |
| `POST` | `/platform/licence/issue` | superuser | — | — | ✅ |
| `GET` | `/platform/licence/list` | superuser | — | — | — |
| `GET` | `/platform/licence/{licenceId}/event/list` | superuser | — | — | — |
| `POST` | `/platform/licence/{licenceId}/file` | superuser | — | — | ✅ |
| `POST` | `/platform/licence/{licenceId}/renew` | superuser | — | — | ✅ |
| `POST` | `/platform/licence/{licenceId}/resume` | superuser | — | — | ✅ |
| `POST` | `/platform/licence/{licenceId}/revoke` | superuser | — | — | ✅ |
| `PUT` | `/platform/licence/{licenceId}/seats` | superuser | — | — | ✅ |
| `POST` | `/platform/licence/{licenceId}/suspend` | superuser | — | — | ✅ |
| `POST` | `/platform/organisation/add` | superuser | — | — | ✅ |
| `GET` | `/platform/organisation/list` | superuser | — | — | — |
| `GET` | `/platform/organisation/{organisationId}/licence` | superuser | — | — | — |
| `GET` | `/platform/plan/list` | superuser | — | — | — |
| `GET` | `/result/bytt/{ttNumber}` | api key | — | — | — |
| `GET` | `/result/export.csv` | api key | — | — | — |
| `GET` | `/sms/gateway` | member | `staff_manage` | — | — |
| `PUT` | `/sms/gateway` | member | `staff_manage` | — | ✅ |
| `POST` | `/sms/gateway/test` | member | `staff_manage` | — | ✅ |
| `GET` | `/staff/checkin/list` | member | `staff_manage` | — | — |
| `GET` | `/staff/list` | member | `staff_manage` | — | — |
| `POST` | `/staff/role/add` | member | `acl_manage` | — | ✅ |
| `GET` | `/staff/role/list` | member | `acl_manage` | — | — |
| `DELETE` | `/staff/role/{roleId}` | member | `acl_manage` | — | ✅ |
| `PUT` | `/staff/role/{roleId}` | member | `acl_manage` | — | ✅ |
| `GET` | `/staff/session/list` | session | — | — | — |
| `GET` | `/staff/session/organisation` | member | `staff_manage` | — | — |
| `DELETE` | `/staff/session/organisation/{sessionId}` | member | `staff_manage` | — | ✅ |
| `DELETE` | `/staff/session/{sessionId}` | session | — | — | ✅ |
| `DELETE` | `/staff/{staffId}` | member | `staff_manage` | — | ✅ |
| `PUT` | `/staff/{staffId}` | member | `staff_manage` | — | ✅ |
| `PUT` | `/staff/{staffId}/access` | member | `acl_manage` | — | ✅ |
| `PUT` | `/staff/{staffId}/secondfactor` | member | `staff_manage` | — | ✅ |
| `POST` | `/user/add` | member | `user_manage` | required | ✅ |
| `POST` | `/user/appearance` | session | — | — | — |
| `GET` | `/user/capability/list` | session | — | — | — |
| `POST` | `/user/language` | session | — | — | — |
| `GET` | `/user/list` | member | `user_manage` | — | — |
| `GET` | `/user/me` | session | — | — | — |
| `POST` | `/user/password` | session | — | — | ✅ |
| `POST` | `/user/signin` | public | — | — | — |
| `POST` | `/user/signin/resend` | public | — | — | — |
| `POST` | `/user/signin/verify` | public | — | — | — |
| `POST` | `/user/signout` | session | — | — | — |
| `POST` | `/user/signup` | public | — | — | — |
| `PATCH` | `/user/{userId}` | member | `user_manage` | — | ✅ |

## Appendix B — Capabilities and the built-in roles

| Capability | What it allows | Administrator | NOC Engineer | Viewer |
|---|---|:---:|:---:|:---:|
| `dns_site_view` | View the DNS site inventory. | ✅ | ✅ | ✅ |
| `dns_site_manage` | Add, edit and remove DNS sites. | ✅ | — | — |
| `diagnostic_run` | Submit a diagnostic against a TT number. | ✅ | ✅ | — |
| `diagnostic_view_own` | View diagnostics this person submitted. | ✅ | ✅ | ✅ |
| `diagnostic_view_all` | View every diagnostic in the organisation. | ✅ | — | ✅ |
| `sla_view` | View SLA policies and breach status. | ✅ | ✅ | ✅ |
| `sla_manage` | Create and edit SLA policies and alert channels. | ✅ | — | — |
| `schedule_manage` | Create and edit scheduled monitoring. | ✅ | ✅ | — |
| `report_view` | View reports and analytics. | ✅ | ✅ | ✅ |
| `report_export` | Download PDF and CSV reports. | ✅ | ✅ | — |
| `user_manage` | Invite and deactivate user accounts. | ✅ | — | — |
| `staff_manage` | Edit staff records and assign roles. | ✅ | — | — |
| `acl_manage` | Change roles and permission grants. | ✅ | — | — |
| `auditlog_view` | Read the activity trail. | ✅ | — | — |
| `ldap_manage` | Configure directory integration. | ✅ | — | — |
| `apikey_manage` | Issue and revoke Results API keys, and set up the result export. | ✅ | — | — |
| `licence_view` | View the organisation's licence and seat usage. | ✅ | — | — |

Built-in roles are shared by every organisation and cannot be edited; an organisation adds its own (§2.2). The organisation **owner** and the **platform superuser** are not roles: the owner holds every capability in its organisation, and the superuser none — it administers the platform, not a tenant.

## Appendix C — Migrations

Embedded in the server binary and applied in order at boot (`packetpulsego/pkg/common/dbclient/migrations/`). They only ever move forward (§5.5).

| File | What it does |
|---|---|
| `0001_2026_09_30_initial_schema.sql` | PacketPulse initial schema. |
| `0002_2026_09_30_tenancy_staff_and_access.sql` | Tenancy, staff and the permission model. |
| `0003_2026_09_30_platform_licensing.sql` | Platform licensing: currencies, price book, licences and their evidentiary |
| `0004_2026_09_30_diagnostics_and_api_keys.sql` | The diagnostic request: a NOC engineer investigating a trouble ticket. |
| `0005_2026_09_30_telecom_sla_and_schedules.sql` | Telecom depth: path analysis, service-level targets, alerting and scheduled |
| `0006_2026_09_30_audit_trail.sql` | The activity trail. |
| `0007_2026_09_30_ldap_integration.sql` | Directory integration, per organisation. |
| `0008_2026_09_30_user_appearance.sql` | Appearance preferences, per user. |
| `0009_2026_09_30_holding_org_and_seats.sql` | Addenda 1-3: only a superuser creates organisations and issues licences, and |
| `0010_2026_10_01_client_observations.sql` | Client-side observations: a measurement taken by the customer's own device |
| `0011_2026_10_01_maintenance_and_alert_damping.sql` | Maintenance windows, and discipline for alerting. |
| `0012_2026_10_01_result_partitioning_and_retention.sql` | Monthly partitioning for the result tables, and a retention policy. |
| `0013_2026_10_01_fault_verdict.sql` | The fault verdict, stored with the result it describes. |
| `0014_2026_10_01_dns_site_endpoint_uniqueness.sql` | One monitored endpoint per organisation - by the endpoint actually probed. |
| `0015_2026_10_01_audit_trail_outlives_its_subjects.sql` | The audit trail must outlive what it describes, unchanged. |
| `0016_2026_10_02_alert_escalation.sql` | An alert that gets worse inside its cooldown is announced. |
| `0017_2026_10_02_licence_period.sql` | The period a licence's amount pays for. |
| `0018_2026_10_02_ldap_ca_certificate.sql` | The authority an organisation's directory certificate is checked against. |
| `0019_2026_10_03_two_step_sign_in.sql` | Two-step sign-in. |
| `0020_2026_10_03_session_checkin.sql` | Where people sign in and out. |
| `0021_2026_10_03_device_test.sql` | The device test: what it tests against, and how fast the line was. |
| `0022_2026_10_03_loss_and_jitter.sql` | Loss and jitter on the ticket, not only on each site. |
| `0023_2026_10_04_dual_stack_results.sql` | Dual-stack sites, measured on both families and reported separately. |
| `0024_2026_10_04_result_export.sql` | Results delivered to an organisation's own server, on a schedule, and |
| `0025_2026_10_05_otp_only_and_licence_files.sql` | Sign-in by one-time code only, and licences from signed files. |
| `0026_2026_10_05_own_results_places_and_indexes.sql` | Whose tests a person reads, where each test was run, and the indexes the busiest reads need. |
| `0027_2026_10_05_path_analysis_removed.sql` | Path analysis is removed: hops and fault verdicts are no longer written. |
| `0028_2026_10_05_rename_the_partition_function.sql` | The partition function under its PacketPulse name, on every database. |
| `0029_2026_10_07_endpoint_location_and_report_branding.sql` | Where each endpoint is, where each test was run, how a test's endpoints were chosen, and an organisation's own logo and icon. |

## Appendix D — Settings

Read from the environment, falling back to `.env` (see `.env.example`); defined in `packetpulsego/pkg/common/config/Config.go`.

| Setting | Default | What it decides |
|---|---|---|
| `APP_ENV` | `development` | `production` refuses the settings in Part IV `NFR-018` |
| `APP_NAME` | `PacketPulse` | The name in reports and the user agent |
| `PORT` | `8080` | Where the API listens |
| `DATABASE_URL` | local `packetpulsedb` | The database (through PgBouncer on the live host) |
| `JWT_SECRET` | — (required, 32+ characters) | Signs sessions **and** seals every stored secret: rotating it signs everyone out and makes stored SMS tokens, bind passwords and export credentials unreadable until re-entered |
| `JWT_TTL` | `12h` | How long a session lasts |
| `CORS_ORIGINS` | `*` | Must be explicit origins in production |
| `TRUSTED_PROXIES` | `127.0.0.0/8, ::1/128` | Whose `X-Forwarded-For` is believed — the client address in sessions, limits and the trail |
| `AUTH_ATTEMPTS_PER_MINUTE` | `10` | Password-step requests per client address |
| `SECOND_STEP_ATTEMPTS_PER_MINUTE` | `30` | Second-step requests per client address |
| `ALLOW_SIGNUP` | `true` | Whether the owner named in a licence may sign up on this server |
| `OWNER_PASSWORD`, `OWNER_NAME` | — | The platform superuser (`superuser@rummaan53.com`, fixed in code), created at boot on the console only |
| `LICENCE_DIR` | — | The folder of licence files this server signs organisations in under, read at every sign-in |
| `LICENCE_SIGNING_KEY` | — | Makes this server the PacketPulse console, which issues licences. Only the vendor's own server has it |
| `LICENCE_PUBLIC_KEY` | — | Replaces the vendor key licences are checked against — **development builds outside production only** |
| `SMTP_HOST`, `SMTP_PORT` | `smtp.gmail.com`, `587` | Where sign-in codes are emailed from |
| `SMTP_FROM_n`, `SMTP_PASSWORD_n` (n = 0–3) | — | Gmail accounts (with app passwords) that send codes, spread across to stay inside Gmail's daily cap. Production needs at least one |
| `OTP_OUTBOX_FILE` | — | Writes codes to a file instead of sending them, for tests — **refused in production** |
| `SEED_DEMO`, `DEMO_EMAIL`, `DEMO_PASSWORD`, `DEMO_NAME`, `DEMO_ORGANISATION_NAME` | `false`, … | The demo organisation, its logins, sites and licence — **console only** |
| `DEMO_SMS_CODE`, `DEMO_SEATS` | `111111`, `5` | The demo's on-screen code, and how many demo walkthroughs may run at once — at least 5; a session unused for 30 minutes gives up its place |
| `DEMO_ORGANISATION_CODE` | — | Gives an organisation that already exists, by its licence's code, the demo sign-in: the on-screen code instead of an emailed or texted one, no password changes, the `DEMO_SEATS` cap, and public-only sweeps. Seeds nothing. **Console only**, and not with `SEED_DEMO` |
| `TRIAL_DAYS`, `TRIAL_SEATS`, `TRIAL_SITES` | `14`, `5`, `25` | A trial licence's terms |
| `PING_COUNT`, `PING_TIMEOUT`, `PING_INTERVAL`, `PING_CONCURRENCY` | `4`, `5s`, `200ms`, `16` | The probe engine's defaults |
| `PING_TCP_FALLBACK` | `true` | TCP connect where ICMP cannot be sent |
| `PING_ALLOW_PRIVATE` | `true` | Whether RFC 1918 and unique-local addresses may be probed |
| `LDAP_ALLOW_LOOPBACK` | `false` | A directory on PacketPulse's own host (development only) |
| `EXPORT_ALLOW_LOOPBACK` | `false` | A result-export server on PacketPulse's own host (development only) |
| `REVERSE_GEOCODING` | `auto` | `off` stops addresses being looked up for sign-in positions |
| `GOOGLE_MAPS_API_KEY` | — | Google Geocoding; without it, OpenStreetMap Nominatim |
| `GEOCODE_CONTACT` | `https://packetpulse.rummaan53.com` | Sent to Nominatim, which requires a contact |

## Appendix E — Response codes and actions

Every refusal is a JSON envelope: `{"error": {"code", "message", "request_id", "details", "action"}}`. The app shows `message`, names `details` against their fields, and acts on `action`.

| `code` | Status | Meaning |
|---|:---:|---|
| `bad_request` | 400 | Malformed input — an id that is not a UUID, an unreadable time, no organisation |
| `validation_failed` | 422 | Fields to correct, each named in `details` |
| `unauthorized` | 401 | No session or key, or one that has ended |
| `forbidden` | 403 | A capability the caller lacks |
| `not_found` | 404 | Absent — or another organisation's, which is the same thing to the caller |
| `conflict` | 409 | The state forbids it — e.g. a diagnostic with no results to report, a changed server key |
| `rate_limited` | 429 | Too many attempts from this address, or texts for this person |
| `internal_error` | 500 | Logged with the `request_id`; never carries the cause |
| `licence_expired`, `licence_suspended` | 403 | The organisation's licence stops this |
| `seat_limit_reached` | 403 / 409 | The licence covers no more people (adding someone), or fewer than are on it (signing in) |
| `invalid_code` | 401 | A wrong second-step code, with attempts left |
| `challenge_expired` | 401 | The sign-in must start again |
| `locked_out` | 429 | Too many wrong codes; wait |
| `location_required` | 403 | The organisation requires a position to sign in |

| `action` | What the app does |
|---|---|
| `retry` | Offers **Try again** |
| `renew` | Names who can renew the licence |
| `ask_administrator` | Says to ask an administrator |
| `restart_sign_in` | Returns to the password step |
| `wait` | Says when to try again |
| `share_location` | Explains how to allow location and try again |

## Appendix F — Glossary

| Term | Meaning |
|---|---|
| **Dual-stack** | A site reachable over both IPv4 and IPv6 |
| **Licence file** | The signed file in `LICENCE_DIR` that lets an organisation sign in on a server |
| **MOS** | Mean Opinion Score — estimated call quality, 1–5 (ITU-T G.107 E-model) |
| **Report only** | A result shown and kept but not counted: the IPv6 half of a dual-stack site |
| **RFC 3550 jitter** | Interarrival jitter: the smoothed difference between consecutive round trips |
| **Sweep** | One run of probes over a set of sites |
| **Sign-in code** | The 6-digit, ten-minute code emailed or texted for the second step |
| **TT number** | A trouble ticket number from the operator's own ticketing system |
| **Watermark** | How far the result export has delivered: results of diagnostics finished after it are still to send |

---

*Guide version `v2026.10-PROD-v2` — verified against source on 2026-10-05, and continuously re-verified by `./packetpulsetest.sh docs`.*
