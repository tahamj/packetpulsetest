<div align="center">

# 📡 PacketPulse Manual Testing Guide

### *& Master Knowledge Manual*

**The single source of truth for testing, understanding, selling, operating, and releasing PacketPulse.**

<br>

`v2026.10-PROD-v2`  ·  `Verified against source 2026-10-24`

<br>

| 🧭 Screens | 🔌 API routes | 🔐 Capabilities | 🗄️ Migrations | 💬 Catalogue strings | ❓ Help topics | 🧪 Guard suites |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **17** <br><sub>in the navigation rail</sub> | **100** <br><sub>under `/api/v1`</sub> | **16** <br><sub>3 built-in roles</sub> | **35** <br><sub>applied at boot</sub> | **1186** <br><sub>English, server-served</sub> | **20** <br><sub>one per screen</sub> | **11** <br><sub>+ `load`, opt-in</sub> |

<br>

*Quad-Lens Architecture — an **executable QA playbook**, an **onboarding & user manual**, a **commercial & sales pitch**, and a **developer release confidence guide**.*

<sub>🔒 Every number above is asserted against source by **`./packetpulsetest.sh docs`**. This guide is **generated** from `packetpulsetest/manual-testing/parts/` by `packetpulsetest/manual-testing/build/build_guide.py` — edit the parts, not the outputs.</sub>

</div>

---

## 🧭 How To Read This Manual

PacketPulse is a multi-tenant network-diagnostics service for telecom operators: a Go API, a Flutter client (web, desktop, phone) and PostgreSQL, live at **https://packetpulse.rummaan53.com**. It is structurally a server-side request forgery primitive with a user interface — an authenticated person names an address and the server sends traffic to it — so much of what follows is about what must **not** happen.

| If you are a… | Read, in this order | What you get |
|---|---|---|
| 💼 **Sales & account lead** | Part L → every chapter's **🌟 Commercial Presentation** → Journeys `JRN-001`, `JRN-004`, `JRN-006` | What to say, to whom, and the evidence behind each claim. |
| 👥 **Operator / NOC user** | Part L → every chapter's **📖 User Guide** → Journeys `JRN-002`, `JRN-003` | Which screen does what, what each status means, and how to recover from each refusal. |
| 🧪 **QA & field tester** | Part 0 → Part II test tables → Part III → Part IV → Part V §3 | Concrete steps, expected results, and the **🛑 Must NOT happen** assertions. |
| ⚙️ **Backend engineer** | Part 0 §0.3–§0.6 → each chapter's **⚙️ Developer Guide** → Appendix A–D | Routes, capabilities, migrations, invariants, and the suites that guard them. |
| 📱 **Flutter engineer** | Part 0 §0.4 → Part II → Part IV §4.1–§4.3 | Screens, states (skeleton, empty, error, 403), catalogue strings, and form-factor rules. |
| 🛡️ **Security reviewer** | Part IV §4.4 → `AUTH-*`, `AUD-*`, `EXP-*`, `SITE-*` cases → Appendix A | Tenancy, the destination policy, the activity trail, and secrets handling. |

<br>

### Reading conventions

| Badge | Meaning |
|:---:|---|
| ✅ **VERIFIED** | Checked against source on 2026-10-04. The file, route or constant is cited. |
| 🔒 **INVARIANT** | A guarantee of the system. Seeing it violated is a **P0 release blocker** (Part V §5.2). |
| 🛡️ **GUARD** | A server-side refusal, with its status code and envelope `code`. |
| ⚠️ **TRAP** | A place testers have drawn the wrong conclusion before. Read it before filing a bug. |
| 🌟 **SALES** | The customer value and how to say it. |
| 📖 **USER** | Which screen, which button, what it means. |
| ⚙️ **DEV** | Routes, tables, invariants and the automated coverage. |
| 🛑 **MUST NOT HAPPEN** | The negative assertion each test case exists to make. |

> [!TIP]
> Every manual test case carries a stable **Test ID** (`AUTH-004`, `JRN-003`, …). Quote it in bug reports and run records. IDs are **never reused** — a retired case leaves its number vacant — and `./packetpulsetest.sh docs` fails if two cases share one.

> [!CAUTION]
> **Never test against live customer data.** Use a local server with a disposable database (Part 0 §0.2), or the live demo organisation for read-only walkthroughs. Tenancy, ACL and audit cases need **two organisations** of your own — testing them from a single account proves nothing.

---

## 📋 What Changed in the 17 October 2026 Review

> [!IMPORTANT]
> **Deployed 2026-10-17: release 20261017-105318, and the reflector row in 20261017-190629.** The review of 17 October 2026 asked for PacketPulse to be able to help inside a customer's organisation, for the endpoint to be picked by position, and for an ended session to land on sign-in; then for PacketPulse's own reflector to answer every location. Where this table and an older one disagree, this one is the product.

| Change | What changed | Where to test | Source of truth |
|---|---|---|---|
| **PacketPulse support sign-in** | From **Platform**, the operator signs in to an organisation as **PacketPulse support** for one hour, giving a reason. Support is the organisation's own support account, its built-in Administrator: no one on its licence, not in its people list, never able to sign in by password. The organisation's activity trail names the operator and the reason; its administrators see the session under **Who is signed in** and can end it. A banner says so on every screen; **End support**, or the hour running out, returns to the console. | `PLAT-011`–`PLAT-018` | `packetpulsego/pkg/usermicroservice/userservice/UserSupport.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0034_2026_10_17_support_sign_in.sql` |
| **Every endpoint has a position** | **Latitude \*** and **Longitude \*** are required on the endpoint form, and for a file's new rows. Run diagnostic picks the endpoint nearest the device only when every enabled endpoint has a position; until then everyone chooses, and the page says why. **Bulk import** still adds plain addresses, which show *No position*. | `SITE-006`, `SITE-015`, `SITE-016`, `DEV-028` | `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteService.go`<br>`packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticEndpointField.dart` |
| **An ended session lands on sign-in** | When a session's twelve hours run out the app returns to sign-in then, not at the next click, and closes any dialog left open. A token the device thinks has expired is still sent, so the server's refusal is read as the session ending; sign-in says *Your session has ended. Sign in again to continue.* | `AUTH-014`, `AUTH-017`, `AUTH-018` | `packetpulseflutter/lib/main.dart`<br>`packetpulseflutter/lib/common/services/PacketPulseSessionStore.dart` |
| **PacketPulse's reflector answers every endpoint** | PacketPulse's own standalone reflector is its test reflector for every location: any endpoint, of any organisation, may name it as its **UDP reflector address** and its UDP test runs to it. It was installed with one endpoint's key and refused every other endpoint (*installed with a different key*), so a Bengaluru engineer, given the nearest endpoint, could not test to it. The server now signs for it with one network key (`REFLECTOR_NETWORK_HOSTS`); customer reflectors keep their endpoint keys. | `UDP-044`, `UDP-013` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticReflectorRelay.go`<br>`packetpulsego/cmd/packetpulseserver/ReflectorNetworkKey.go` |

---

## 📋 What Changed on 16 October 2026

> [!IMPORTANT]
> **Deployed 2026-10-16, release 20261016-100901.** Two gaps found while preparing the customer briefing, closed. Where this table and an older one disagree, this one is the product.

| Change | What changed | Where to test | Source of truth |
|---|---|---|---|
| **Change password on Settings** | **Settings → Change password** asks for the current password and the new one twice, and holds the server's rule (10 characters counted as characters, a letter and a number) before sending. A wrong current password is now a field error (422 with `details.current_password`), not a 401 the app would read as an ended session. Someone who signs in with their directory password is told to change it there and offered no form; the route refuses them with 409 `password_from_directory`. The owner keeps the form. | `SET-004`, `SET-006` | `packetpulseflutter/lib/common/presentation/PacketPulseSettingsScreen.dart`<br>`packetpulsego/pkg/usermicroservice/userservice/UserService.go` |
| **The product site fits a phone** | Card grids drop to one column, the footer links wrap, and the phone layout starts at 560 px, so neither `/` nor `/selftest.html` scrolls sideways from 360 px up. | `WEB-005` | `packetpulseweb/assets/packetpulse-site.css` |

---

## 📋 What Changed in the 13–14 October 2026 Review

> [!IMPORTANT]
> **Deployed 2026-10-14, release 20261014-103600.** The review of 13 October 2026 removed automatic monitoring, graded the device's own tests, and asked that every grade say why. Where this table and an older one disagree, this one is the product.

| Change | What changed | Where to test | Source of truth |
|---|---|---|---|
| **Automatic monitoring removed** | Schedules and their runner, alert channels and alerts, maintenance windows, the SLA report and its PDF, the dashboard's 30-day trend and the Monitoring screen are gone, with the `schedule_manage` capability; their routes answer 404. Nothing was dropped from the database. Policies no longer take damping fields. | `MON-011`, `MON-015`, `HIST-004` | `packetpulsego/pkg/pingmicroservice/pingservice/PingSlaEvaluator.go` |
| **Device tests graded** | A browser's UDP test and an app's ICMP echo are graded against the organisation's targets, and a breach is counted on the ticket; an HTTPS round-trip stays *Not graded*, with the reason under the results. | `MON-013`, `MON-014`, `UDP-031`, `DIAG-017` | `packetpulsego/pkg/pingmicroservice/pingservice/PingSlaEvaluator.go` |
| **Every grade against its targets** | Each graded figure is kept with its target, where Degraded begins and its verdict, and shown under the results, in the attach card and on the PDF. Loss to one decimal; charts and tiles coloured by their own figure. | `DIAG-018`, `DIAG-019` | `packetpulsego/pkg/common/dbclient/migrations/0033_2026_10_14_sla_checks.sql`<br>`packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go` |
| **UDP the browser's only method, at a chosen rate** | No *HTTPS to the endpoint*; **Packets per second** 10, 25, 50 or 100. The hosted reflector allows 120 a second and 6000 packets a session. | `UDP-016`, `UDP-029` | `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorSession.go` |
| **"UDP is blocked" only when it is** | A second, fresh connection before calling UDP blocked; *no UDP sent* and *could not be completed* told apart; the reflector logs how each session ended; a run records its real duration. | `UDP-030`, `UDP-032`, `UDP-033`, `UDP-034` | `packetpulseweb/assets/packetpulse-udp.js` |
| **Hidden, still measured** | No MOS and no standard deviation on any screen, report or the self-test; MOS is still graded and in the CSV. | `DEV-025`, `DEV-026` | `packetpulseflutter/lib/common/config/PacketPulseConfig.dart` |
| **The ticket, from where it was filed** | Notes in History and on the PDF; **View ticket** beside *Attached to ticket*; the results table fitted to its card; chosen chips readable; a read that got no answer is sent once more. | `DEV-024`, `DEV-027`, `DIAG-020`, `NFR-020` | `packetpulseflutter/lib/common/services/PacketPulseApiClient.dart` |

---

## 📋 What Changed in v2 — Licence Files, Emailed Codes and One Run Screen

> [!IMPORTANT]
> **Built 2026-10-05; not yet deployed.** The release waits for the `packetpulse.rummaan53.com` DNS record (Part V). Every row below must be verified before it is signed off. Where v2 and v1 disagree, v2 is the product.

| Change | What changed | Where to test | Source of truth |
|---|---|---|---|
| **Codes, not an authenticator** | The second step is a 6-digit code, **emailed**, or **texted** through the organisation's own SMS gateway. There is no authenticator app and there are no recovery codes. An owner's code is always emailed. | `AUTH-*`, `SMS-*` | `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0025_2026_10_05_otp_only_and_licence_files.sql` |
| **Signed licence files** | An organisation signs in only under a licence file signed by PacketPulse, in the server's `LICENCE_DIR` and checked at every sign-in. The superuser exists only on PacketPulse's own console. The first sign-up with the licence's owner address becomes the Administrator. | `PLAT-*`, `AUTH-*` | `packetpulsego/pkg/common/licence/Store.go` |
| **People, not sign-ins** | The licence counts the people on the books, switched on or off. Someone who leaves is switched off; someone with no records can be deleted, and stops counting. | `STF-*`, `WHO-*` | `packetpulsego/pkg/staffmicroservice/staffservice/StaffService.go` |
| **Own tests, and everyone's** | An engineer sees the tests they ran. An administrator sees everyone's, narrowed by person, place, status and dates. Each test records who ran it and where. | `HIST-*` | `packetpulsego/pkg/common/dbclient/migrations/0026_2026_10_05_own_results_places_and_indexes.sql` |
| **Path analysis removed** | No traceroute, no fault verdict, and no *Where the faults lay* card. Paths stored before stay in the database, unshown; the CSV keeps its `fault_class` column. | `DIAG-*` | `packetpulsego/pkg/common/dbclient/migrations/0027_2026_10_05_path_analysis_removed.sql` |
| **Speed test hidden** | **Measure speed** is no longer offered. A speed filed with a run before still shows with it. | `DEV-*` | `packetpulseflutter/lib/common/config/PacketPulseConfig.dart` |
| **One Run diagnostic page** | *Run diagnostic* and *Test from this device* are one page: the sweep **From the server** on top, the test **From this device** below, and one Customer ID, TT number and note for both. There is no mode switch (since 2026-10-06; v2 shipped two modes). Superseded on 2026-10-07: see the next row. | `DEV-*` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart` |
| **Device-test target kept internal** | The device test no longer names what it measures and offers no **Change target**; it still measures the organisation's host, or Cloudflare DNS when none is chosen. A filed device run reads *Device test*, with no address, in History, its PDF and every CSV; the stored row keeps the host. The target is set through `PUT /organisation/settings` (since 2026-10-06). | `DEV-001`, `DEV-002`, `DEV-006`, `DEV-007` | `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticdomain/shared/DeviceTarget.go` |
| **Run diagnostic from this device, to an endpoint** | The sweep **From the server** left the page; **Run diagnostic** is the test from this device. Customer ID, TT number and **Endpoint** are required before **Start test**; the endpoint nearest the device is chosen from each endpoint's latitude and longitude, an engineer cannot change it, and a manager can. The run measures that endpoint and is filed, and named, under it; the server refuses another organisation's (since 2026-10-07). | `DEV-017`–`DEV-022`, `DEV-006` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticEndpointField.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go` |
| **UDP test in Run diagnostic** | In a browser, **Run diagnostic** measures the ticket's endpoint with **UDP, like a call** — the only method offered there since 13 October 2026 — for 10, 30 or 60 seconds at 10, 25, 50 or 100 packets a second, grades it against the organisation's targets, and files where each lost packet went - up, down or late - with jitter each way, on the ticket, its page and its PDF. It measures to the reflector beside the endpoint when the endpoint names one (**UDP reflector address**, §3.1), and otherwise to the PacketPulse network, filed as *PacketPulse network*: the customer's access line, never the endpoint. The desktop and phone apps keep ICMP (since 2026-10-09). | `UDP-016`–`UDP-031`, `SITE-011`–`SITE-013` | `packetpulseflutter/lib/diagnosticmicroservice/service/udp/UdpProbeRun.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticUdpObservation.go` |
| **The demo's ticket filled in** | On the public demo, **Run diagnostic** opens with the Customer ID and TT number already filled in with the example the fields show (`CUST-88412`, `TT-2026-00731`); a walkthrough can change them. The session and `GET /user/me` say `is_demo`, true only for the demo organisation, so no other organisation's ticket is filled in (since 2026-10-09). | `DEMO-006` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart`<br>`packetpulsego/pkg/usermicroservice/userservice/UserService.go` |
| **DNS sites for endpoint managers** | **DNS sites** is in the rail only for whoever manages the endpoints (`dns_site_manage`). Engineers and Viewers no longer see the inventory; **Run diagnostic** still offers them the endpoints (since 2026-10-09). | `SITE-014` | `packetpulseflutter/lib/common/presentation/PacketPulseShell.dart` |
| **Device test, live** | **From this device** shows its IPv4 and IPv6 addresses and the nearest edge as three cards with an *IPv6 active* / *IPv4 only* badge (no ISP or AS number). **IP version** holds a run against the default target to IPv4 or IPv6. Eight figures, the round-trip chart with jitter, **Round-trips by band** and a **Probe log** fill in as each sample comes back; **Clear results** and **Clear log** empty them (since 2026-10-06). | `DEV-010`–`DEV-015` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientProbeCharts.dart`<br>`packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientProbeConsole.dart` |

---

## 📋 What Changed in v1 — the October 2026 Customer Release

> [!IMPORTANT]
> **Deployed to https://packetpulse.rummaan53.com on 2026-10-04.** It answers eight requirements from Northwind Telecom plus their remark that administrator logins be tracked. Every row below must be verified before a release is signed off.

| Requirement | What changed | Where to test | Source of truth |
|---|---|---|---|
| **1. Show IPv6** | Client IPs are read only from trusted proxies and shown canonically; a dual-stack site is measured over **both** families, the IPv6 result **reported, not counted**; the device test shows its IPv4 and IPv6 egress. | `V6-*`, `DEV-004` | `packetpulsego/pkg/common/probeguard/ProbeGuardPolicy.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql` |
| **2. Two-step sign-in** | Every sign-in has a second step. *v2 replaced the authenticator app and recovery codes with an emailed or texted code.* | `AUTH-*`, `SMS-*` | `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0019_2026_10_03_two_step_sign_in.sql` |
| **3. No target dropdown** | The device test measures the target the organisation chose (Cloudflare DNS by default), with one-tap presets for whoever may change it. *In v2 it is the From this device half of Run diagnostic, and since 2026-10-06 the target is not shown there.* | `DEV-*` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart` |
| **4. Location at check-in** | Sign-in and sign-out record the device's position for people set to *Record location*; field engineers on a separate **Check-ins** screen with address and map link; a per-organisation *Require location* rule. | `CHK-*`, `LOC-*` | `packetpulsego/pkg/common/dbclient/migrations/0020_2026_10_03_session_checkin.sql`<br>`packetpulsego/pkg/common/geocode/Geocode.go` |
| **5. Loss and jitter** | Loss and RFC 3550 jitter on every result, ticket, history row, dashboard row and PDF. | `DIAG-005`, `HIST-*`, `MON-*` | `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql` |
| **6. CSV, FTP and download** | **Export CSV** on every ticket; a Results API bulk pull; a scheduled push to the organisation's own **SFTP / FTPS / FTP** server with the SSH host key confirmed first. | `CSV-*`, `EXP-*` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`<br>`packetpulsego/pkg/exportmicroservice/exportservice/ExportService.go` |
| **7. English only** | The other 22 languages were removed; wording is still served by the server, in English. | `SET-003` | `scripts/intellicodegen/packetpulsestrings.py` |
| **8. Load testing** | A platform load suite: 20 people at once, a 200-site sweep, API and CSV pulls against p95 budgets. *The device line-speed test of v1 is hidden in v2.* | Part IV §4.5 | `packetpulsetest/golang/loadtest/load_test.go`<br>`.github/workflows/packetpulse-load.yml` |
| **Admin logins tracked** | Administrators' sign-ins and sign-outs in the **Activity** trail with device, browser, address and place; **failed** attempts on their accounts too, with the reason. | `AUD-*` | `packetpulsego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go` |

---
# Part 0 — Operational Prerequisites

Everything you need before the first test case: where to run PacketPulse, how to sign in to it, and how the automated gate relates to the manual one.

## 0.1 Environments

| Environment | Address | Database | Use it for | Never use it for |
|---|---|---|---|---|
| **Local development** | `http://localhost:8080` (API) · `flutter run -d chrome` (app) | `packetpulsedb` on the local PostgreSQL | Building and trying features | Destructive cases you cannot undo by hand |
| **Disposable test** | any free port, e.g. `http://localhost:18080` | a database you create for the run (e.g. `packetpulse_manual_test`) | **Every manual test pass** and the integration suites | — |
| **Live** | **https://packetpulse.rummaan53.com** (site `/`, app `/app/`, API `/api/`) | `packetpulsedb` on the Oracle host, behind PgBouncer | Read-only walkthroughs in the demo organisation; the post-deploy checks of Part V §5.4 | Tenancy, ACL, lockout or delete cases; anything that changes a real customer's data |

> [!WARNING]
> ⚠️ **TRAP — the `.env` fallback.** The Go database tests read `DATABASE_URL`, and when it is unset they fall back to the repository's `.env` — your development database — and **apply every migration to it on first connect**. Always export a disposable `DATABASE_URL` before `go test`, or run tests only through `./packetpulsetest.sh`, which does.

## 0.2 Toolchain

| Tool | Version | Why |
|---|---|---|
| Go | 1.26 | `packetpulsego/` and the integration suites in `packetpulsetest/golang/` |
| Flutter | 3.35 | `packetpulseflutter/` — web, macOS, Windows, Android, iOS |
| PostgreSQL | 18 (client tools too) | The database, and `pg_dump` for the backup suite |
| Python | 3.9+ (3.13 recommended) | `scripts/intellicodegen/packetpulsestrings.py`, the coverage floors, this guide's generator |
| gpg, rsync | any recent | The backup suite (`scripts/deploy/backup/`) |
| govulncheck | latest | Part of the `unit` suite |
| Node | 18+ | The public self-test page's probe engine tests |

## 0.3 Starting a disposable stack

```bash
# 1. a database of your own
createdb packetpulse_manual_test

# 2. the API, migrating that database at boot (migrations are embedded in the binary)
cd packetpulsego
DATABASE_URL=postgres://packetpulse:…@localhost:5432/packetpulse_manual_test?sslmode=disable \
PORT=18080 APP_ENV=test REVERSE_GEOCODING=off go run ./cmd/packetpulseserver

# 3. the app, pointed at it
cd ../packetpulseflutter
flutter run -d chrome --dart-define=PACKETPULSE_API_BASE_URL=http://localhost:18080
```

`REVERSE_GEOCODING=off` stops sign-in positions being sent to a map service from a test run. Every other setting comes from `.env` (copy `.env.example`); Appendix D lists them all.

With the development key pair in `.env` (`LICENCE_PUBLIC_KEY` and `LICENCE_SIGNING_KEY`), the stack is a **console**: the superuser can sign in and issue licence files (**Platform → Download licence file**). Copy a file into `LICENCE_DIR` and its organisation can sign in; the server verifies it at every sign-in. Only the demo's own licence is written there for you. A customer's server has no signing key, only the licence files copied into its `LICENCE_DIR`.

## 0.4 Accounts and sign-in codes

| Account | How it exists | Second step |
|---|---|---|
| **Platform superuser** | `superuser@rummaan53.com` (fixed in code) with `OWNER_PASSWORD`, created at boot **on the console only** | A code emailed to that address. |
| **Organisation owner** | The first sign-up with the address named in the organisation's licence file becomes its Administrator | A code emailed to the owner, always. |
| **Anyone else** | **Staff → Add** by an administrator, who shares the sign-in with them | A code emailed to their sign-in address, or texted when they are set to text and the organisation's SMS gateway is on. |
| **Demo logins** (`SEED_DEMO=true`, console only) | `DEMO_EMAIL` / `DEMO_PASSWORD` (owner), plus `demo-engineer@…` and `demo-viewer@…` | The texted-code step, with the code (`DEMO_SMS_CODE`) shown on screen instead of sent. At most `DEMO_SEATS` walkthroughs at once. |
| **A hand-made demo tenant** (`DEMO_ORGANISATION_CODE`, console only) | Everyone in the organisation that licence code names — on the live console, Northwind (`NWTEL`) | The same on-screen code, for the owner too; nothing is emailed. |

To act as a person in a manual test, read their code from the **code outbox**: when `OTP_OUTBOX_FILE` is set, every code is appended to that file as one JSON line (`challenge_id`, `channel`, `to`, `code`, `at`) instead of being emailed or texted. Without it, codes go to real mailboxes through the Gmail accounts in `SMTP_FROM_n` / `SMTP_PASSWORD_n`.

> [!IMPORTANT]
> 🔒 **INVARIANT** — production refuses `OTP_OUTBOX_FILE`, because a file of live sign-in codes is a second factor held by everyone who can read the file. It also refuses to boot with no SMTP account, since email is everyone's fallback, and refuses a `LICENCE_PUBLIC_KEY` that would replace the vendor's key. `packetpulsego/pkg/common/config/Config.go` returns an error at boot; `REL-004` checks the live host.

## 0.5 The automated gate

`./packetpulsetest.sh` runs every suite; `./packetpulsetest.sh unit client` is the fast pair. Manual testing **adds to** this gate — it does not replace it. A release needs both (Part V).

| Suite | What it proves | Needs |
|---|---|---|
| `tenancy` | One organisation can never read or change another's data | server |
| `assignment` | Licence files: owner sign-up, sign-in only under a genuine licence, people counted, switched off and deleted | server |
| `acl` | Every role gets exactly its capabilities, both ways | server |
| `audit` | The registry covers every mutation; the chain verifies; sign-ins (and failed ones) are recorded | server |
| `translation` | The catalogue is complete and served | server |
| `contract` | The Results API and CSV shapes, the export's destination guard | server |
| `grading` | A device's UDP test is graded and its breach counted; an HTTPS round-trip never is; automatic monitoring answers nothing | server |
| `unit` | Every Go package with `-race -shuffle`, coverage floors, govulncheck, the string generator | database |
| `client` | Every Flutter screen, `flutter analyze`, the client coverage floor | — |
| `backup` | The backup, drill and restore scripts, for real | database |
| `docs` | This guide is current and every number, path, route and capability in it matches source | — |
| `load` *(opt-in)* | 20 people, a 200-site sweep, API and CSV pulls against p95 budgets | server |

Integration suites read `PACKETPULSE_TEST_URL` (default `http://localhost:8080`). `PACKETPULSE_TEST_REQUIRE_SUPERUSER=1` turns a missing superuser into a failure rather than a skip — a skipped suite looks exactly like a passing one.

## 0.6 Deploying

`scripts/deploy/PacketPulseDeploy.sh --host mshop.rummaan53.com` builds from the working tree, runs `unit client backup` and the integration suites against `PACKETPULSE_TEST_URL`, cross-compiles for linux/arm64, builds the web app with base href `/app/`, backs up the live database, then switches the release and restarts. `--api-only`, `--web-only` and `--site-only` ship one part; `--skip-guards` and `--skip-backup` are for a logged emergency only. Part V §5.4 is what to check afterwards.

---
# Part L — Learn PacketPulse

The ideas every other part assumes, in plain words. Read this once, whatever your role.

## L.1 Organisations, people and the licence

An **organisation** is one customer — a telecom operator such as Northwind Telecom. Everything it owns (sites, diagnostics, keys, settings, its trail) is invisible to every other organisation. 🔒 **INVARIANT** — every tenant query carries the caller's `organisation_id`; a request with none is refused, never answered for everyone.

Each organisation runs on a **licence file** signed by PacketPulse and copied onto its server. The server checks it at **every sign-in**: without a genuine licence in force, nobody in that organisation signs in. The licence names the owner's address and how many **people** it covers. Everyone on the books counts, switched on or switched off, because a person who leaves is switched off rather than erased. Someone who never ran a test or signed in can be **deleted**, and then no longer counts. Signing up does not create an organisation: the first person to sign up with the owner's address, proved by an emailed code, becomes its Administrator and adds everyone else.

People hold a **role** (Administrator, NOC Engineer, Viewer, or one of the organisation's own) made of **capabilities** such as `diagnostic_run` or `staff_manage` (Appendix B). A person can be given an individual *allow* or *deny* on top. Changing someone's authority signs them out at once.

## L.2 Sites, tickets and sweeps

A **site** is an address worth testing: a customer router, a point of presence, a resolver. It may be an IP address or a hostname (resolved at test time, so a name that stops resolving is itself a fault).

A **diagnostic** tests sites and files the result against a **TT number** — a trouble ticket in the operator's own system — and a **Customer ID**. A ticket can be tested many times; the attempts together show how the fault was worked. The engine sends ICMP echoes (or TCP connects where ICMP is unavailable), in parallel.

## L.3 What the numbers mean

| Figure | Meaning | Good |
|---|---|---|
| **Loss** | Packets sent minus received, as a percentage. A browser's UDP test counts real packets and says where each lost one went: upstream, downstream or too late to play. Only an HTTPS test from a device counts *query* loss, because HTTPS resends what the network drops. | 0% |
| **Round trip** | Average, fastest and slowest response time, in ms. | Under the site's target |
| **Jitter** | RFC 3550 interarrival jitter — how much *consecutive* round trips differ. Voice cares about this more than raw latency. | Low single digits of ms |
| **MOS** | Mean Opinion Score (ITU-T G.107 E-model): how a phone call would sound over this path, 1–4.5. Computed, stored, graded against a MOS target and exported in the CSV, but shown on no screen or report since 14 October 2026 (§4.1). | Above 4.0; below 3.6 is noticed on a call |
| **SLA** | **OK**, **Degraded** (within 80% of a threshold) or **Breached**, against the site's targets — each figure shown against its own target, with one line saying why. **Not graded** when the result was not held to targets (§5.1). | OK |

## L.4 IPv4, IPv6 and "not counted"

A site whose name resolves to both an IPv4 and an IPv6 address is measured over **both**. The IPv6 result is shown, kept and exported, but marked **not counted**: the site's totals and breaches are carried by IPv4, because the customer's agreement is for the service and IPv4 still carries it. A site that is IPv6 *only* counts over IPv6 — otherwise it could never fail. When the server itself has no IPv6 route, an IPv6 address is reported as *could not be tested from here* rather than as an outage.

## L.5 Two vantage points

PacketPulse measures from two places. Its server sweeps the organisation's endpoints from PacketPulse's data centre when the API asks it to. **Run diagnostic** measures from wherever the person is, over their own connection, to the endpoint nearest them — the right tool for "is it slow for me?". The two will not match, and both are correct: they measure different layers from different places. A device test is **attached** to a ticket, under the endpoint it measured, labelled as measured on a device.

## L.6 Two-step sign-in

After the password, everyone enters a one-time **6-digit code**. It is **emailed** to their sign-in address, or **texted** to their mobile when an administrator chose text for them and the organisation's own SMS gateway is on. An owner's code is always emailed, so the account that runs the organisation never depends on a gateway. A code is good for ten minutes and five tries; ten wrong codes in all lock the account for fifteen minutes, and nobody is sent more than ten codes an hour. A wrong *password* never locks anything, so a stranger cannot lock someone out by typing their address. A lost phone is handled by an administrator, who signs that device out and switches the person to email.

## L.7 Location at sign-in

For people set to **Record location**, the device's position is recorded when they sign in and out — with the browser's or phone's permission, and a refusal recorded as one. A field engineer's appear on **Check-ins**; an administrator's beside their entries in **Activity**. An organisation can make sharing a position a condition of signing in. The address is looked up afterwards (Google, or OpenStreetMap when no key is set), so it may appear a moment later; the position is the record.

## L.8 The activity trail

Every successful change — who, what, from where, when — is written to a **hash chain**: each entry's hash covers the one before, so altering or removing any entry breaks every entry after it, and **Verify** names the first that does not follow. Administrators' sign-ins and sign-outs are in it, and so are **failed** sign-ins to administrators' accounts (at most twenty an hour per account). A position is shown beside an entry but never sealed into its hash: it is personal data that may have to be erased, and a chain entry never can be.

## L.9 Results leaving PacketPulse

| Way out | Who uses it | What |
|---|---|---|
| **Export PDF** | A person, from a ticket | The evidence document for the ticket or the customer |
| **Export CSV** | A person, from a ticket | The same results as a spreadsheet, one row per site and family |
| **Results API** | A machine with an API key | One ticket by TT number (JSON), or every result of up to 31 days as CSV |
| **Result export** | PacketPulse, on a schedule | A CSV file per period, delivered to the organisation's own SFTP, FTPS or FTP server |

All CSV comes from one writer with one header — the contract an IT system's importer is written against (`packetpulsetest/contracts/result_export_columns.json`).

---
# Part II — Feature Chapters

One chapter per area of the product, in the order a new customer meets them. Every chapter carries all four lenses — **🌟 Commercial**, **📖 User Guide**, **🧪 Testing Playbook** and **⚙️ Developer** — and `./packetpulsetest.sh docs` fails if one is missing.

Run every case on a disposable stack (Part 0 §0.3) with **two organisations of your own**: `ACME` (yours) and `RIVAL` (someone else's). Unless a case says otherwise, "an administrator" is ACME's owner and "an engineer" holds ACME's built-in *NOC Engineer* role.

| Group | Chapters | Test IDs |
|---|---|---|
| 1 · Access | Sign-in and the second step · Sign-in security | `AUTH-*`, `SMS-*`, `LOC-*` |
| 2 · People | Staff · Roles and permissions · Sessions and licence · Check-ins · Activity | `STF-*`, `ACL-*`, `WHO-*`, `CHK-*`, `AUD-*` |
| 3 · Diagnostics | Sites · Diagnostics and results · IPv6 · CSV · History and dashboard · Network reports | `SITE-*`, `DIAG-*`, `V6-*`, `CSV-*`, `HIST-*`, `RPT-*` |
| 4 · The customer's side | Run diagnostic, from this device | `DEV-*` |
| 5 · Service targets and grading | Targets through the API; how every result, a device's UDP test included, is graded | `MON-*` |
| 6 · Integrations | Results API keys · Result export · Directory | `API-*`, `EXP-*`, `LDAP-*` |
| 7 · Platform and settings | Platform console · Settings · Public site | `PLAT-*`, `SET-*`, `WEB-*` |

---
## Group 1 — Access

Signing in, the second step, and the organisation's own rules for both.

---

### 1.1 🔑 Sign-in and the second step

**Screen:** the sign-in page at `/app/` · **Routes:** `POST /user/signin`, `/user/signin/verify`, `/user/signin/resend`, `/user/signout`

- 🌟 **Commercial Presentation & Sales Pitch**: Two-step sign-in for *everyone*, not a premium add-on. A stolen password alone opens nothing: after it comes a one-time code, emailed to the person's sign-in address or — for field staff the administrator chooses — texted through **the operator's own Twilio account** (their sender, their DLT registration, their bill). Owners always get theirs by email, so the account that matters most never depends on a gateway someone else configures. There is no app to install and nothing to recover: a lost phone is one action for an administrator.
- 📖 **User Guide & Operational Flow**:
  - **Every sign-in:** email and password → *We emailed a 6-digit code to a•••@acme.example* (or *We texted a 6-digit code to •••• 1234* when your administrator set you up for text) → enter it → you are in. A code is good for 10 minutes; *Resend* after 30 seconds.
  - **The owner's first sign-in:** *First time here? Set up the owner's account* — only the address named in this server's licence can do this (§7.1). The emailed code finishes it, and the owner arrives as the Administrator.
  - **Lost phone:** your administrator signs that device out under **Who is signed in** (§2.3) and changes your number or switches you to email under **Staff → How they sign in**.
  - **Location:** if you are asked to share your location, the browser or phone asks your permission. See §1.2 for when sharing is required.
  - **When a session ends:** a session lasts twelve hours (`JWT_TTL`). When it runs out, or an administrator signs the device out, the app returns to sign-in and says *Your session has ended. Sign in again to continue.*; a dialog left open closes with it. Nothing typed in an unsaved form survives.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `AUTH-001` | New engineer | The first sign-in is the same as every other | Administrator adds the engineer (**Staff → Add**). Engineer signs in with email and password | The code step names their inbox, partly hidden; the emailed code opens the app. 🛑 **Must NOT** reach the app on the password alone, or show the whole address in the hint |
  | `AUTH-002` | Engineer | Normal sign-in | Sign out; sign in; enter the emailed code | Signed in. **Who is signed in** shows this device |
  | `AUTH-003` | Engineer | A wrong code says how many tries are left | Enter `000000` | Refused (`invalid_code`) with attempts left; after **5** wrong codes on one sign-in it is spent (`challenge_expired`, *Start again*). 🛑 **Must NOT** accept a 6th guess on the same sign-in |
  | `AUTH-004` | Engineer | A code belongs to its own sign-in | Sign in with a code; sign out; start another sign-in and enter the first code | Refused; the code emailed for this sign-in works. 🛑 **Must NOT** accept a code from an earlier sign-in |
  | `AUTH-005` | Tester | Codes cannot be pumped | *Resend* repeatedly; then sign in over and over within the hour | *Resend* unavailable for 30 s and at most 3 sends per sign-in; the 11th code to one person in an hour is refused (`rate_limited`). 🛑 **Must NOT** send email or texts without limit |
  | `AUTH-006` | Tester | Ten wrong codes lock the account, a wrong password never does | Give 10 wrong codes across fresh sign-ins; then the right password and right code. Separately, 15 wrong **passwords** on another account | First account: *Locked — try again later* (`locked_out`, 15 min) even with the right code. Second account: still signs in with the right password. 🛑 **Must NOT** lock an account by wrong passwords — a stranger could lock anyone out |
  | `AUTH-007` | Tester | Sign-in does not reveal which addresses exist | Sign in with an unknown email; then a known email with a wrong password | The same message and a similar response time, and no code sent. 🛑 **Must NOT** say "no such user" or answer the unknown address noticeably faster |
  | `AUTH-008` | Owner | An owner's code is always emailed | Set the owner to *Text message* in **Staff → How they sign in** with a gateway on; sign in as the owner | Code emailed, not texted. 🛑 **Must NOT** text an owner's code |
  | `AUTH-009` | Engineer set to text | A texted code, with resend limits | Gateway on (§1.2), engineer has a mobile number and *Text message*. Sign in; wait; *Resend* | Code arrives; the hint shows the last four digits; *Resend* unavailable for 30 s; at most 3 sends per sign-in and 10 codes an hour (`rate_limited`) |
  | `AUTH-010` | Engineer set to text | No gateway means email, not a lock-out | Switch the gateway **off**; sign in as the engineer | The code is emailed instead. 🛑 **Must NOT** refuse the sign-in for want of a gateway |
  | `AUTH-011` | Engineer set to text | A failing gateway is said, not hidden | Gateway on with a wrong auth token; sign in as the engineer | 503 *Your sign-in code could not be sent just now. Try again.* 🛑 **Must NOT** claim a code was sent |
  | `AUTH-012` | Administrator | A lost phone | **Who is signed in → Sign this device out** on the engineer's phone; **Staff → How they sign in → Email** | The phone's session ends at once; the engineer's next sign-in emails the code. Activity records who signed the device out |
  | `AUTH-013` | Tester | The superuser exists only on the console | On a customer's server (no `LICENCE_SIGNING_KEY`), sign in as `superuser@rummaan53.com` — or as any account flagged superuser in its database | Refused (*This account has been deactivated*). On the console the code is emailed to `superuser@rummaan53.com`. 🛑 **Must NOT** admit a superuser on a customer's server |
  | `AUTH-014` | Engineer | An ended session returns to sign-in once | Revoke the engineer's session from another device, then use the app | One clean return to the sign-in page, saying *Your session has ended. Sign in again to continue.*, and one `POST /user/signout` in the browser's network panel. 🛑 **Must NOT** loop between sign-in and an error, or send sign-outs over and over |
  | `AUTH-017` | Engineer | A session that runs out lands on sign-in by itself | On a test server with `JWT_TTL=2m`, sign in, open **DNS sites → Edit** on an endpoint, and wait three minutes without touching anything | The dialog closes and sign-in shows *Your session has ended…* at the two-minute mark; no `POST /user/signout` is sent. 🛑 **Must NOT** stay on the screen until the next click, or leave the dialog over the sign-in page |
  | `AUTH-018` | Engineer | A device whose clock runs fast still gets the server's answer | Set the laptop clock 13 hours ahead with the app open and signed in; reload any screen | The request carries the token (network panel: `Authorization`), the server answers, and the app returns to sign-in only when the server refuses it. 🛑 **Must NOT** send a request with no token and then show *Sign in to continue* on the screen instead of the sign-in page |
  | `AUTH-015` | Tester | A forged client address is ignored | `curl -H 'True-Client-IP: 6.6.6.6' -H 'X-Forwarded-For: 6.6.6.6'` a sign-in from a machine that is not a trusted proxy | Sessions and Activity show the real peer address. 🛑 **Must NOT** record `6.6.6.6` |
  | `AUTH-016` | Tester | A password is checked in characters, and one too long to keep is a field error | Set a password of nine Devanagari letters and a digit; then one of 80 characters (`a1` forty times), at sign-up, **Staff → Add** and **Change password** | The first is refused as *at least 10 characters*; the second as *at most 72 characters, fewer if it has accented or non-Latin letters*, on the password field. 🛑 **Must NOT** answer either with *Something went wrong*, or keep the long one by cutting it |
  | `DEMO-001` | Visitor | The public demo signs in through the real two-step flow | On the console with `SEED_DEMO` on, sign in as `demo-engineer@<domain>` with the shared demo password | The real SMS step appears with the code **shown on screen** and filled in; one tap opens the app. 🛑 **Must NOT** text or email anything |
  | `DEMO-002` | Tester | The demo code is not a master key | Use the demo code (`DEMO_SMS_CODE`, default `111111`) as the second-step code for a **non-demo** account | Refused as a wrong code. 🛑 **Must NOT** admit any account outside the demo organisation |
  | `DEMO-003` | Tester | The demo is capped | Open `DEMO_SEATS` (default and minimum 5) concurrent demo sessions, then sign in once more | The extra sign-in is refused with *the demo is busy* (`demo_busy`). 🛑 **Must NOT** let shared demo accounts open unlimited sessions |
  | `DEMO-005` | Tester | A walkthrough left idle gives up its place | Fill the demo as in `DEMO-003`; close every tab without signing out; wait 30 minutes; sign in again | Admitted: only sessions used in the last 30 minutes count. Reopening a closed tab within its 12 hours still works. 🛑 **Must NOT** keep visitors out because earlier ones walked away |
  | `DEMO-004` | Tester | The shared demo cannot be hijacked or turned on the network | As a demo account try **Change password**; then run a diagnostic against a private/internal address added to the demo inventory | Password change refused (`demo_read_only`); the private target is refused by the probe policy. 🛑 **Must NOT** let a visitor change a shared credential or probe `10.x`/internal space |
  | `DEMO-006` | Visitor, then Engineer | A walkthrough is given the example ticket | Sign in to the demo (`DEMO-001`, or any account in the `DEMO_ORGANISATION_CODE` organisation); open **Run diagnostic**; change the TT number to `TT-DEMO-2`; **Start test**; **Attach to a ticket**. Then sign in as an engineer of a non-demo organisation and open **Run diagnostic** | The demo's **Customer ID** reads `CUST-88412` and its **TT number** `TT-2026-00731` as the page opens; the run is filed under `CUST-88412` and `TT-DEMO-2`. The other organisation's fields are empty. 🛑 **Must NOT** fill in a ticket for any organisation but the demo, or put back the example over what the visitor typed |
- ⚙️ **Developer Guide & Release Confidence**:
  - Flow: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go` — the password step returns a challenge (`email` or `sms`), never a session; `newSession` runs only after verify, and checks the licence again (§7.1). Challenge ids are 32 random bytes, stored as SHA-256; codes are stored as keyed hashes bound to their challenge.
  - 🔒 Attempts are counted **before** the code is checked, in one conditional `UPDATE … RETURNING`, so parallel guesses cannot slip under the limit.
  - Limits: `maxAttempts = 5`, `lockAfter = 10`, `lockFor = 15m`, `resendAfter = 30s`, `maxSendsPerChallenge = 3`, `maxCodesPerHour = 10` by either channel; a separate rate limiter for the second step (`SECOND_STEP_ATTEMPTS_PER_MINUTE`).
  - Email: `packetpulsego/pkg/common/mailer/Mailer.go` — STARTTLS with a verified certificate (credentials are never sent otherwise), header injection refused, the load spread over up to four sending accounts (`SMTP_FROM_0…3`, `SMTP_PASSWORD_0…3`). A development or test server writes codes to `OTP_OUTBOX_FILE` instead (`packetpulsego/pkg/common/mailer/Outbox.go`); production refuses it.
  - Client IPs: `packetpulsego/pkg/common/apiratelimit/ApiRateLimit.go` walks `X-Forwarded-For` from the right past `TRUSTED_PROXIES` only, and returns canonical addresses (`::ffff:a.b.c.d` → `a.b.c.d`).
  - An ended session (since 2026-10-17): `PacketPulseApiClient` names the token a 401 refused, and the auth bloc's `PacketPulseAuthSessionEnded` acts only while that token is the session in use - a late refusal of a session already replaced ends nothing. `readToken` in `packetpulseflutter/lib/common/services/PacketPulseSessionStore.dart` sends a token past its time rather than dropping it, so the server decides. `packetpulseflutter/lib/main.dart` ends a session when its `expires_on` comes and closes every route above sign-in when the session changes.
  - Coverage: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps_test.go`, `packetpulsego/pkg/common/mailer/Mailer_test.go`, the `assignment` and `audit` suites (every integration test finishes its sign-in with the code from the outbox); for an ended session, `packetpulseflutter/test/app_test.dart`, `packetpulseflutter/test/plumbing_test.dart` and `packetpulseflutter/test/sign_in_test.dart`.

---

### 1.2 🛡️ Sign-in security — the SMS gateway and the location rule

**Screen:** Configure → **Sign-in security** · **Routes:** `GET/PUT /sms/gateway`, `POST /sms/gateway/test`, `GET/PUT /organisation/settings` · **Capability:** `staff_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: The operator brings its own SMS provider account, so codes go out under its registered sender at its negotiated rates — and in India, under its own DLT template. One switch makes **sharing a location a condition of signing in**, for operators who must prove where field staff were. Owners are always let in, so a browser that will not share a position can never lock the organisation out.
- 📖 **User Guide & Operational Flow**:
  - **SMS gateway:** Twilio account SID (`AC…`), auth token (stored encrypted, never shown again — leave empty to keep it), a sender number or a messaging service SID, and the message with `%s` where the code goes. **Send a test to my mobile** texts *your own* number from your staff record.
  - **Location at sign-in:** *Require location to sign in* refuses a sign-in without a position for people whose location is recorded. Who is recorded is set per person on **Staff → Record location at sign-in and sign-out**.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SMS-001` | Administrator | The form names each mistake | Save with SID `AB123`, sender `98765`, message without `%s`, and a 301-character message | Each field shows its own error (`validation_failed`). 🛑 **Must NOT** save any part of it |
  | `SMS-002` | Administrator | The token is kept, never shown | Save a full gateway; reload; save again with the token empty | *A token is stored. Leave this empty to keep it.* hint; the second save keeps it. `GET /sms/gateway` carries `has_auth_token: true` and no token. 🛑 **Must NOT** return the token or clear it on an empty save |
  | `SMS-003` | Administrator | Switching on needs everything a send needs | Turn the gateway on with no sender and no token | Refused against those fields. Saved switched **off**, the half-filled form is kept |
  | `SMS-004` | Administrator | The test goes only to my own mobile | Clear your own mobile number on the Staff screen; *Send a test to my mobile*; then add it and send again | First: *Add your own mobile number on the Staff screen*. Second: sent, naming your number. 🛑 **Must NOT** accept a number from the request — the endpoint takes none |
  | `SMS-005` | Administrator | A refused account is named | Save a wrong token, switch on, send a test | *The SMS provider refused these credentials* against the token. 🛑 **Must NOT** show the provider's raw reply (it can carry the SID) |
  | `LOC-001` | Administrator | The location rule saves alone | Turn *Require location to sign in* on | Saved; the device-test target (§4.1) is unchanged. 🛑 **Must NOT** reset any other organisation setting |
  | `LOC-002` | Engineer | Required means refused without a position | Rule on; engineer with *Record location* on; deny the browser's location prompt at sign-in | Refused (`location_required`, action *share location*) with guidance to allow it. Allowing it signs in. The owner, denying it, still signs in |
  | `LOC-003` | Engineer | Not required means recorded as refused | Rule off; deny the prompt | Signed in; **Check-ins** shows *Location refused* for that sign-in |
  | `LOC-004` | Engineer | A garbled position is not trusted | Send `POST /user/signin/verify` with `latitude: 123` | Refused as an invalid location (422) at sign-in; at sign-out recorded as *unavailable* rather than refusing the sign-out |
- ⚙️ **Developer Guide & Release Confidence**:
  - Gateway: `packetpulsego/pkg/smsmicroservice/smsservice/SmsService.go` (token sealed with the server's secret box, keyed from `JWT_SECRET`); provider: `packetpulsego/pkg/common/smsprovider/` (Twilio, fixed host).
  - Settings: `packetpulsego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`; `PUT /organisation/settings` requires **both** `require_checkin_location` and `device_test_target` — a whole-row save.
  - Policy: superusers and owners are always asked for a position and never required to give one; no staff row means not asked; otherwise the person's `capture_location` and the organisation's rule decide. An unreadable policy fails closed.
  - Coverage: `packetpulsego/pkg/smsmicroservice/**`, `packetpulsego/pkg/usermicroservice/userservice/UserCheckin_test.go`, `packetpulseflutter/test/sign_in_security_screen_test.dart`.

---
## Group 2 — People

Who is in the organisation, what each may do, who is signed in, and the record of what they did.

---

### 2.1 👥 Staff

**Screen:** Administer → **Staff** · **Routes:** `GET /staff/list`, `PUT/DELETE /staff/{staffId}`, `PUT /staff/{staffId}/secondfactor`, `POST /user/add` · **Capability:** `staff_manage` (adding people: `user_manage`)

- 🌟 **Commercial Presentation & Sales Pitch**: One screen to add a colleague, give them a role, decide how they sign in and whether their position is recorded — and to stop them the moment they leave. Switching someone off signs them out at once, not when a token happens to expire, which is the property an auditor asks about first; every test they ran stays on the record. Someone added by mistake can be deleted, which frees their place on the licence.
- 📖 **User Guide & Operational Flow**: **Add** a person with email, a starting password and a role, and give them the password yourself: staff do not sign themselves up. Edit to change their role, staff code (unique in the organisation), department and designation (labels only — authority comes from the role). **How they sign in** sets where their code goes — email, the default, or text with their mobile number — and **Record location at sign-in and sign-out**. **Disable** switches off someone who has left: they are signed out everywhere and cannot sign in, their tests stay, and they still count on the licence. **Delete** is only for someone added by mistake who has never signed in or run a test, and frees their place.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `STF-001` | Administrator | Add a colleague | **Add** with role *NOC Engineer* | They appear with the role; their first sign-in emails a code (`AUTH-001`); **People on the licence** (§2.3) rises by one |
  | `STF-002` | Administrator | Staff codes are unique inside the organisation | Give two people the same staff code; then give RIVAL's person the same code | First: refused (`conflict`). RIVAL: allowed — uniqueness is per organisation |
  | `STF-003` | Administrator | Disabling signs out at once | Engineer signed in on another browser; **Disable** them | The engineer's next action returns them to sign-in; signing in is refused before any code is sent; they still count on the licence. 🛑 **Must NOT** leave the session working until it expires |
  | `STF-004` | Administrator | How they sign in, saved whole | Set *Text message* with mobile `+919876543210` and *Record location* off; save. Then edit only the department and save | Second save keeps the method, number and location setting. 🛑 **Must NOT** reset what the department edit did not show |
  | `STF-005` | Administrator | A bad mobile number is refused | Set *Text message* with `98765 43210` | Refused: international form needed (E.164) |
  | `STF-006` | Administrator | Changing a role signs the person out | Engineer signed in; change their role to *Viewer* | Their next action returns them to sign-in; signed in again, they see only Viewer's screens |
  | `STF-007` | Viewer | Without `staff_manage` there is no Staff screen | Sign in as a Viewer; call `GET /staff/list` | No **Staff** in the rail; the API answers 403 |
  | `STF-008` | Administrator | Only someone with no records can be deleted | Add a person and **Delete** them. Add another, let them sign in once, and **Delete** them | First: gone, and **People on the licence** falls by one. Second: refused (409) — *switch them off instead*. 🛑 **Must NOT** delete a person who has signed in or run a test |
  | `STF-009` | Administrator | A full licence refuses the next person, and says why | Licence for 3 people; 3 on it, one of them disabled; **Add** a fourth | Refused (`seat_limit_reached`): *Your licence covers no more users. Delete someone added by mistake, or ask for a licence for more users.* 🛑 **Must NOT** add the fourth, or count only those switched on |
  | `STF-010` | Administrator | Someone deleted by mistake can be added back | Delete a person who never signed in; **Add** the same address again | They are back with the new password and role, and counted again. 🛑 **Must NOT** refuse the address as taken |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/staffmicroservice/staffapp/StaffRouteHandler.go`; the second-step settings are a **separate** route (`PUT /staff/{staffId}/secondfactor`) because the whole-row staff update would otherwise wipe them; it requires `second_factor`, `phone_number` and `capture_location`.
  - Delete is a soft delete (`deleted_on`), refused while the person has records — any diagnostic submitted, any session ever opened — so no test or trail entry points at a removed person; the account is switched off in the same statement. Adding the address again revives the row (`StaffReviveWithCredential`).
  - 🔒 Every write is scoped by the caller's organisation; editing RIVAL's `staffId` answers 404.
  - Coverage: `packetpulsego/pkg/staffmicroservice/**`, `packetpulseflutter/test/staff_screen_test.dart`, the `acl`, `tenancy` and `assignment` suites.

---

### 2.2 🧩 Roles and permissions

**Screen:** Administer → **Permission matrix** · **Routes:** `GET/POST /staff/role/*`, `PUT /staff/{staffId}/access`, `GET /user/capability/list` · **Capability:** `acl_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Sixteen plain-language capabilities, three built-in roles (Administrator 16, NOC Engineer 6, Viewer 5) and as many of the operator's own as it likes — plus a per-person *allow* or *deny* for the exception that does not deserve a role. Changes take effect on the next request, not the next sign-in.
- 📖 **User Guide & Operational Flow**: The matrix lists roles across and capabilities down. **Add a role**, tick what it may do, save — the whole set is saved, so nothing is left to an invisible default. Built-in roles are read-only: an engineer reads the tests they ran (`diagnostic_view_own`), while Administrator and Viewer read everyone's (`diagnostic_view_all`). A role somebody holds cannot be deleted. Per person: *Inherit*, *Allow* or *Deny* each capability.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `ACL-001` | Administrator | Built-in roles cannot be changed | Try to edit *Viewer* | Not editable; `PUT /staff/role/{viewerId}` answers 403 or 409. 🛑 **Must NOT** change a role every organisation shares |
  | `ACL-002` | Administrator | A custom role grants exactly what is ticked | Add *Shift lead* with `diagnostic_run` and `report_view` only; give it to an engineer | They can run diagnostics and see the dashboard; **Staff** and **API keys** are absent. Each refused route answers 403 |
  | `ACL-003` | Administrator | A role in use cannot be deleted | Delete *Shift lead* while someone holds it | Refused, naming that people hold it |
  | `ACL-004` | Administrator | A per-person deny wins over the role | Deny `report_export` to one engineer | That engineer has no **Export PDF**/**Export CSV**; `GET /diagnostic/{id}/report.csv` answers 403. Other engineers unaffected |
  | `ACL-005` | Tester | Every role, both directions | Run `./packetpulsetest.sh acl` | Passes — each role is refused what it lacks and allowed what it holds |
  | `ACL-006` | Administrator | An engineer reads their own tests; an administrator everyone's | Two engineers each run a diagnostic; each opens **History**; then the administrator does | Each engineer sees only their own, with *These are the tests you ran*; the administrator sees both (`HIST-005`). 🛑 **Must NOT** show an engineer a colleague's test, in the list, by its id or by its ticket |
  | `ACL-007` | Administrator | A role held only by someone deleted is still in use | Give *Shift lead* to a person, delete them, then delete *Shift lead* | Refused, naming that people hold it - the deleted person's records keep their role. 🛑 **Must NOT** answer *Something went wrong* |
- ⚙️ **Developer Guide & Release Confidence**:
  - Capabilities: `packetpulsego/pkg/common/packetpulseaccess/PacketPulseAccessCategory.go` (Appendix B); guards: `packetpulseaccess.RequireCapability` on each route, or `RequireAnyCapability` where either of two will do — reading diagnostics opens to `diagnostic_view_own` or `diagnostic_view_all`, and the server decides whose rows come back (Appendix A).
  - 🔒 Authority is read fresh on every request; a role change revokes the person's sessions.
  - Coverage: `packetpulsetest/golang/aclconformance/acl_conformance_test.go` (the full matrix), `packetpulseflutter/test/permission_matrix_screen_test.dart`.

---

### 2.3 🪪 Who is signed in, and the licence

**Screen:** Administer → **Who is signed in** · **Routes:** `GET /staff/session/list`, `DELETE /staff/session/{sessionId}`, `GET /staff/session/organisation`, `DELETE /staff/session/organisation/{sessionId}`

- 🌟 **Commercial Presentation & Sales Pitch**: A licence covers a number of **people** — everyone on the books, working or switched off — not sign-ins. One engineer on a laptop and a phone is one person. An administrator sees at a glance how many places are used and who is signed in where, and signs a lost phone out in one action.
- 📖 **User Guide & Operational Flow**: Three figures: **People on the licence**, **People licensed** and **Live sessions**. When every place is taken a warning says so, before the next person is refused. Your own sessions list every device you are signed in on, with its address (IPv4 or IPv6, in full in the tooltip). Administrators also see everyone signed in, and **Sign this device out** ends one session — for a lost phone, or a machine somebody walked away from. Signing a device out frees no place: the licence counts people.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `WHO-001` | Owner | The licence counts people, not sign-ins | Sign in on a laptop and a phone | **People on the licence** 1, **Live sessions** 2. 🛑 **Must NOT** count the second device as a second person |
  | `WHO-002` | Administrator | Switched-off people still count; deleted ones do not | Licence for 2: the owner and an engineer who never signed in. **Disable** the engineer; then **Delete** them | Disabled: still 2, and adding a person is refused (`STF-009`). Deleted: 1, and adding a person succeeds |
  | `WHO-003` | Administrator | Sign a device out | **Sign this device out** on a colleague's session | Their next action on that device returns them to sign-in; Activity records who signed out whose device; **People on the licence** unchanged |
  | `WHO-004` | Engineer | An IPv6 address shows whole | Sign in over IPv6 (or seed a session with `2401:4900:1c2a:8e1f::1`) | One line, ellipsised, full address in the tooltip and selectable. 🛑 **Must NOT** wrap across lines or overflow at phone width |
  | `WHO-005` | Administrator | A licence figure that cannot be read is not a number | Give a custom role `staff_manage` but not `licence_view`; open the screen as someone holding it | **People licensed** shows `–`; the sessions still list. 🛑 **Must NOT** show 0 or "unlimited" |
- ⚙️ **Developer Guide & Release Confidence**:
  - The count is `StaffCountUsers` — staff rows not deleted, switched on or off (`packetpulsego/pkg/staffmicroservice/staffdomain/repository/StaffRepositoryPostgres.go`). It is checked at every sign-in against the installed licence file (`enforceUserLimit`, the owner excepted, §7.1). It is also checked when a person is added or revived, by `ensureRoomForAnotherUser` in the same file.
  - 🔒 That check runs inside the transaction that writes the person, after locking the organisation's row `FOR NO KEY UPDATE`. The lock and the count are separate statements, because a count in the statement that waited for the lock reads from before the other addition committed. Before October 2026 the count ran before the write: twenty additions racing for one place created up to ten people, and a licence for twenty ended with twenty-two.
  - Coverage: `packetpulsetest/golang/tenancyassignment/`, `packetpulsetest/golang/tenancyisolation/`, `packetpulsego/pkg/usermicroservice/userservice/UserLicenceFile_test.go`, `packetpulseflutter/test/session_screen_test.dart`.

---

### 2.4 📍 Check-ins

**Screen:** Administer → **Check-ins** · **Route:** `GET /staff/checkin/list` · **Capability:** `staff_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Where every field engineer was when they signed in and out — position, accuracy, street address and a map link — on a list of its own, separate from the administrators' audit trail. Proof of attendance without a separate app.
- 📖 **User Guide & Operational Flow**: One row per session of a person whose location is recorded and who is **not** an administrator: who, signed in (time, place, *Open in Maps*), signed out (time, place, or *still signed in* / *expired* / *ended*), address and device. Filter by date range; the default is the last seven days.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CHK-001` | Engineer | A sign-in records where | *Record location* on; sign in and allow location | **Check-ins** shows the time, the position ± accuracy and — a moment later — the address; *Open in Maps* opens the spot |
  | `CHK-002` | Engineer | A sign-out records where it ended | Sign out, allowing location | The same row gains the sign-out time and place |
  | `CHK-003` | Engineer | A session left to expire says so | Sign in; let the session expire (or shorten `JWT_TTL`) | *Expired*, not a sign-out time |
  | `CHK-004` | Engineer | A refusal is recorded as one | Deny the location prompt (rule off) | *Location refused* — no position, no map link |
  | `CHK-005` | Administrator | Administrators are not on Check-ins | The owner signs in with location | No Check-ins row; the sign-in is in **Activity** (`AUD-001`) |
  | `CHK-006` | RIVAL administrator | Another organisation's check-ins are never shown | RIVAL opens Check-ins | Only RIVAL's people. 🛑 **Must NOT** show ACME's |
- ⚙️ **Developer Guide & Release Confidence**:
  - `session_checkin`, one row per session, keyed by `staff_session.session_id` (`packetpulsego/pkg/common/dbclient/migrations/0020_2026_10_03_session_checkin.sql`); coordinates only with status `captured`, checked by `CHECK`s.
  - Addresses: a background worker (`packetpulsego/pkg/common/geocode/Geocode.go`) — Google with `GOOGLE_MAPS_API_KEY`, else OpenStreetMap Nominatim at one request a second. `REVERSE_GEOCODING=off` disables it.
  - Coverage: `packetpulsego/pkg/staffmicroservice/staffservice/`, `packetpulseflutter/test/checkin_screen_test.dart`, `packetpulsetest/golang/tenancyisolation/`.

---

### 2.5 🧾 Activity

**Screen:** Administer → **Activity** · **Routes:** `GET /auditlog/list`, `GET /auditlog/verify` · **Capability:** `auditlog_view`

- 🌟 **Commercial Presentation & Sales Pitch**: A tamper-evident record of every change and every administrator sign-in — device, browser, address, second step and where they were — and of every **refused** attempt on an administrator's account, with why. Hash-chained: altering or deleting one entry breaks every entry after it, and *Verify chain* names the first. Positions sit beside entries, never inside their hash, so the record stays verifiable *and* erasable under a retention policy.
- 📖 **User Guide & Operational Flow**: Newest first. Select a row for its details: role, second step, device, **browser** ("Chrome 128 on macOS", with the raw string beneath), place with *Open in Maps*, session id and entry hash. **Sign-ins and sign-outs only** narrows the list. A *Sign-in failed* row shows **Why**: wrong password, wrong code, or locked after too many wrong codes.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `AUD-001` | Administrator | An administrator's sign-in is recorded in full | Sign in as the owner from Chrome, allowing location; open **Activity** | *Signed in* by the owner: role, second step, device, *Chrome NN on <system>*, the IP (v4 or v6), the place with a map link, the session id |
  | `AUD-002` | Administrator | And their sign-out | Sign out with location; sign back in | A *Signed out* row with its own place |
  | `AUD-003` | Tester | A wrong password on an administrator's account is recorded | From another browser, sign in as the owner with a wrong password | *Sign-in failed* — Why: *Wrong password*, from that address. 🛑 **Must NOT** carry a position in the entry |
  | `AUD-004` | Tester | A wrong code, and the lock | Owner's right password, then wrong codes until locked | *Sign-in failed* rows: *Wrong code* (naming how the code was sent) until the 10th, which says *Locked after too many wrong codes*; the next right password is recorded as locked too |
  | `AUD-005` | Tester | Not everyone's failures are activity | An engineer's wrong code; an unknown email's wrong password | Neither appears. 🛑 **Must NOT** record a field engineer's mistype, or list addresses people guessed |
  | `AUD-006` | Tester | The record cannot be flooded | 25 wrong passwords at the owner's address within an hour | At most **20** *Sign-in failed* rows for that account in the hour; the rest go to the server log |
  | `AUD-007` | Administrator | Verify the chain | *Verify chain* | *Intact*, with how many entries were checked and the head hash |
  | `AUD-008` | Tester | Tampering is caught (disposable stack only) | `UPDATE auditlog_activity SET actor_email='x' WHERE activity_id = <some id>`; *Verify chain* | *Broken* at exactly that entry, with the reason |
  | `AUD-009` | Engineer | A refused change leaves no entry | As a Viewer, try to add a site (403) | No entry. Only what happened is recorded |
  | `AUD-010` | Administrator | Exports and integrations are on the record | Download a ticket's CSV; save the result export; test its connection | Rows: *Exported* (diagnostic), *Changed* and *Tested* (result_export) |
  | `AUD-011` | RIVAL administrator | Another organisation's trail is never shown | RIVAL opens Activity | Only RIVAL's entries |
- ⚙️ **Developer Guide & Release Confidence**:
  - Registry: `packetpulsego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go` — every mutating route is there or excluded with a reason, and `TestEveryMutationRouteIsRegistered` fails a route added without either.
  - Sign-ins are recorded by UserMS itself (`recordSignIn`, `recordSignOut`, `recordFailedSignIn` in `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go`) because the routes are public; failed ones are capped by `maxFailedEntriesPerHour = 20` per account.
  - 🔒 The list joins positions from `session_checkin` by session id; the chain's `details` never hold coordinates.
  - Coverage: `packetpulsetest/golang/auditconformance/`, `packetpulsego/pkg/common/auditlog/`, `packetpulseflutter/test/audit_log_screen_test.dart`, `packetpulseflutter/test/user_agent_test.dart`.

---
## Group 3 — Diagnostics

The core of the product: what to test, testing it against a ticket, and reading what came back.

---

### 3.1 🌐 Sites

**Screen:** Configure → **DNS sites** · **Routes:** `GET /dnssite/list`, `POST /dnssite/add`, `POST /dnssite/bulkimport`, `POST /dnssite/importfile`, `PUT/DELETE /dnssite/{dnsSiteId}` · **Capabilities:** `dns_site_view`, `dns_site_manage` — the page is offered only to whoever holds both

- 🌟 **Commercial Presentation & Sales Pitch**: Every city its own destinations, as many as it needs: give each endpoint its position and a field engineer's test sweeps the city they are standing in, without their choosing anything. Name the reflector beside an endpoint and a ticket's UDP test measures the path to it. Drop the customer's spreadsheet on the page to add or place hundreds at once — one bad row never rejects the rest — and PacketPulse refuses to be turned against its own host: loopback, link-local and cloud-metadata addresses are never probed, whatever a site says.
- 📖 **User Guide & Operational Flow**: **Add site** with a name, an IP address or hostname, an optional circuit ID, its **City or region**, its **Latitude \*** and **Longitude \*** (required since 2026-10-17: a test is aimed at the endpoint nearest the device) and an SLA policy. The list shows each endpoint's position, or *No position* for one added before positions were required or by **Bulk import**; editing it asks for one. **Import file** takes a CSV file dropped on the page or chosen, or rows pasted from a spreadsheet: the first row names the columns — *Name*, *Address*, *City*, *Latitude*, *Longitude*, in any order. A row whose address is already monitored updates that endpoint with the columns given and leaves the rest; any other row adds one, and needs its *Latitude* and *Longitude*. Each row is reported, a refused one by its line. **Bulk import** still takes plain addresses by lines, commas or spaces; `Branch 12=10.0.0.1` names a site. A plain address has no position, so these show *No position* until edited or placed by a file. There is no limit on how many endpoints an organisation or a city has.
  - **Who is given the page:** the endpoints are configuration, so **DNS sites** is in the rail only for whoever manages them (`dns_site_manage`, with `dns_site_view` to read the list). An engineer or a viewer reads the endpoints under `dns_site_view` alone, which is what lets **Run diagnostic** offer them (§4.1), and is not given the page (since 2026-10-09).
  - **UDP reflector address** (optional) is where the endpoint's PacketPulse reflector answers: a host or IP, with `:port` when its signalling port is not 50000. Often not the endpoint's own address — a router cannot run the daemon, a small host beside it can. Left blank, a ticket's UDP test measures to the PacketPulse network instead (§4.2). PacketPulse's own address there does the same: this server's reflector answers through the app, never on port 50000 (`UDP-026`). PacketPulse's own standalone test reflector (`REFLECTOR_NETWORK_HOSTS`, live `140.245.28.232`) answers any endpoint that names it, of any organisation, so a development centre can test to it from wherever its nearest endpoint is (`UDP-044`). Whoever manages the endpoints gets a key icon on each row, **Reflector install key**, with the key and the install command to copy; the reflector answers only tests this organisation signs for this endpoint.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SITE-001` | Administrator | Add by address and by name | Add `Mumbai POP` = `203.0.113.10`; add `Resolver` = `one.one.one.one` | Both listed; the hostname is resolved at test time, not stored as an address |
  | `SITE-002` | Administrator | One bad line does not sink a paste | Bulk import `Pune=203.0.113.11`, `not an address`, `Pune=203.0.113.11` | First added, second invalid with a reason, third a duplicate |
  | `SITE-003` | Administrator | Labels with spaces survive | Bulk import `Branch 12=10.0.0.12` | One site named `Branch 12`. 🛑 **Must NOT** split into `Branch` and `12` |
  | `SITE-004` | Engineer | The host is never a target | Add `127.0.0.1`, `169.254.169.254` and `::1`; run a diagnostic over them | Each result is a refusal (*destination refused by probe policy*), never a measurement. 🛑 **Must NOT** send traffic to loopback or cloud metadata |
  | `SITE-005` | RIVAL administrator | The same address in two organisations | RIVAL adds `203.0.113.10` too | Allowed — the endpoint is unique per organisation, not globally |
  | `SITE-006` | Administrator | A position is required, whole and on the Earth | Add a site with both blank; then only a latitude; then a longitude of `190`; then both in range | **Latitude \*** and **Longitude \*** marked required; *Enter the latitude.* and *Enter the longitude.*, then the longitude asked for, then *Longitude is a number between -180 and 180.*, nothing sent each time; the last saves. `POST /dnssite/add` with no position answers 422 with both fields named. 🛑 **Must NOT** store a site without a position from the form, half a position, or 0, 0 for none |
  | `SITE-007` | Administrator | A dropped spreadsheet adds and places | In a browser, **Import file**; drag a CSV with `Name,Address,City,Latitude,Longitude` and four rows, one with no address, onto the page; **Import** | *2 added and 1 updated, of 4 rows*-style summary; the row with no address named by its line; the list shows each position. 🛑 **Must NOT** reject the whole file for one bad row |
  | `SITE-008` | Administrator | Positions for endpoints you already have | Edit a site's circuit and notes; import a file of `Address,Latitude,Longitude` naming it | The site gains its position; its name, city, circuit, notes and enabled state are unchanged. 🛑 **Must NOT** clear what the file had no column for |
  | `SITE-009` | Administrator | Rows pasted from Excel | Copy four rows (heading first) from Excel or Sheets; paste into **Import file**; **Import** | Read as tab-separated; each row reported. A Windows-saved CSV with `São Paulo` imports with the name intact |
  | `SITE-010` | RIVAL administrator | An import never reaches across | RIVAL imports a file naming ACME's addresses with new positions | RIVAL gains its own endpoints; ACME's are unchanged |
  | `SITE-011` | Administrator | A reflector address is kept | Edit a site: **UDP reflector address** `Reflector.Pune.Example:50001`; save; edit only its notes; save; then `PUT /dnssite/{id}` without `reflector_address`, as an app from before 2026-10-09 sends it | Stored as `reflector.pune.example:50001` and still there after both edits; *Optional* — no `*` on the label. Clearing the field and saving clears it. 🛑 **Must NOT** clear the address on an edit that did not touch it |
  | `SITE-012` | Administrator | A wrong reflector address is refused on its field | Enter `not an address`, then `203.0.113.20:70000` | Refused under the field, the rest of the edit not saved |
  | `SITE-013` | Administrator | The install key | Press the key icon on a site; **Copy** the install command; press it again | The key and the `install-reflector.sh --key …` line, copied with *Copied to the clipboard.*; the same key every time for that site. An engineer never reaches the key: no **DNS sites** in their rail (`SITE-014`), and the API refuses them (`UDP-010`) |
  | `SITE-015` | Administrator | A file adds no endpoint without its position | **Import file** with `Name,Address,City,Latitude,Longitude`: one new row with both, one new row with neither, and an existing endpoint's row with neither | The first added; the second refused by its line with *This endpoint is new, so give its latitude and longitude.*; the third *Already monitored; nothing to change.* 🛑 **Must NOT** add an endpoint with no position from a file, or refuse the whole file for one row |
  | `SITE-016` | Administrator | An endpoint without a position is placed when edited | **Bulk import** `Branch 9=203.0.113.99`; the list says *No position*; edit only its notes and **Save**; then give it a latitude and longitude and **Save** | The first save is held with *Enter the latitude.* and *Enter the longitude.*, nothing sent; the second saves and the list shows the position. 🛑 **Must NOT** save the edit of an endpoint while it still has no position |
  | `SITE-014` | Engineer, then Viewer, then Administrator | The inventory is the endpoint managers' | Sign in as an engineer; look for **DNS sites** in the rail, then open **Run diagnostic**; repeat as a Viewer; then as an administrator | No **DNS sites** for the engineer or the Viewer, while **Run diagnostic** still lists the endpoints for the engineer; the administrator has **DNS sites** under Configure. 🛑 **Must NOT** offer an engineer the endpoint inventory, or take the endpoints off their **Run diagnostic** |
  | `SITE-017` | Administrator, then Engineer | An endpoint can show its reflector only | Edit an endpoint with a **UDP reflector address**: **Reports show** *The reflector only*; save. As a NOC Engineer, run a UDP test of it; open the ticket, its PDF and **Export CSV**; then as the Administrator | The engineer sees the reflector and the IP the browser reached it at (*Reflector … at …*), and an empty address and resolved IP in the screen, the PDF and the CSV; the Administrator sees both. An edit that does not name the choice keeps it. 🛑 **Must NOT** show the endpoint's address to anyone without `dns_site_manage`, or clear the choice on an edit from an older app |
- ⚙️ **Developer Guide & Release Confidence**:
  - Policy: `packetpulsego/pkg/common/probeguard/ProbeGuardPolicy.go` — every address a name resolves to is checked, and one refused address refuses them all; IPv4-mapped IPv6 is judged as IPv4.
  - Positions: `latitude`/`longitude` on `dns_site`, both or neither, in `packetpulsego/pkg/common/dbclient/migrations/0029_2026_10_07_endpoint_location_and_report_branding.sql`. The form and a file's new rows require them since 2026-10-17 (`positionProblems` with `required` in `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteService.go`, and `importRow` in `DnsSiteImport.go`); the column stays nullable because **Bulk import** and endpoints added before then have none. The nearest city is a pure function, `ChooseNearestCity` in `packetpulsego/pkg/dnssitemicroservice/dnssitedomain/shared/DnsSite.go`: the nearest **enabled** endpoint with a position names the city (its region, ignoring case), and every enabled endpoint in that city is swept; with no city, the endpoints at the nearest one's position.
  - Import: `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteImport.go` reads the heading row, the delimiter (tab, semicolon or comma) and a spreadsheet's byte-order mark; an existing endpoint is updated with `COALESCE`, so a column the file lacks is never cleared. Up to 5000 rows a file. The drop zone and chooser are `packetpulseflutter/lib/common/services/PacketPulseFileSource.dart`, browser-only; elsewhere the paste box alone.
  - Reflector address: `dns_site.reflector_address` in `packetpulsego/pkg/common/dbclient/migrations/0031_2026_10_09_udp_split_and_endpoint_reflector.sql`, validated by `ValidateReflectorAddress` in `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteAddressValidator.go`. The edit is whole-row, so an edit that does not name the field — from an app that predates it — keeps the stored address (`DnsSiteUpdate` in `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteService.go`); imports never touch it.
  - Who is offered the page: the `dns_sites` destination in `packetpulseflutter/lib/common/presentation/PacketPulseShell.dart` asks for `dns_site_manage` and `dns_site_view`, so `DnsSiteScreen` has no read-only mode: it always offers adding, importing, editing, the key and removal. The rail only hides it; the server still refuses every change without `dns_site_manage` (Appendix A).
  - Coverage: `packetpulsego/pkg/dnssitemicroservice/**`, `packetpulsego/pkg/common/probeguard/`, `packetpulseflutter/test/dns_site_screen_test.dart`, `packetpulseflutter/test/dns_site_position_import_test.dart`, `packetpulseflutter/test/shell_test.dart`.
  - ⚠️ **TRAP** — a licence carries a **site limit** (shown on the Platform console), but adding or importing sites does not check it yet. Do not file a site count above the limit as a regression; it is a known gap.

---

### 3.2 🩺 Diagnostics and results

**Screen:** a ticket's result, from **History**; no screen starts a sweep (since 2026-10-07: **Run diagnostic** is the test from this device, §4.1) · **Routes:** `POST /diagnostic/submit`, `GET /diagnostic/{requestId}`, `GET /diagnostic/{requestId}/report.pdf` · **Capabilities:** `diagnostic_run`, `diagnostic_view_own` or `diagnostic_view_all`, `report_export` · **Licence:** required to run

- 🌟 **Commercial Presentation & Sales Pitch**: One ticket, one report. Every figure a NOC argues about — loss, latency, RFC 3550 jitter — from PacketPulse's server, and from the engineer's own device on **Run diagnostic**. A lapsed licence stops new tests but never takes away the evidence already gathered.
- 📖 **User Guide & Operational Flow**: A sweep runs on PacketPulse's server when `POST /diagnostic/submit` asks for one; nothing sweeps on a schedule since automatic monitoring was removed (§5.1). **Run diagnostic** no longer starts one (since 2026-10-07); it is the test from this device against one endpoint (§4.1). A sweep given a position tests the endpoints of the city nearest it (§3.1); with no position — location refused, no fix, or no endpoint placed yet — every enabled endpoint is swept. Only whoever manages the endpoints (`dns_site_manage`) may name endpoints instead; an engineer's list is ignored. The result says **Tested from** — the device's position as the test ran, beside the sign-in place — and **Endpoints**: the nearest city and how far away it was, every endpoint, or chosen by hand. The result shows headline cards (sites reachable, average loss, average jitter), then a row per site and family: reachable, the packet line verbatim (*Sent = 4, Received = 4, Lost = 0*), round trips, jitter and SLA grade. A result that was not graded reads *Not graded*, and a note under the table says what that means and why for each reason present: an HTTPS test from a device (HTTPS resends what the network drops), a device test filed before 14 October 2026 (before device tests were graded), or an address PacketPulse could not test. A result filed under planned work before October 2026 reads *Excluded*, with its own line. The voice-quality estimate (MOS) is not shown (§4.1). **Export PDF** and **Export CSV** sit at the top.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DIAG-001` | Tester | A sweep files against the ticket | `POST /diagnostic/submit` against every enabled site with TT `TT-MAN-001` | Status *Completed*; one row per site (and family); the packet line exactly as specified |
  | `DIAG-002` | Tester | A ticket keeps every attempt | Submit `TT-MAN-001` again | **History** for the ticket shows both attempts, newest first |
  | `DIAG-003` | Engineer | A failure is reported, not traced | Include a black-holed address (`203.0.113.99`); submit; open the result, its PDF and the **Dashboard** | The row shows the failure and its packet line. No path, hops or fault verdict anywhere, and no *Where the faults lay* card. 🛑 **Must NOT** offer a *Trace failures* switch |
  | `DIAG-004` | Engineer | The PDF is evidence | **Export PDF** | Named `packetpulse-<TT>-<time>.pdf` in the organisation's time zone; Customer ID, TT, every time in that zone and named (*Checked (IST)*, *Started … IST* for India), a UDP test's reflector and where it was reached, a Jitter column and no MOS column, average loss and jitter cards; a long IPv6 address wraps onto two lines; the foot names who ran it and where — *Triggered by Asha Rao  -  at MG Road, Pune*. 🛑 **Must NOT** truncate an address |
  | `DIAG-005` | Engineer | Loss and jitter on the ticket | Open the result; open **History** | Headline cards show average loss and average jitter; the history row shows the same figures |
  | `DIAG-006` | Engineer | A lapsed licence stops new runs only | Platform suspends the licence; submit a sweep, and attach a device run (§4.1); open an old one and export it | Both refused (`licence_suspended`, who can renew named); the old result opens and exports. 🛑 **Must NOT** hide recorded evidence |
  | `DIAG-007` | RIVAL engineer | Another organisation's ticket is not found | Open `/diagnostic/<ACME request id>` as RIVAL | 404. 🛑 **Must NOT** reveal that the ticket exists |
  | `DIAG-008` | Tester | A malformed id is refused before it is looked up | `GET /diagnostic/not-a-uuid` | 400 |
  | `DIAG-018` | Customer | A device UDP ticket reads as the test | File a 30 s UDP test that loses 16 of 1494 packets against the built-in 1% target; open it in **History**; **Export PDF** | Heading *TT-…* (not *TT TT-…*) and *… 1 endpoint(s) in 3x.x s*; **Average loss** (not *query loss*) *1.1%*; the packet line *Lost = 16 (1.1% loss)* and the PDF's Loss column *1.1%* beside *Breached*; the **Round-trip time** bar blue, not red; the **Packet loss** chart scaled to 4%, its bar named *PacketPulse network*. An HTTPS device ticket still reads *Average query loss*. 🛑 **Must NOT** print *1%* beside a breach of a 1% target |
  | `DIAG-019` | Customer | Every grade says why, in numbers | With no targets of the organisation's own, file a UDP test from a line around 90 ms (Bengaluru to the PacketPulse network, 14 October 2026: 89.7 ms, 0.6% loss, 20.1 ms jitter) and one over 100 ms (Dubai: 117.6 ms); open each in **History**, and in **View ticket**; **Export PDF** | **Service targets** under the results: per result its grade, name and *Degraded: round trip 89.7 ms is close to the 100 ms target.* (or *Breached: round trip 117.6 ms is over the 100 ms target.*), then a row per figure - *Round trip 89.7 ms*, a gauge marked at the target with the Degraded band shaded, *Target 100 ms or less* / *Degraded from 80 ms*, *Close to target* in amber; *Packet loss 0.6%*, *Target 1% or less*, *Within target* in green. The PDF: an **SLA** column, *SLA targets: round trip 100 ms or less; …* under the table, and under each entry *SLA: Degraded - …* with a table *Figure / Measured / Target / Degraded from / Verdict*. The loss tile, the loss bar and the PDF's Loss cell are coloured by the loss's own verdict - 0.6% is not red. A result graded before 14 October 2026 reads *Graded before PacketPulse kept the figures behind a grade.* 🛑 **Must NOT** show a grade with nothing to say why |
  | `DIAG-020` | Customer | The results table fits its window | Open a device ticket from **View ticket** on a laptop, and in **History** | The table is as wide as its card, with a scrollbar under it when its columns need more; the packet line, round trip and jitter can be scrolled to. 🛑 **Must NOT** run the columns off the dialog's edge |
  | `DIAG-017` | Customer | An ungraded result says why | In **History**, open a device UDP test filed before 14 October 2026, a device HTTPS test, and a sweep with a refused address | Each row's status reads *Not graded*, not *Unknown*; under the table *Not graded: this result was not held to your service targets…*, then one line per reason present — *Tests from a device were not graded before 14 October 2026…*, *An HTTPS test from a device is not graded…*, *PacketPulse could not test this address…* — each once however many rows share it. A ticket graded throughout has no such note. 🛑 **Must NOT** show a bare *Unknown* with no explanation |
  | `DIAG-009` | Engineer | The PDF is made from the record when it is asked for | Export a ticket's PDF twice, a minute apart; look for a stored copy on the server | Both carry the same measurements, drawn from the stored results at the moment each was asked for; no PDF is kept on the server. 🛑 **Must NOT** depend on a file kept on disk |
  | `DIAG-010` | Engineer | A colleague's test cannot be exported | As an engineer, `GET /diagnostic/<colleague's request id>/report.pdf` (and `.csv`) | 404, as though it did not exist. 🛑 **Must NOT** hand an engineer someone else's evidence |
  | `DIAG-012` | Tester | The nearest city is tested | With endpoints placed in Bengaluru and Mumbai (`SITE-007`), `POST /diagnostic/submit` with a `location` in Bengaluru | Only Bengaluru's endpoints in the result; *Endpoints: The city nearest the device, Bengaluru, N km away*; *Tested from* the phone's coordinates. 🛑 **Must NOT** sweep Mumbai |
  | `DIAG-013` | Tester | No position, nothing lost | Submit with `location` `{"status": "denied"}` | Every enabled endpoint swept; *Tested from: Unknown: location was refused on the device*; *Endpoints: Every enabled endpoint* |
  | `DIAG-014` | Engineer | An engineer cannot choose | As an engineer, `POST /diagnostic/submit` with `dns_site_ids` naming a Mumbai endpoint and a Bengaluru position | The request sweeps Bengaluru's. 🛑 **Must NOT** honour an engineer's endpoint list |
  | `DIAG-015` | Administrator | A manager may still choose | Submit with `dns_site_ids` naming one Mumbai endpoint and a Bengaluru `location` | Only that endpoint; *Endpoints: Chosen by the person who ran the test* |
  | `DIAG-016` | Engineer | The PDF says where | **Export PDF** of `DIAG-012` | Under the run details: *Tested from 12.97…, 77.75…, within N m    \|    Endpoints: the city nearest the device, Bengaluru, N km away* |
- ⚙️ **Developer Guide & Release Confidence**:
  - Engine: `packetpulsego/pkg/pingmicroservice/pingprobe/PingProbeRunner.go` (pro-bing, unprivileged ICMP, TCP fallback); jitter: `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go` (`InterarrivalJitter`, RFC 3550).
  - 🔒 Results are saved under their **own** deadline, never the sweep's: a sweep that runs long still stores everything it measured (`load` suite §4.5 proves it, and goes red with the old bug restored).
  - No screen starts a sweep: the app's sweep form was removed when **Run diagnostic** became the test from this device (2026-10-07, §4.1), and the client no longer calls `POST /diagnostic/submit`. The route stays, for the API and scripts. The case for one page holding the sweep and the device test is retired, its number left vacant.
  - Path analysis was removed in `packetpulsego/pkg/common/dbclient/migrations/0027_2026_10_05_path_analysis_removed.sql`: `ping_hop` and the fault verdict columns are no longer written, and the rows already stored stay. A client built before then still sends `trace_failures`; the server accepts and ignores it, because decoding is strict and refusing it would fail every run that client made.
  - Endpoint choice: `chooseEndpoints` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go`; the handler empties `dns_site_ids` for anyone without `dns_site_manage`. The request records `test_location_*`, `endpoint_selection` (`chosen`, `nearest`, `all`), `selection_city` and `selection_distance_km` (migration 0029).
  - Ticket figures (`avg_loss_pct`, `max_loss_pct`, `avg_jitter_ms`) are computed in `RequestFinish` from the run's counted results (`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql`).
  - The PDF is built on the fly from the database each time it is asked for (`packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`); nothing is written to disk. Its foot names the person who ran the test and the place their session checked in from.
  - Service targets: `EvaluateSla` in `packetpulsego/pkg/pingmicroservice/pingservice/PingSlaEvaluator.go` returns a check per figure, stored in `ping_result.sla_checks` (`packetpulsego/pkg/common/dbclient/migrations/0033_2026_10_14_sla_checks.sql`); the app's card, gauge and attach-card summary are `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/SlaChecks.dart` (`packetpulseflutter/test/sla_checks_test.dart`); the PDF's are `SlaHeadline`, `SlaRows` and `TargetsKey` in `packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`.
  - *Not graded*: `notGradedReason` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/DiagnosticResultView.dart` names one reason per result from its grade, method and packets sent; the note under the table lists each reason once (`diagnostic_result_selection_test.dart`).
  - Coverage: `packetpulsego/pkg/diagnosticmicroservice/**`, `packetpulsego/pkg/pingmicroservice/**`, `packetpulseflutter/test/diagnostic_result_selection_test.dart`, `packetpulseflutter/test/diagnostic_history_screen_test.dart`, the `tenancy` and `contract` suites.

---

### 3.3 🌍 IPv4 and IPv6

**Where:** every result list, the PDF, the Results API and the CSV

- 🌟 **Commercial Presentation & Sales Pitch**: A site healthy over IPv4 and dark over IPv6 is a real, common, hard-to-prove fault. PacketPulse measures both, side by side, without letting an IPv6 problem the customer did not buy an SLA for change their availability figure.
- 📖 **User Guide & Operational Flow**: A dual-stack site has two rows. The IPv6 one carries a **not counted** tag (hover for why). The PDF marks it `icmp v6 *` with a footnote.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `V6-001` | Engineer | Both families, one counted | Add a dual-stack site (e.g. `one.one.one.one`) on a host with IPv6; run | Two rows: IPv4 counted, IPv6 tagged *not counted* with its tooltip |
  | `V6-002` | Engineer | An IPv6 failure does not breach the site | A dual-stack site whose IPv6 fails while IPv4 answers | The IPv6 row is graded but not counted: the site counts as reachable, and the ticket's breach count is the IPv4's. 🛑 **Must NOT** count a breach for report-only IPv6 |
  | `V6-003` | Engineer | IPv6-only counts | A site with only an IPv6 address that fails | Counted as a failure, and graded as usual |
  | `V6-004` | Engineer | No IPv6 route is said plainly | On a server without IPv6, a dual-stack site | The IPv6 row reads *There is no IPv6 route from this vantage point, so this IPv6 address could not be tested from here.* 🛑 **Must NOT** report it as the site being down |
  | `V6-005` | Engineer | A device's UDP test is run and filed over each family | On a network with IPv6, run a UDP test of an endpoint whose reflector has an IPv6 address (`UDP-045`); attach; open the ticket; then **Network reports** for the day | Two runs, IPv4 then IPv6; **By IP version** shows both; the ticket has two rows, one tagged IPv6 reading *Reflector … at 2001:…*; the report has *UDP (reflector) IPv4* and *UDP (reflector) IPv6* rows. 🛑 **Must NOT** leave a device test with no family, or add IPv4 and IPv6 together |
- ⚙️ **Developer Guide & Release Confidence**:
  - A device's UDP test takes its family from the address the browser reached its reflector at (`ping_result.reflector_ip`, from the engine's `remoteAddress`), else from what its target resolved to: `PingResult.MeasuredIp` (`packetpulsego/pkg/pingmicroservice/pingdomain/shared/PingResult.go`).
  - `ping_result.ip_version` (4, 6 or NULL when nothing was probed) and `report_only`; the daily rollup is keyed by family (`packetpulsego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql`).
  - `probeguard.ResolveAllAndCheck` returns one checked address per family; the runner probes them in parallel and marks IPv6 report-only only when there is more than one family.

---

### 3.4 📄 CSV download

**Where:** **Export CSV** on every result · **Route:** `GET /diagnostic/{requestId}/report.csv` · **Capability:** `report_export`

- 🌟 **Commercial Presentation & Sales Pitch**: The same results, in the columns an IT system imports, in one click — and safe to open: a site name typed as a spreadsheet formula is written as text, so an export can never run code on someone else's desk.
- 📖 **User Guide & Operational Flow**: **Export CSV** saves `packetpulse-<TT>-<time>.csv` beside the PDF of the same ticket, both named and dated in the organisation's time zone. One row per site and family; empty cells (not zeros) where a probe had no reply. `checked_on` is in the organisation's time zone with its offset (`2026-10-21T11:30:00+05:30` in India): its own choice under *Scheduled reports*, else its country's. `test_latitude` and `test_longitude` say where the device was when the test ran; empty when it was not known. `resolved_ip` is what the endpoint's name resolved to, and `reflector_address` and `reflector_ip` the reflector a UDP test measured to and where the browser reached it. For an endpoint set to show its reflector only, someone who does not manage endpoints gets empty `ip_address` and `resolved_ip`.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CSV-001` | Engineer | The file has the contracted header | **Export CSV**; open it | Header exactly as in `packetpulsetest/contracts/result_export_columns.json`, `result_id` first; `checked_on` in the organisation's zone with its offset (`…+05:30` in India), never `…Z`; an unreachable row has empty round-trip cells |
  | `CSV-002` | Engineer | A formula is written as text | Name a site `=HYPERLINK("http://x","y")`; run; export; open in Excel or Sheets | The cell shows the text starting `'=`. 🛑 **Must NOT** become a live formula or link |
  | `CSV-003` | Engineer | Names in any script survive | Name a site `पुणे केंद्र`; export | The name intact (UTF-8). 🛑 **Must NOT** show mojibake |
  | `CSV-004` | Viewer | No export without `report_export` | Open a result as a Viewer | No **Export CSV** or **Export PDF**; the route answers 403 |
  | `CSV-005` | Tester | The download is on the record | Export a CSV; open **Activity** | An *Exported* entry naming the diagnostic |
  | `CSV-006` | Engineer | Where the test ran, as numbers | Export `DIAG-012`'s CSV; export one of a sweep submitted with no `location` | Six-place `test_latitude`/`test_longitude` on every row of the first; empty on the second. A longitude west of Greenwich is a plain negative number. 🛑 **Must NOT** be written as text starting `'-`, or as 0 |
  | `CSV-007` | Engineer | A UDP test names where it was measured | Run a UDP test of an endpoint named by a hostname, with a reflector; **Export CSV** | `resolved_ip` holds the address the name resolved to; `reflector_address` and `reflector_ip` the reflector and the IP the browser reached it at. 🛑 **Must NOT** leave `resolved_ip` blank for a test the server could resolve |
- ⚙️ **Developer Guide & Release Confidence**:
  - One writer for every CSV: `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`; one projection and scanner: `packetpulsego/pkg/diagnosticmicroservice/diagnosticdomain/repository/DiagnosticExportPostgres.go` (the run's organisation must match the ticket's).
  - The browser download is `text/csv` (`packetpulseflutter/lib/common/services/PacketPulseFileSaverWeb.dart`); desktop and phone use the share sheet.

---

### 3.5 🗂️ History and dashboard

**Screens:** Operate → **History**, **Dashboard** · **Routes:** `GET /diagnostic/list`, `GET /diagnostic/tt/{ttNumber}`, `GET /ping/dashboard`

- 🌟 **Commercial Presentation & Sales Pitch**: Every ticket investigated, with its loss and jitter on the row, who ran it and where — and every graded figure shown against its target, so a supplier review argues about the line, not the arithmetic.
- 📖 **User Guide & Operational Flow**: An engineer's **History** is the tests they ran, with the note *These are the tests you ran. Administrators see everyone's.* An administrator's is everyone's, and narrows by **Person**, **Place** (*Where it was run, or a site or region*), **Status**, **From** and **To**; **Clear filters** puts them back. Both search by TT number or Customer ID. Each row shows reachable/total, breaches, loss and jitter, and who ran it and where: *Customer CUST-1 · Asha Rao · at MG Road, Pune*. **Dashboard** shows the monitored endpoints (and how many are enabled), the diagnostics run, the last sweep's reachable count, the licence (days remaining, and people on it of those it covers) and the recent tickets. The 30-day availability line went with automatic monitoring (`HIST-004`).
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `HIST-001` | Engineer | Find a ticket | Search `TT-MAN-001` | Its attempts, with loss and jitter on each row |
  | `HIST-002` | Engineer | A failed search is not an empty result | Stop the API; search | The failure, with *Try again*. 🛑 **Must NOT** show *No diagnostics yet* |
  | `HIST-003` | Engineer | A device run reads as one | Attach a device test (§4.1); find it in History | Marked as measured on a device; its loss is *query* loss |
  | `HIST-004` | Engineer | No availability line | Open **Dashboard** with the browser's network panel open | No *Availability, last 30 days* card, and no request to `/monitor/trend`: the trend was automatic monitoring's, removed in October 2026 (§5.1). 🛑 **Must NOT** draw a line from tickets alone |
  | `HIST-005` | Engineer | An engineer reads only the tests they ran | Two engineers each run a test; each opens **History**; then one opens the other's by id (`GET /diagnostic/<id>`) and by ticket (`GET /diagnostic/tt/<TT>`) | Each sees only their own, with the note; by id: 404; by ticket: nothing. 🛑 **Must NOT** show an engineer a colleague's test by any route |
  | `HIST-006` | Administrator | The filters narrow | **Person** = one engineer; then **Place** = `Pune`; then **Status** = *Failed*; then **From** and **To** the same day; then **Clear filters** | Each narrows to what it names — *Place* matches the address the test was run from, or a site or region it probed; a *To* date includes the whole of that day — and **Clear filters** shows everything again |
  | `HIST-007` | Tester | A filter that cannot be read is refused | `GET /diagnostic/list?staff_id=nope`, then `?status=lost`, then `?from=yesterday` | 422, naming the field. 🛑 **Must NOT** ignore the filter and answer with everything |
  | `HIST-008` | Engineer | Who and where, on the row and the PDF | Sign in sharing your location; run a test; open **History** and the test's PDF | The row reads *· <your name> · at <the address you signed in from>*; the PDF's foot says the same. A test run without a position names the person only |
- ⚙️ **Developer Guide & Release Confidence**:
  - Row figures: `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticFigures.dart`.
  - 🔒 Whose tests come back is the server's decision, made in SQL: `readerOf` (`packetpulsego/pkg/diagnosticmicroservice/diagnosticapp/DiagnosticHandlers.go`) gives an engineer `OwnOnly` and anyone with `diagnostic_view_all` `WholeOrganisation`, and every read — list, detail, ticket, PDF, CSV — passes it to the repository's `submitted_by_user_id` predicate. The Results API reads the whole organisation, as its key does.
  - Place is the check-in address of the session the test was run from (`diagnostic_request.session_id`, `packetpulsego/pkg/common/dbclient/migrations/0026_2026_10_05_own_results_places_and_indexes.sql`) or, matched by the filter, a probed site's region or name. The same migration indexes a person's own tests and the administrator's filters.
  - Coverage: `packetpulseflutter/test/diagnostic_history_screen_test.dart`, `packetpulseflutter/test/dashboard_screen_test.dart`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticdomain/repository/DiagnosticRepositoryPostgres_test.go` (scope, who and where, every filter), `packetpulsego/pkg/diagnosticmicroservice/diagnosticapp/DiagnosticRouteHandler_test.go`.

---

### 3.6 📊 Network reports

**Screen:** Operate → **Network reports** · **Routes:** `GET /report/network`, `GET /report/network.csv`, `GET /report/network.pdf` · **Capability:** `report_view` (to read) or `report_export` (to download), and `diagnostic_view_all`

- 🌟 **Commercial Presentation & Sales Pitch**: Packet loss, latency and jitter for every reflector site — each city's endpoints — and for every site combined — *Pan India* for an operator in India — as the average, minimum and maximum, for any day, week, month or year, on screen and as CSV and PDF, and dropped on the operator's own server as each period closes (§6.2). Every figure says how it was calculated and what it covers, so a monthly review argues about the network, not the arithmetic.
- 📖 **User Guide & Operational Flow**:
  - **By reflector site / All sites combined**, **Daily / Weekly / Monthly / Yearly**, and the period button (a calendar). It opens on today so far, so a test just run is in it: the period button reads *…, so far (to 09:12)*, the PDF says *So far, to 09:12 IST*, the files are named `…-so-far`, and each CSV row's `period_end` is when it was made. Any earlier period is picked from the calendar. Days, weeks (Monday to Sunday), months and years begin at midnight in the organisation's time zone: its own choice with the scheduled reports on **Result export**, else its country's (India Standard Time for India). The zone is said under the controls.
  - **By reflector site**: one card per site — the **Region** given in **DNS sites**. Its first rows, in bold, are the site over all its endpoints (*All endpoints*); each endpoint's rows follow. Endpoints with none set are *City not set on the endpoint*. UDP tests to the PacketPulse network, with no endpoint, have a card of their own.
  - **All sites combined**: every site as one, then the period before, a trend hour by hour (a day), day by day (a week or a month) or month by month (a year), quiet steps included, and the five sites that lost the most packets.
  - Each figure as its **average, minimum and maximum**: loss (packets lost over packets sent across the row, so a long test counts for more; minimum and maximum of single tests), latency (the round trip: the average weighted by replies; minimum and maximum of single replies) and jitter. HTTPS rows show no loss.
  - One method and family a row (UDP to an endpoint's reflector, UDP to the PacketPulse network, ICMP or TCP from the device or the server, HTTPS; IPv4 or IPv6): no row adds one to another. A row on fewer than 5 tests is marked `*`.
  - Under each report, what it covers: results from devices and from the server, over IPv4 and over IPv6.
  - **Export CSV** and **Export PDF** save the report on screen. Every CSV row starts with the organisation's name, so a file forwarded or merged with another still says whose it is.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `RPT-001` | Administrator | A site's figures add up | Two device UDP tests of Mumbai endpoints — one losing 10 of 1000 packets, one 5 of 10 — and one of a Pune endpoint; open **Network reports** on that day | Mumbai's *All endpoints* row shows **2** tests, loss average **1.5%** (15 of 1010), minimum **1%**, maximum **50%**; each endpoint has its own row beneath. 🛑 **Must NOT** average the two percentages |
  | `RPT-002` | Administrator | A test is filed under its endpoint's site, never its tester's city | Run a server sweep (§3.2) of an endpoint whose **Region** is `Agra`, from a sign-in in Pune | An **Agra** card with *ICMP (server)*; no Pune card. 🛑 **Must NOT** file a measurement under where its tester stood |
  | `RPT-003` | Administrator | Methods and families are never added together | At one site, a device UDP test over IPv4, one over IPv6 (`V6-005`) and a device HTTPS test | Three rows: *UDP (reflector) IPv4*, *UDP (reflector) IPv6*, and the HTTPS row with loss `—`. 🛑 **Must NOT** pool a UDP and an HTTPS figure, or IPv4 and IPv6 |
  | `RPT-004` | Administrator | A day is the organisation's day | An organisation in India with no zone chosen; a test at 00:30 IST on the 21st and one at 23:30 IST on the 20th; open the 21st | Only the first; *Days begin at midnight, Asia/Kolkata.* 🛑 **Must NOT** cut the day in UTC or the device's zone |
  | `RPT-005` | Administrator | Every site combined | **All sites combined** | Title *Pan India network performance*; one row a method and family over every site; *The period before: …* beneath |
  | `RPT-006` | Administrator | A week beside the week before | **All sites combined**, **Weekly** | The period before's rows; seven day rows per method, a quiet day as `0` tests; the five worst sites by loss |
  | `RPT-007` | Administrator | A year, month by month | **Yearly** on each view | The year's sites and their endpoints; *Month by month* with twelve rows a method; files `packetpulse-city-yearly-<year>.csv` and `packetpulse-pan-india-yearly-<year>.pdf` |
  | `RPT-008` | Engineer | An engineer is not shown everyone's figures | As a NOC Engineer (`diagnostic_view_own`); look for **Network reports**; then `GET /api/v1/report/network` | Not in the menu; 403. 🛑 **Must NOT** show a colleague's tests added up |
  | `RPT-009` | Viewer | Reading without downloading | As a Viewer (no `report_export`) | The report, and no **Export CSV** or **Export PDF** |
  | `RPT-010` | Administrator | The files are the report on screen | **Export CSV**, then **Export PDF** | `packetpulse-city-daily-<day>.csv` under the contracted header (`packetpulsetest/contracts/network_report_columns.json`), the organisation's name in the first column of every row, each figure's `_avg`, `_min` and `_max`, and the PDF with the same figures, the organisation, the zone and the definitions; *Pan India* files are `packetpulse-pan-india-…` |
  | `RPT-013` | Administrator | A file says whose it is | Merge two organisations' CSVs in a spreadsheet; filter on **organisation**; then name an organisation `=SUM(1,1)` and download again | Each row names its own organisation; the formula-like name shows as text (`'=SUM(1,1)`). 🛑 **Must NOT** run a name as a formula |
  | `RPT-014` | Administrator | A test just run is in the report | Run a UDP test; open **Network reports** | Today, *so far (to <now>)*, with the test in its site; **Export CSV** saves `packetpulse-city-daily-<today>-so-far.csv`. Pick yesterday: the whole day, no *so far*. 🛑 **Must NOT** open on yesterday, or name a period still under way as if it were whole |
  | `RPT-011` | Administrator | A failed load is not a quiet day | Stop the API; open **Network reports** | The failure, with *Try again*; no *No tests in this period*, and the export buttons disabled |
  | `RPT-012` | Tester | A query that cannot be read is refused | `GET /report/network?period=fortnight`, `?view=town`, `?date=21-10-2026`, a date in the future | 422 naming the field; the CSV route answers the same, never with a file |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/reportmicroservice/`: `reportstats` holds every rule — periods in the zone (`ReportPeriod.go`), weighting, minimum and maximum, and grouping by site, endpoint, method and family (`ReportStats.go`); `reportfile` writes the CSV and PDF; `reportservice` resolves a query and runs the scheduled drops; `reportdomain/repository` reads a period's results with their endpoint's region and their family.
  - 🔒 One organisation predicate on `ping_run`, and the test (`diagnostic_request`) and the endpoint (`dns_site`) are joined *within* that organisation, so a row naming another tenant's run or endpoint describes nothing.
  - The organisation's zone: `report_settings.time_zone`, else its country's (`packetpulsego/pkg/common/orgzone/OrgZone.go`); the same zone writes a ticket's PDF and CSV and the result export.
  - Coverage: `packetpulsego/pkg/reportmicroservice/**`, `packetpulseflutter/test/network_report_screen_test.dart`, `packetpulsetest/golang/apicontract/network_report_test.go`, `packetpulsetest/golang/tenancyisolation/tenancy_isolation_test.go`.

---
## Group 4 — The Customer's Side

Measuring from where the person is, not from where PacketPulse is.

---

### 4.1 📱 From this device

**Screen:** Operate → **Run diagnostic** · **Routes:** `GET /dnssite/list` (`dns_site_view`), `POST /diagnostic/clientobservation` (`diagnostic_run`, licensed), and `GET /organisation/settings` for the test with no ticket; that test's target is set with `PUT /organisation/settings` (`staff_manage`), not on this screen

- 🌟 **Commercial Presentation & Sales Pitch**: "Is it slow for me?" answered from the customer's own connection, to the customer's own endpoint: the engineer opens the page and the endpoint nearest them is already chosen, from the position entered for each endpoint. It shows the device's IPv4 **and** IPv6 addresses, then files the run against the ticket, under the endpoint it measured — labelled as measured on a device.
- 📖 **User Guide & Operational Flow**: **Run diagnostic** is the test from this device, filed against a ticket (since 2026-10-07 the page no longer sweeps from the server). The ticket comes first: **Customer ID**, **TT number**, optional **Notes** and the **Endpoint**; **Start test** does nothing until the customer, the TT number and the endpoint are given, and marks the one missing. As the page opens the device is asked where it is, and the endpoint nearest it — by the latitude and longitude entered under DNS sites (§3.1) — is chosen, with *Nearest to this device, N km away*. An engineer is given that endpoint and cannot change it (*Chosen for you: the endpoint nearest where this device is.*). Whoever manages the endpoints (`dns_site_manage`) may pick another, and go back with **Choose by my location**. With no position to go by — location refused or unknown, or an enabled endpoint not placed yet — everyone picks by hand, and the page says why. Every enabled endpoint must have a position before one is chosen for anyone (since 2026-10-17): chosen from the placed ones alone, a test went to the one placed endpoint, however far, while a nearer one waited for its position. Choosing another endpoint empties a finished run, and stops one under way. The position chooses; it is filed with a run only for people whose location is recorded (§2.1). Someone who may not run diagnostics, or has no organisation, gets the device test alone, with no ticket, against the organisation's internal host or Cloudflare DNS, which the screen does not name. The line-speed test is hidden. On the public demo the **Customer ID** and **TT number** open filled in with the example the fields show, `CUST-88412` and `TT-2026-00731`, so a walkthrough goes straight to **Start test**; they can be changed, and what is in them is what is filed (since 2026-10-09).
  - Above the controls, three cards: the **IPv4** address, the **IPv6** address (or *No IPv6 connectivity*) and the **Nearest edge** that answered, with an *IPv6 active* or *IPv4 only* badge. No network operator or AS number is shown: finding one would send the engineer's address to a third party.
  - In a browser the test is **UDP, like a call** (§4.2), and nothing else: since the review of 13 October 2026 there is no **Measurement method** choice and no *HTTPS to the endpoint* beside it. The app on a desktop or phone sends a real ICMP echo where the device allows it, and offers no UDP test. Only a browser without WebRTC falls back to an HTTPS request to the endpoint's address; an endpoint that does not serve HTTPS — most routers and resolvers do not — then shows every sample unanswered, even when it is up, and the note beside the figures says so.
  - Below the controls, the route says what the test goes to (since 2026-10-09): *Start will test*, then *Testing now*, then *This run tested*, over *This device → method → destination*. With a ticket it is marked **Path to the endpoint** (green, a tick) or **Not the endpoint** (amber, a warning sign) and says why in one sentence: *HTTPS requests from this device to Pune Edge…*, *A real ping from this device to Pune Edge…*, or, for the UDP test, the reflector beside the endpoint or the PacketPulse network (§4.2). With no ticket it reads *The test target*, naming no host, and carries no mark.
  - **IP version** — *Auto*, *IPv4 only*, *IPv6 only* — holds a run against the default target to one address family, to compare the two paths. It is offered only on the test with no ticket, when the organisation tests the default over HTTPS: an endpoint is reached at its own address, a browser cannot choose the family a hostname resolves to, and the native ping is IPv4.
  - While a run goes, six figures fill in — average, fastest, slowest, jitter, loss and progress, and for the UDP test **Lost up · down** — with **Round-trip, sample by sample** (jitter as a band, the average dashed, a red mark per unanswered sample; point at it to read one), **Round-trips by band** (under 20, 20–50, 50–100, 100–200, 200 ms and over) and a **Probe log** of every reply. **Clear results** empties the figures and charts and keeps the log; **Clear log** empties the log. No standard deviation is shown, here or on the public self-test (review of 13 October 2026): beside jitter it read as a second figure for the same thing. It is still measured and filed. Nor is the voice-quality estimate (MOS), on any screen or report or on the self-test (review of 13 October 2026): it is still computed, filed, graded against a MOS target and exported in the CSV, and one switch shows it again — `PacketPulseConfig.voiceQualityShown` in the app, `ReportOptions.ShowVoiceQuality` for the PDF.
  - **Attach to a ticket** files the finished run against the ticket at the top of the page and says *Attached to ticket TT-…* beside the button, with **View ticket** (since 14 October 2026). It opens the ticket as **History** shows it — its results and grades, the notes, **Export PDF** and **Export CSV** — over the test, from what the server returned on filing, so nothing more is fetched; on a phone it takes the whole screen. Closing it leaves the run on screen as it was. Running again clears the confirmation and the button with it.
  - **Notes**, typed at the top of the page, are filed with the run and shown with the ticket in History (*Notes* beside *Tested from*) and on its PDF (*Notes: …* under the run's details). Until the review of 13 October 2026 they were stored and shown nowhere.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DEV-001` | Viewer | The target is internal | Open **Run diagnostic** (no ticket for a Viewer); run a device test | No *Testing against* panel, no **Change target**, no dropdown or free host field, and no *Target* line in *What this run found*. The address card says the address came from *a public address lookup over the same connection*. 🛑 **Must NOT** name a host anywhere on the page: no *1.1.1.1*, *Cloudflare DNS* or the organisation's own |
  | `DEV-002` | Administrator, then Viewer | The internal target is still what is measured | As the administrator, `PUT /organisation/settings` with `device_test_target` `dns.google`; as a Viewer, open **Run diagnostic** and run a device test with the browser's network panel open | Every probe goes to `dns.google`, and the screen names no host. 🛑 **Must NOT** fall back to Cloudflare while a target is set |
  | `DEV-003` | Administrator | The target saves alone | `PUT /organisation/settings` with a new `device_test_target` and the current `require_checkin_location` | The *Require location* rule (§1.2) is unchanged. 🛑 **Must NOT** reset another setting |
  | `DEV-004` | Engineer | Both addresses | Run from a dual-stack connection; then from one without IPv6 | *IPv4 a.b.c.d · IPv6 2401:…*; then *No IPv6 connectivity* |
  | `DEV-005` | Engineer | The speed test is hidden | Open **Run diagnostic**; then open a ticket that had a speed filed before the test was hidden | No **Measure speed** button and no ~12 MB note. The older ticket still shows its line speed. 🛑 **Must NOT** move any data to the speed-test endpoints |
  | `DEV-006` | Engineer | Attach to a ticket | Type a Customer ID and TT `TT-MAN-002` at the top of the page; keep the endpoint chosen; **Start test**; **Attach to a ticket** | The ticket shows the device run, *measured on a device*, named after the endpoint with its address, and no line speed; in a browser it is the UDP test, so loss is *packet* loss and the run is graded against the organisation's targets (§5.1). Its **Export PDF** and **Export CSV** name it the same way. A device run filed before 2026-10-07, with no endpoint, still reads *Device test* with no address. 🛑 **Must NOT** show the internal target's host for a run that had no endpoint |
  | `DEV-007` | Viewer | Only people managers change the target | As a Viewer, `PUT /organisation/settings` | 403. No **Change target** on the screen, for anyone |
  | `DEV-008` | Viewer | An unreadable setting does not block the test | Stop the API after the screen loads; reopen it | The test still runs, against Cloudflare DNS, and no error about the target is shown |
  | `DEV-009` | Superuser | No organisation, still a test | Sign in as the platform superuser; open **Run diagnostic** | The device test alone, with no ticket fields, testing Cloudflare DNS. It needs no organisation and no licence |
  | `DEV-010` | Engineer | The figures fill in as the run goes | Set 25 samples and 500 ms interval; **Start test** and watch | **Progress** counts up (*4 of 25 samples*, …) and the average, fastest, slowest, jitter and loss change with each answer; at the end *100%*, *25 of 25 samples*. 🛑 **Must NOT** show zeros before the first answer — dashes until then |
  | `DEV-011` | Viewer | IPv6 only | On a dual-stack connection choose **IP version → IPv6 only**; run with the browser's network panel open; then repeat on a connection with no IPv6 | Every probe goes to the resolver's IPv6 address, and the log reads *Started: DNS over HTTPS · IPv6 only, …*; with no IPv6, every sample is *No answer* and loss is 100%. Choosing another version clears the run on screen. 🛑 **Must NOT** name the address or the resolver anywhere on the page |
  | `DEV-012` | Administrator, then Viewer | No switch where it cannot hold | Open **Run diagnostic** as the administrator; then `PUT /organisation/settings` with `device_test_target` `dns.google` and open it as a Viewer | No **IP version** control for the administrator's endpoint, nor for the Viewer's host, and *How this test works* has no *One family at a time* point. On the desktop app with **ICMP echo** chosen, no **IP version** either. 🛑 **Must NOT** offer a choice the run would ignore |
  | `DEV-013` | Engineer | The charts | Run 25 samples; point at the round-trip chart | **Round-trips by band** counts add up to the answered samples; pointing reads *#n*, the round-trip and the jitter so far, or *unanswered* for a gap. 🛑 **Must NOT** join the line across an unanswered sample |
  | `DEV-014` | Engineer | The probe log | Set **Timeout** to 200 ms on a slow line; run; **Clear results**; then **Clear log** | Each line is stamped *[HH:MM:SS]*: replies in green, *No answer: seq=n, timed out after 200 ms* in red, then *Finished. Queries: Sent = …* (amber when anything was lost). **Clear results** keeps the log; **Clear log** empties it to *Nothing logged yet.* |
  | `DEV-015` | Engineer | The address cards | Open from a dual-stack connection; then from an IPv4-only one | *IPv6 active* with both addresses and the nearest edge; then *IPv4 only* with *No IPv6 connectivity*. 🛑 **Must NOT** show an ISP or AS number, or contact `ipinfo.io` |
  | `DEV-016` | Engineer | A filed run says where it was taken | With location recorded for the engineer, run on a phone and **Attach to a ticket** | The ticket shows *Tested from* the phone's coordinates as it was filed, beside the sign-in place. With location refused: *Unknown: location was refused on the device* |
  | `DEV-017` | Engineer | The nearest endpoint is chosen | With endpoints placed in Pune and Mumbai (`SITE-007`), open **Run diagnostic** on a phone in Pune, allowing location | **Endpoint \*** shows the Pune endpoint with *Chosen for you: the endpoint nearest where this device is.* and *Nearest to this device, N km away*; no endpoint chips. **Start test** probes the Pune endpoint (network panel). 🛑 **Must NOT** offer an engineer another endpoint while a position chose one |
  | `DEV-018` | Administrator | A manager may choose another | As in `DEV-017`, as the administrator; pick Mumbai; run; then **Choose by my location** | Every enabled endpoint as a chip, Pune chosen; Mumbai can be picked and is what is probed; **Choose by my location** goes back to Pune and empties the Mumbai run on screen |
  | `DEV-019` | Engineer | No position, choose by hand | Refuse location in the browser; open **Run diagnostic**; then, with location allowed, in an organisation whose endpoints all came from **Bulk import** (no positions) | Chips with no endpoint chosen and *This device's location is not available, so choose the endpoint to test.*; then *None of your endpoints has a position yet, so choose the endpoint to test.* |
  | `DEV-028` | Engineer | One endpoint not yet placed means everyone chooses | With Pune and Mumbai placed, **Bulk import** a third endpoint (no position); open **Run diagnostic** on a phone in Pune, allowing location | No endpoint chosen; chips for all three; *Not every endpoint has a position yet, so choose the endpoint to test.* The engineer may pick any one, and that is what is probed. Give the third a position and reopen: Pune is chosen again, with no chips. 🛑 **Must NOT** choose the nearest of only the placed endpoints, or lock an engineer to it |
  | `DEV-020` | Engineer | Nothing runs without the ticket | With location refused, press **Start test** with every field empty | *Enter the Customer ID.*, *Enter the TT number.* and *Choose the endpoint to test.*, scrolled into view; the network panel shows no probe. 🛑 **Must NOT** test the internal target in the endpoint's place |
  | `DEV-021` | Engineer | A browser tests UDP, never HTTPS | In a browser, open **Run diagnostic**, fill the ticket, and run against an endpoint that answers ping but serves no HTTPS; then the desktop app | The browser: no **Measurement method** row, the UDP test with **Test length** and **Packets per second**, and figures from the reflector. The desktop app with **ICMP echo** answers. 🛑 **Must NOT** offer *HTTPS to the endpoint*, or **Run the HTTPS test** after a failure |
  | `DEV-023` | Engineer, then Viewer | The route says what is tested | In the desktop app read the route; **Start test**; then sign in as a Viewer (no ticket) | *Start will test*, *This device → ICMP echo → Pune Edge · its address*, **Path to the endpoint** and *A real ping from this device to Pune Edge…*; *This run tested* when it ends. In a browser the route is the UDP test's (`UDP-027`). The Viewer: the method and *The test target*, no mark. Choosing another endpoint changes the destination at once. 🛑 **Must NOT** name the internal target or a resolver in the route |
  | `DEV-024` | Engineer | The notes reach the report | Type *Calls drop every evening* in **Notes**; run; **Attach to a ticket**; open the ticket in **History**; **Export PDF** | *Notes  Calls drop every evening* beside *Tested from*; the PDF prints *Notes: Calls drop every evening* under the run's details. A ticket filed with no notes shows no Notes line on either. 🛑 **Must NOT** drop the notes the engineer typed |
  | `DEV-025` | Engineer | No standard deviation | Run a test; look at the figures, then at `/selftest.html`'s jitter tile | No *Standard deviation* figure and no *Spread of all round-trips*; the self-test's jitter caption is *Variation between consecutive replies*. 🛑 **Must NOT** show a standard deviation anywhere a person reads |
  | `DEV-026` | Engineer | No voice-quality estimate | Run a UDP test; read the figures; **Attach to a ticket**; open it in **History**; **Export PDF**; **Export CSV**; open `/selftest.html` and run a resolver | Six figures, no *Voice quality (MOS)* tile; no *MOS* column on the ticket or in the PDF; the self-test shows no MOS tile. The CSV still has `mos_score`. A MOS target still grades: a result under it is *Breached*. 🛑 **Must NOT** show MOS on any screen or report |
  | `DEV-027` | Engineer | The ticket from where it was filed | Fill the ticket with notes; run; **Attach to a ticket**; **View ticket**; **Export PDF**; close it; then on a phone | *Attached to ticket TT-…* and **View ticket** beside **Attach to a ticket**; the ticket opens over the test with its results, grade, notes and export buttons, and opens at once with no request in the network panel; the PDF downloads from there; closing it leaves the run's figures and log as they were. On a phone it fills the screen, its export buttons under its heading. Before a run is filed there is no **View ticket**. 🛑 **Must NOT** leave the page, or lose the run, to show the ticket |
  | `DEV-022` | RIVAL tester | Another organisation's endpoint is refused | `POST /diagnostic/clientobservation` with ACME's `dns_site_id`; then with a made-up one | 422 `validation_failed` on `dns_site_id`, the same for both, naming neither; nothing is filed under the ticket. 🛑 **Must NOT** file a run under, or name, another organisation's endpoint |
- ⚙️ **Developer Guide & Release Confidence**:
  - The page: `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart`, with the ticket in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticTicket.dart` and the endpoint in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticEndpointField.dart`. The nearest endpoint is `nearestEndpoint` in `packetpulseflutter/lib/dnssitemicroservice/domain/shared/DnsSite.dart`, the same haversine on the same radius as the server's `DistanceKm`. An endpoint's probe target is `ClientProbeTarget.endpoint` in `packetpulseflutter/lib/diagnosticmicroservice/service/ClientProbeEngine.dart`.
  - Filing: the client sends `dns_site_id`. `ClientObservationRecord` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go` reads it among the caller's organisation's **enabled** endpoints only, refuses anything else on `dns_site_id` (422), and stores each result against the endpoint, named as the inventory names it. The server does not demand one: a client older than the change still files, as the internal target.
  - A browser cannot send ICMP, so the web build measures the endpoint with UDP over WebRTC (§4.2) — over HTTPS only where WebRTC is missing — and DNS-over-HTTPS (Cloudflare) or an HTTPS reach to the organisation's host on the test with no ticket; the app on a desktop or phone can also ping. Jitter is RFC 3550. The standard deviation is computed and filed (`stddev_rtt_ms`) but shown nowhere: `packetpulseflutter/test/client_probe_live_panels_test.dart` holds the page to it.
  - Notes: filed as the request's `notes`, read back by `DiagnosticRequest.fromJson` in `packetpulseflutter/lib/diagnosticmicroservice/domain/shared/Diagnostic.dart`, shown by `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/DiagnosticResultView.dart` and printed by `ReportOptions.Notes` in `packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`; `TestTheTicketsNotesArePrintedOnItsReport` and `packetpulseflutter/test/diagnostic_result_selection_test.dart` hold them there.
  - Voice quality hidden: the MOS tile, the results' MOS column and the PDF's are kept and tested behind `voiceQualityShown = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart` (`packetpulseflutter/test/client_probe_live_panels_test.dart`, `packetpulseflutter/test/diagnostic_result_selection_test.dart`) and `ReportOptions.ShowVoiceQuality` in `packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`, whose hidden column's width goes to the site (`PingReportPdfBuilder_test.go`). `EvaluateSla` still grades `mos_score`.
  - View ticket: `_viewTicket` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart` shows the `DiagnosticDetail` the filing returned in a dialog (full screen under 720 px) through `DiagnosticResultView`, the view **History** uses; held by `packetpulseflutter/test/client_probe_run_test.dart`. Below 560 px `PacketPulseSectionHeader` puts its action under the heading, which keeps the ticket's export buttons on a phone screen.
  - Target hidden: the *Testing against* panel, the *Target* line and the **Change target** dialog are kept and tested behind `testTargetShown = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`, as the speed test is.
  - A filed device run with no endpoint keeps its host in `ping_result`. It is replaced when read back — the diagnostic detail, the Results API's ticket, the PDF, and every CSV (ticket, pulled, scheduled push) — by `hideDeviceTargets` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go` and `Row.shown` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`; a result with a `dns_site_id` measured the organisation's own endpoint and is shown as stored. The raw `/ping/run/{runId}` routes, which the app does not call, still show it. The page's wording is held to naming no host by `TestTheRunDiagnosticPageNamesNoDeviceTarget` in `packetpulsego/pkg/initmicroservice/initconstants/PacketPulseStrings_test.go`.
  - Target normalisation (pasted URL → host, lower case, canonical IPs, zones refused): `NormaliseTestTarget` in `packetpulsego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`.
  - Speed, hidden: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientThroughput.dart` (Cloudflare `__down`/`__up`, median after a warm-up) is kept and tested behind `speedTestEnabled = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`; turning it back on is that one line. Egress: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientEgress.dart`.
  - A UDP run's progress and **What happened**: `_progressDone`, `_progressCaption` and `_udpAccount` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart`, counting against the engine's plan (`UdpPlan` in `packetpulseflutter/lib/diagnosticmicroservice/service/udp/UdpProbeRun.dart`, passed from the engine's `measuring` phase by `packetpulseflutter/lib/diagnosticmicroservice/service/udp/UdpProbeRunnerWeb.dart`). Coverage: `packetpulseflutter/test/run_diagnostic_udp_test.dart`, `packetpulseflutter/test/udp_probe_run_test.dart`.
  - The route: `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientProbeRoute.dart`, built by `_route` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart`. Before a run it is foreseen from the endpoint; a UDP run's comes from the server's answer the moment it arrives. Its strings are held to naming no host by the same Go test. Coverage: `packetpulseflutter/test/run_diagnostic_screen_test.dart`, `packetpulseflutter/test/run_diagnostic_udp_test.dart`, `packetpulseflutter/test/client_probe_run_test.dart`.
  - Live panels: the address cards in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientEgressPanel.dart`; the timeline, the bands, `runningJitter` and `latencyBands` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientProbeCharts.dart`; the log in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/ClientProbeConsole.dart`. The figures are `ClientProbeSummary.from` over the samples so far, so the live numbers and the filed ones are one calculation. **IPv6 only** is `ClientProbeTarget.resolverV6` (`ClientProbeTarget.defaultFor`) in `packetpulseflutter/lib/diagnosticmicroservice/service/ClientProbeEngine.dart`; a filed run keeps the IPv6 address in `ping_result`, hidden on read like any device target. The new strings are held to naming no host, the IPv6 literals included, by the same Go test.
  - Coverage: `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/EndpointDeviceRun_test.go`, the `contract` suite's `client_observation_test.go`, `packetpulseflutter/test/dns_site_nearest_test.dart`, `packetpulseflutter/test/client_probe_screen_test.dart`, `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/client_probe_live_panels_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart`, `packetpulseflutter/test/https_probe_transport_test.dart`, `packetpulseflutter/test/client_throughput_test.dart`, `packetpulseflutter/test/client_egress_test.dart`.

---

### 4.2 📶 UDP test and reflectors

**Pages:** Operate → **Run diagnostic**, in a browser · **Routes:** `POST /reflector/session` (public, limited per address; no page calls it since the self-test's UDP option was removed on 17 October 2026), `POST /diagnostic/endpoint/{dnsSiteId}/webrtc-offer` (`diagnostic_run`, licensed), `POST /diagnostic/endpoint/{dnsSiteId}/reflector-key` (`dns_site_manage`, licensed, audited)

- 🌟 **Commercial Presentation & Sales Pitch**: The only browser test that sees real packet loss, and says which way it was lost. Run from **Run diagnostic** where the customer is, in 30 seconds it tells the engineer whether the customer's uplink drops packets, their downlink does, or packets arrive too late to play — the three causes of a choppy call, told apart from where the customer sits.
- 📖 **User Guide & Operational Flow**: The UDP test runs in **Run diagnostic** (below). It sends small UDP packets at the chosen rate (50 a second, 200 bytes each, for a call) to a reflector. The public self-test offered it too, as **PacketPulse network — UDP, like a call**, until 17 October 2026, when it was removed with the site's links to the self-test: it needed PacketPulse's own reflector, which the live host keeps off. Loss is counted at the end, per packet, when the reflector says which packets reached it; until then the loss tile shows a dash. A packet back more than a second after it was sent is *late*, and counts in the loss figure, because a call would have played silence. *Not sent (device busy)* counts slots the browser could not keep; they are left out of every percentage. Keep the tab in front: a hidden tab stops the test.
  - When UDP cannot get through, the test says *UDP is blocked on this network*, and never shows that as 100% loss. It says so only after a second, fresh connection got nothing through either (since 14 October 2026): a first connection that gets no answer is retried at once, with *No UDP came back on that connection; trying once more on a fresh one.* in the log, and a run that then measures logs *The first connection got no UDP through and the second did…* beside its figures. A browser that sent no UDP at all — a setting, an extension or a VPN keeping WebRTC off it — is told *This browser sent no UDP, so the network was not tested…*, not that the network blocks UDP.
  - The reflector logs how each session ended (`/opt/packetpulse/logs/packetpulse.log`, *reflector session ended*): `connected`, `checks_received` (the browser's connectivity checks that reached it), `answers_sent`, `probes_received` and `lifetime_ms`. For a run told *UDP is blocked*, `checks_received=0` means the network dropped the browser's packets on the way in; checks received and answered with `connected=false` means the answers were lost on the way back. It was added for a laptop in Dubai that was intermittently told UDP was blocked while one in Bengaluru was not.
  - A customer's own reflector (`packetpulsereflector`, on a PoP, an SBC's neighbour or a VPS) is installed with its endpoint's key: `POST /diagnostic/endpoint/{dnsSiteId}/reflector-key` returns the key and the install line, `sudo ./install-reflector.sh --key <key> [--public-ip <address>]`. Sessions on it are then opened only through `POST /diagnostic/endpoint/{dnsSiteId}/webrtc-offer`, which signs the request and checks that the answer is signed back.
  - **A customer release is node-locked** (since 16 October 2026). It is built per customer by `packetpulserelease/reflector/BuildReflectorRelease.sh` into `packetpulserelease/<Customer>/<yyyy-mm>/packetpulse-reflector-<version>/` (filed by the month of its UTC build time; `<Customer>/RELEASES.md` lists every build and `<Customer>/current` links to the one to send), with a licence (`packetpulsereflector.licence`) that names the public IP(s) the reflector may run at, each sealed to that build so the file does not show it, and signed with the release machine's key (`~/.packetpulse/reflector-keys`, outside the repository). The release binary will not start without the licence, refuses `-open`, starts only when its `-public-ip` (or an interface address) is one the licence names, refuses any session the relay sent to another address (403), warns daily in the licence's last 30 days and stops when it ends (exit 2). `packetpulsereflector -check` reports all of it without serving, and the installer runs it before starting anything. The install line becomes `sudo ./install-reflector.sh --key <key> --licence packetpulsereflector.licence --public-ip <licensed address> --server-ip <PacketPulse server>`; it also opens firewalld, warns about an unsynchronised clock and installs `reflector-diagnostics.sh`, which collects a support bundle with the key redacted. Each delivery carries `PacketPulse_Reflector_Installation_Guide.pdf`, printed from the customer edition, `packetpulserelease/reflector/docs/PacketPulse_Reflector_Installation_Guide_Customer.html`: installation only, the engineer's account only, the licence in one line, since PacketPulse sets up the DNS site and sends the endpoint key itself (review of 17 October 2026). The full guide, `packetpulserelease/reflector/docs/PacketPulse_Reflector_Installation_Guide.html`, is printed into each build's `vendor/` as PacketPulse's reference and never shipped; `release.sh guide` prints both again into an existing build.
  - **Across the layers**: a signed request the reflector refuses is logged there with its cause (*reflector refused a session request*, `reason=` clock with `clock_offset_s`, key, replay or unsigned), because the server sees all of them as *installed with a different key*. The server logs every failed relay (*reflector relay failed*: `endpoint_id`, `reflector_address`, the dial error and a `hint` naming the layer to check), and a node-locked reflector reached at an unlicensed address answers 403, shown as 409 *The reflector on that endpoint is not licensed for the address PacketPulse reached it at…*. Every reflector error on screen says what happened, then the next step, then *Details for your engineer*: the address dialled, the time (UTC) and what came back — a timeout (a firewall drop), a refusal (nothing listening), a failed lookup, or the reflector's HTTP status. The **Reference** under it is the request ID, and the server's *reflector relay failed* line carries it as `reference=` with `reason`, `reflector_status` and `waited_ms` (`ReflectorRelayError` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticReflectorRelay.go`; the wording in `reflectorMessage`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticapp/DiagnosticHandlers.go`).
  - **In Run diagnostic** (since 2026-10-09), in a browser with a ticket, the test is **UDP, like a call**, the only method offered, with **Test length** 10, 30 or 60 seconds and **Packets per second** 10, 25, 50 or 100 (a call's 50 unless chosen; since 13 October 2026) in place of the HTTPS steppers. The note under the figures says the rate run: *…100 packets a second from this browser…*. It is set up through the signed-in relay, so the person's session caps it, not the office's address. It measures to the reflector beside the ticket's endpoint when the endpoint has a **UDP reflector address** (§3.1), and otherwise to the PacketPulse network. The route below the controls (§4.1) says which before **Start test**, while the test runs and after it: *UDP over WebRTC → The reflector beside Pune Edge*, **Path to the endpoint**; or *UDP over WebRTC → The PacketPulse network*, **Not the endpoint**, in amber, with *Pune Edge has no UDP reflector, so this test goes to the PacketPulse network instead…* — or *…rather than to Pune Edge…* when the PacketPulse network was chosen after the endpoint's reflector failed, or when the endpoint's reflector address is PacketPulse's own. The route follows the server: once it answers, the route shows the reflector that answered, and the probe log says *Answered by the PacketPulse network, not Pune Edge* or *Answered by the reflector beside Pune Edge*. *What this run found* says which: *The reflector beside Pune Edge*, or *The PacketPulse network: the customer's access line, not the path to the endpoint* — and then no endpoint line. **Lost up · down** and *n late* sit with the figures; *Where packets went* reads, for example, *Upstream lost 4 (0.3%); Downstream lost 2 (0.1%); Late 4; Jitter up 9.4 ms, down 4.6 ms*. **Attach to a ticket** files it with the split; the ticket's page and PDF print the same line. A test to the PacketPulse network is filed as *PacketPulse network* at PacketPulse's address, never under the endpoint.
  - **Progress** counts the run's packet slots against what the reflector allows, not against what was asked for: *Testing 383/1500* and *Packet 383 of 1500* while it runs. A slot this device was too busy to keep is skipped, not sent late, and still counts, so a finished run reads **100%** with *1492 sent · 8 skipped* (since 2026-10-09; it used to stop at *99%*). Under the figures, **What happened** says the run in words: *The test ran its full 30 s, with room for 1500 packets at 50 a second. This device sent 1492. The other 8 were skipped: the device was too busy to send them on time… Skipped packets are not counted as lost, and say nothing about the network. All 1492 came back in time.* A stopped run says *The test was stopped part-way: n of its 1500 packets had their turn.*; a run the reflector's limits cut down adds *That is what the reflector's limits allowed, rather than the 1500 packets asked for.* The log adds *Pacing: 1500 packets, 50 a second* and, when any were skipped, *Skipped: 8 of 1500 packets, this device was busy*.
  - When the endpoint's reflector cannot run the test — down, busy, or installed with another key — the page shows the server's reason and offers **Test to the PacketPulse network instead**. *UDP is blocked on this network* is said only when not one UDP packet got through — the browser's ICE checks never answered within 8 s — and offers **Try again**; there is no HTTPS test in its place. When UDP got through but the connection over it did not open within 14 s — DTLS and SCTP shake hands after ICE, and on a lossy line each lost handshake packet waits out a retransmission timer — the page says *UDP got through, but the connection over it could not be completed…* instead, also with **Try again**. Until 13 October 2026 both were called blocked after 8 s. A refusal for who the person is (signed out, no permission, no licence) offers nothing else. **Stop** ends sending early; the packets already sent are still waited for, so a stopped run is a shorter run, not a lossy one.
  - A filed UDP test is graded against the organisation's targets like a sweep (§5.1), and a breach is counted on the ticket.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `UDP-001` | Engineer | A clean line reads clean | In **Run diagnostic**, in a browser, with a ticket, choose an endpoint with a **UDP reflector address** (§3.1); 30 seconds at 50 a second; **Start test** | **Lost up · down** *0 · 0* with *0 late*; *What happened* ends *All 1500 came back in time.* (fewer, the rest skipped, on a busy device). 🛑 **Must NOT** show any downstream loss on a clean line (the first cut showed 4.8%) |
  | `UDP-002` | Engineer | Loss is placed by direction | Drop 2% of outbound UDP to the reflector (macOS `dnctl pipe 1 config plr 0.02` with a `pfctl` rule on `out proto udp to port 50000`; Linux `tc qdisc add dev <if> root netem loss 2%`); run 30 s in **Run diagnostic**; repeat dropping inbound instead | About 2% lost, all of it in the first figure of **Lost up · down** and under *Upstream lost* in *Where packets went*, none down; then the same the other way. 🛑 **Must NOT** split one direction's loss across both |
  | `UDP-003` | Engineer | Late is not lost | Add 1500 ms delay to inbound UDP for a few seconds of a **Run diagnostic** run (`dnctl pipe 1 config delay 1500`) | Those packets in the *n late* caption and under *Late* in *Where packets went*, not *down*; the loss figure includes them |
  | `UDP-004` | Engineer | UDP blocked | Block outbound UDP to port 50000 (a `pf` rule, or a network that blocks UDP); run in **Run diagnostic** | *UDP is blocked on this network…* with **Try again**, only after a second, fresh connection also got nothing through; no loss figure. 🛑 **Must NOT** show 100% loss |
  | `UDP-005` | Engineer | PacketPulse's reflector switched off | On a server with `REFLECTOR_UDP_PORT=0` (the live host since 16 October 2026), run the UDP test in **Run diagnostic** for an endpoint with no **UDP reflector address** | *PacketPulse's own reflector is switched off on this server, so a test to the PacketPulse network cannot run…*, naming the endpoint's reflector address or PacketPulse support as the way on. 🛑 **Must NOT** be reported as a network fault |
  | `UDP-006` | Visitor | A hidden tab stops the run | Start a 60 s run; switch to another tab for five seconds; come back | *The page was hidden, so the browser slowed its timers and the test stopped.* No figures presented as a finished run |
  | `UDP-007` | Visitor | One address, limited | Start seven UDP tests within a minute from one address (`REFLECTOR_SESSIONS_PER_MINUTE` 6) | The seventh: *Too many UDP tests from this address. Try again in a minute.* |
  | `UDP-008` | Administrator | An endpoint's reflector key | `POST /diagnostic/endpoint/{dnsSiteId}/reflector-key` for one of the organisation's endpoints, twice; then for another endpoint | 200 with a 43-character `key` and the `install` line, the same key both times; another endpoint, another key. **Activity** shows an *issue* against the endpoint |
  | `UDP-009` | RIVAL tester | Another organisation's endpoint | `POST …/reflector-key` and `POST …/webrtc-offer` with ACME's `dns_site_id` | 404 *That endpoint does not exist.* for both, naming nothing. 🛑 **Must NOT** reach ACME's reflector or derive its key |
  | `UDP-010` | Engineer | Keys are the endpoint manager's | As an engineer (no `dns_site_manage`), `POST …/reflector-key` | 403 |
  | `UDP-011` | Operator | Install a customer reflector | On a RHEL 8 host behind NAT, from a customer release (`UDP-042`): `sudo ./install-reflector.sh --key <key> --licence packetpulsereflector.licence --public-ip <licensed public address> --server-ip <PacketPulse server>`; `curl -s http://127.0.0.1:50000/healthz`; `sudo firewall-cmd --list-all` | The pre-flight report ending *RESULT: OK to start.*, then `"signed":true`, `"licensed":true`, `"udp_port":50000`; `/etc/default/packetpulsereflector` is `0600` root, the licence `/etc/packetpulse/packetpulsereflector.licence` `0644`; firewalld lists `50000/udp` and a rich rule for TCP 50000 from the server only. Without `--key` or `--open` the installer refuses and installs nothing; `--licence` without `--public-ip` is refused |
  | `UDP-012` | Operator | A signed reflector refuses strangers | `curl -X POST http://<reflector>:50000/api/v1/reflector/session -d '{"sdp":"v=0"}'` | 401 *This reflector accepts only signed requests.* |
  | `UDP-013` | Administrator | The relay reaches the right reflector only | `POST …/webrtc-offer` with a browser's offer for the endpoint with the reflector; then reinstall it with another endpoint's key and repeat | 200 with an `sdp` holding `a=ice-lite` and the session's `limits`; then 409 *…installed with a different key, or its clock is more than 60 seconds out…*, naming `journalctl -u packetpulsereflector` as the next step, and *Details for your engineer: … it answered HTTP 401…* |
  | `UDP-035` | Operator | The reflector is node-locked | With a release licensed for address A: `./packetpulsereflector-linux-amd64 -check -key <key> -licence packetpulsereflector.licence -public-ip A`; then the same with another address B; then the installer with `--public-ip B` | A: *authorised: yes — this host at A* and *RESULT: OK to start.*, exit 0. B: *authorised: NO — this host (B) is not one the licence names*, *RESULT: FAIL…*, exit 2; the installer stops at its pre-flight check and starts nothing. 🛑 **Must NOT** start at an address the licence does not name |
  | `UDP-036` | Operator | A release cannot be opened up | With the release binary: `packetpulsereflector -open`; then `-key <key>` with no `-licence` | Exit 2 with *this build is node-locked and cannot run with -open*; then *this build requires -licence…*. A development build (`UDP-015`) still runs `-open` |
  | `UDP-037` | Administrator | Each session must reach a licensed address | With `UDP-011` done, set the endpoint's **UDP reflector address** to a host name for the licensed IP; run; then back to the IP and run | First: 409 *The reflector on that endpoint is not licensed for the address PacketPulse reached it at…*; the reflector's journal *reflector refused a session sent to an address it is not licensed for* with `reached_at=` the name; the server log *reflector relay failed* with its hint. Then the test runs to *The reflector beside …*. 🛑 **Must NOT** serve a session sent to an unlicensed address |
  | `UDP-038` | Operator | An edited licence does not verify | Change one character inside `"licence"` in the licence file; run `-check` | *RESULT: FAIL — the licence does not verify: … the signature does not match…*, exit 2. Re-indenting the file without changing a value still verifies |
  | `UDP-039` | Operator | A clock out of step is named, not taken for a key | On the reflector host: `sudo timedatectl set-ntp false; sudo date -s '-5 min'`; run a test; then `sudo timedatectl set-ntp true` and run again | First: 409 *…installed with a different key…*; the reflector's journal *reflector refused a session request* with `reason="clock…"` and `clock_offset_s=300` (about); the server's *reflector relay failed* hint names the clock. After NTP: the test runs. 🛑 **Must NOT** log `reason="key…"` for a clock fault |
  | `UDP-040` | Operator | The licence end is enforced while running | Issue a licence whose last day is within 30 days (`VALID_UNTIL`, `GRACE_DAYS=0`); install it; read the journal | *the reflector licence ends soon* with `days_left`, at start and daily. At the end: *reflector stopped* with *the reflector licence has ended*, exit 2, and systemd leaves the service failed rather than restarting it (`RestartPreventExitStatus=2`). Proved without waiting by `TestARunningDaemonStopsWhenItsLicenceEnds` |
  | `UDP-041` | Operator | A support bundle without the key | `sudo /opt/packetpulsereflector/reflector-diagnostics.sh`; unpack the archive | The journal, the pre-flight report, firewalld, chrony, SELinux, listeners and the licence; `settings-redacted.txt` reads `PACKETPULSE_REFLECTOR_KEY=<redacted>`. 🛑 **Must NOT** contain the endpoint key anywhere (`grep -r` for it finds nothing) |
  | `UDP-043` | Engineer | A firewall drop is named, with what to hand a developer | Leave TCP 50000 to the endpoint's reflector closed at its edge firewall or cloud security list; run; note the **Reference** under the error; then find it in the server log | After about 8 s: *PacketPulse could not reach the reflector at <address>: the connection timed out, so a firewall is dropping it. Allow TCP 50000 from the PacketPulse server…*, then *Details for your engineer: at hh:mm:ss UTC … no reply came within 8s…*. The server log's *reflector relay failed* line carries the same reference (`reference=`), `reason=timed out` and `waited_ms`. 🛑 **Must NOT** say only *did not answer* |
  | `UDP-042` | Developer | A customer release is built and licensed | Fill `packetpulserelease/reflector/customers/<Customer>/reflector-ips.<Customer>.txt`; `./release.sh all`; then empty the list and run `./release.sh licence`; then point `LICENCE_SIGNING_KEY` at the key in `.env.example` and issue a licence | `packetpulserelease/<Customer>/<yyyy-mm>/packetpulse-reflector-<version>/` with one zip per CPU, `packetpulse-reflector-<Customer>-<version>-linux-amd64.zip` and `-linux-arm64.zip` (each under 20 MB, small enough to attach to a mail), and *every listed address accepted; an unlisted one refused*; unzipped, each holds only its own CPU's binary, keeps the execute bits, passes `sha256sum -c SHA256SUMS`, and its `README.txt` and guide name that zip and that binary (*This zip is for x86_64 hosts…*); every check in the guide's acceptance checklist (5.1–5.12) gives the full command to type and what it prints - `curl -s http://127.0.0.1:50000/healthz`, never `/healthz` alone and walk through firewall, key, install and a first test, with no listed address in clear. `./release.sh guide` on a build from before 18 October 2026 splits it the same way without new binaries, keeping the old zip in `vendor/`. `<Customer>/RELEASES.md` lists the build first as *current - send this*, with each zip's size and sha256, and `<Customer>/current` opens it; a second `./release.sh build` shows *built, not licensed yet* above it and leaves `current` alone until it is licensed, when the older becomes *superseded*. An empty list stops before signing; the published development key is refused. 🛑 **Must NOT** ship `vendor/` (it shows the addresses) |
  | `UDP-014` | Administrator | The relay obeys the probe policy | With `PING_ALLOW_PRIVATE=false`, `POST …/webrtc-offer` for an endpoint at `10.x`; then as the demo organisation for one at `192.168.x` with private addresses allowed | 400 *That endpoint's address is not one this server may send to.*, both times. 🛑 **Must NOT** send anything to the private address |
  | `UDP-016` | Engineer | UDP is the browser's only method | In a browser open **Run diagnostic**; fill the ticket; then open it in the desktop app | No **Measurement method** row; **Test length** *10 s · 30 s · 60 s* with 30 chosen and **Packets per second** *10 · 25 · 50 · 100* with 50 chosen; no Samples, Interval or Timeout. The desktop app has no UDP choice and keeps **ICMP echo**. 🛑 **Must NOT** offer *HTTPS to the endpoint* |
  | `UDP-017` | Engineer | A test to the PacketPulse network is the access line's | Endpoint with no reflector address; **Start test** (10 s); **Attach to a ticket**; open the ticket and its PDF | *Measured to: The PacketPulse network: the customer's access line…*, no *Pune Edge · …* line; the ticket's row is *PacketPulse network* at PacketPulse's address, with *Where packets went*; the PDF prints the same line. 🛑 **Must NOT** name or file it under the endpoint |
  | `UDP-018` | Engineer | A test to the endpoint's reflector | Endpoint with a reflector (`SITE-011`, `UDP-011`); run; attach | *Measured to: The reflector beside Pune Edge*; filed under the endpoint. The reflector's `/healthz` shows an active session during the run |
  | `UDP-019` | Engineer | Loss placed on the ticket | As `UDP-002`, through **Run diagnostic**, dropping 2% outbound | The ticket's *Upstream lost* is about 2% and *Downstream lost* about 0%; the figures match the screen's. 🛑 **Must NOT** split one direction's loss across both |
  | `UDP-020` | Engineer | The endpoint's reflector is down | Stop the endpoint's reflector; run | *PacketPulse reached <host>, but nothing answers on TCP 50000. Check that the reflector service is running on that host…* and *Details for your engineer: … the connection was refused…* with **Try again**, and **Test to the PacketPulse network instead**; taking it runs and files as `UDP-017`. Another endpoint chosen afterwards tries its own reflector first |
  | `UDP-026` | Engineer | An endpoint naming PacketPulse's own reflector | Edit an endpoint: **UDP reflector address** PacketPulse's address (`REFLECTOR_PUBLIC_IP`, live `140.245.28.232`), then `140.245.28.232:50000`; run each time; attach | The test runs at once, with no *did not answer*; *Measured to: The PacketPulse network…* and filed as `UDP-017`. 🛑 **Must NOT** fail with *The reflector on that endpoint did not answer.*, nor file PacketPulse's reflector under the endpoint |
  | `UDP-044` | Engineer | PacketPulse's test reflector answers every endpoint | With PacketPulse's standalone reflector installed with the key from `packetpulseserver reflector-network-key` and its address in `REFLECTOR_NETWORK_HOSTS` (live `140.245.28.232`): set it as the **UDP reflector address** of two endpoints, one in each of two organisations (on live, *Chennai Core* and *Lab host (TEST-NET-1) Bengaluru*); run the UDP test for each | Both run: *UDP over WebRTC → The reflector beside …* for each endpoint, filed under it. Then an endpoint naming a customer's reflector still runs only with that endpoint's own key (`UDP-013`). 🛑 **Must NOT** answer 409 *…installed with a different key…* for any endpoint naming PacketPulse's reflector, nor let a customer's reflector answer a request signed with the network key |
  | `UDP-045` | Engineer | Each IP family is tested on its own | Run a UDP test of an endpoint whose reflector has both an IPv4 and an IPv6 address, from a network with both; attach | *The same test over IPv6.* after the first run; **By IP version** shows IPv4 and IPv6 figures; two observations filed, each with its `reflector_ip` (one `x.x.x.x`, one `…:…`). 🛑 **Must NOT** let the browser pick one family and report it as both |
  | `UDP-046` | Engineer | A family that could not be tested says why | Run the same test from a network without IPv6; then to a reflector with no IPv6 address; then with IPv6 UDP blocked at a firewall | **By IP version**: *Not tested: this device has no IPv6 connection.*, then *Not tested: the reflector has no IPv6 address.*, then *Not tested: UDP over IPv6 did not get through.*; the IPv4 run is filed alone each time. 🛑 **Must NOT** call a missing IPv6 path *UDP blocked*, or fail the IPv4 run because of it |
  | `UDP-027` | Engineer | The route says when the endpoint is not what is tested | Choose an endpoint with no reflector address; read the route; **Start test** (10 s) and watch it; then repeat with an endpoint whose reflector is installed (`UDP-018`), and with one whose reflector address is PacketPulse's own (`UDP-026`) | No reflector: before, during and after, *The PacketPulse network*, **Not the endpoint** in amber, and *…has no UDP reflector…*; the log says *Answered by the PacketPulse network, not …*. Its own reflector: *The reflector beside Pune Edge*, **Path to the endpoint**. PacketPulse's address: *The reflector beside …* before Start, then *The PacketPulse network* and **Not the endpoint** as soon as the test begins. 🛑 **Must NOT** show **Path to the endpoint** for a test the PacketPulse network answered |
  | `UDP-028` | Engineer | A finished run reaches 100% and says what happened | Run 30 s while loading the device (open a heavy page in another window, or throttle the CPU in the browser's developer tools); let it finish; then run again and **Stop** after ten seconds | The first run: **Progress** *100%* with *n sent · m skipped*, and **What happened** naming the m skipped packets as the device being busy, not loss; *Where packets went* shows *Not sent m (device busy)*; the loss figure is out of the n sent. The stopped run: below 100%, and *The test was stopped part-way…*. 🛑 **Must NOT** stop at *99%* on a run that finished, nor count a skipped packet as lost |
  | `UDP-021` | Engineer | UDP blocked | As `UDP-004`, through **Run diagnostic** | After about 16 s, and *No UDP came back on that connection; trying once more on a fresh one.* in the log at about 8 s: *UDP is blocked on this network: … tried on two fresh connections, but no UDP packet got through. Calls and games on this connection will fail the same way.* with **Try again**, which runs the UDP test again. No figures, nothing to attach. The server log has two *reflector session ended* lines for the person with `connected=false checks_received=0`. 🛑 **Must NOT** offer the HTTPS test, nor a 100% loss run for filing |
  | `UDP-032` | Engineer | A first connection that gets nothing through is retried | Block outbound UDP to port 50000; **Start test**; lift the block within 6 seconds | The log says *No UDP came back on that connection; trying once more on a fresh one.*, the run measures, and *The first connection got no UDP through and the second did…* is logged with its figures. No *UDP is blocked* banner. 🛑 **Must NOT** call UDP blocked on one failed connection |
  | `UDP-033` | Engineer | A browser that sends no UDP is told so | In Chrome, set **WebRTC IP handling policy** to *disable_non_proxied_udp* (an extension or enterprise policy), or in Firefox set `media.peerconnection.ice.relay_only` with no relay; **Start test** | *This browser sent no UDP, so the network was not tested: a browser setting, an extension or a VPN is keeping WebRTC off UDP…* with **Try again**, after one attempt. 🛑 **Must NOT** say the network blocks UDP |
  | `UDP-034` | Engineer | The ticket says how long the test took | Run a 30 s UDP test; **Attach to a ticket**; open it in **History** | The heading's subtitle reads *… 1 endpoint(s) in 3x.x s*, the run's own duration. A run filed by an older client reads as before. 🛑 **Must NOT** read *in 0 ms* |
  | `UDP-029` | Engineer | The rate is a choice | Choose **Packets per second** *100* and **Test length** *10 s*; **Start test**; then *10* and run again | The log's *Pacing: 1000 packets, 100 a second*, **What happened** *…room for 1000 packets at 100 a second*, the note *…100 packets a second…*; then *100 packets, 10 a second*. Changing the rate while a test runs is not possible. 🛑 **Must NOT** clamp 100 a second on PacketPulse's network (it allows 120 a second and 6000 packets) |
  | `UDP-030` | Engineer | A slow handshake is not called blocked | Drop 40% of inbound UDP from port 50000 for the first 15 seconds of a run (`dnctl pipe 1 config plr 0.4`, then `plr 0`); **Start test** | Either the test starts late and runs, or *UDP got through, but the connection over it could not be completed…* with **Try again**. 🛑 **Must NOT** say *UDP is blocked* while ICE got through |
  | `UDP-031` | Engineer | A filed UDP test is graded | Drop 5% of inbound UDP for a 30 s run; **Attach to a ticket**; open it in **History**; then repeat on a clean line | The lossy run: status *Breached*, **SLA breaches** *1*, **Service level** *Breached 1 100%*. The clean run: *OK*, **SLA breaches** *0*, *Within SLA*. 🛑 **Must NOT** read *UNKNOWN*, or draw a breach the tile does not count |
  | `UDP-022` | Engineer | A lapsed licence | As a licence-expired organisation's engineer, start a UDP test | The licence message with who can renew it. 🛑 **Must NOT** offer the PacketPulse network as a way round it |
  | `UDP-023` | Engineer | Stopped part-way | Start a 60 s test; **Stop** after ten seconds | The figures for what was sent, loss as on the line (no packets in flight counted lost); it can be attached |
  | `UDP-024` | Engineer | A hidden tab | Start a test; switch tabs for five seconds | *The test stopped because this tab was hidden…* with **Try again** |
  | `UDP-025` | RIVAL tester | A split the test could not have measured | `POST /diagnostic/clientobservation` with an `https-reach` observation carrying `packets_lost_up`; then a `udp-webrtc` one with `reflector: "endpoint"` for an endpoint with no reflector address; then one whose up + down + late exceed what was lost | 422 on `observations` each time; nothing filed |
  | `UDP-015` | Developer | The development loop | Set `REFLECTOR_UDP_PORT=50000` in the development `.env`; run `go run -tags packetpulsedev ./cmd/packetpulseserver` and the app in a browser against it (`flutter run -d chrome --dart-define=PACKETPULSE_API_BASE_URL=http://localhost:8080`); in **Run diagnostic** test an endpoint with no **UDP reflector address** | The test runs to the development server's own reflector: *UDP over WebRTC → The PacketPulse network*. (Until 17 October 2026 this loop used the self-test with `?reflector=`; the engine still honours the override, but no page offers a UDP test to use it) |
- ⚙️ **Developer Guide & Release Confidence**:
  - The reflector: `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorEngine.go` (pion, ICE-lite, passive DTLS, multicast DNS off, one multiplexed UDP port), `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorSession.go` (the caps and the per-sequence bitmap) and `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorProtocol.go` (the frames). An echo is exactly as long as its probe, so the reflector cannot amplify. Offers declaring `a=ice-lite` are refused: with both ends lite, pion selects a pair without any check and sends to it, and `TestALiteOfferNeverMakesTheReflectorSendToACallerChosenAddress` proves a caller-chosen address hears nothing.
  - Real loss: `TestRealLossIsSplitExactlyByDirection` in `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorLoss_test.go` drops real datagrams through a UDP proxy and checks the bitmap and the echoes against the drops; it also fails if the probe channel ever retransmits.
  - Signalling: `packetpulsego/pkg/reflectormicroservice/reflectorapp/ReflectorRouteHandler.go` (per-address limit, signatures, replay refusal, the signed answer) with `packetpulsego/pkg/reflectormicroservice/reflectorsign/ReflectorSign.go` (keys derived from `REFLECTOR_KEY_SECRET`, never stored). The hosted reflector: `packetpulsego/cmd/packetpulseserver/PacketPulseReflector.go`. The customer daemon: `packetpulsego/cmd/packetpulsereflector/main.go`, `scripts/reflector/install-reflector.sh`, `scripts/reflector/packetpulsereflector.service`.
  - The node-lock: `packetpulsego/pkg/reflectormicroservice/reflectorlicence/ReflectorLicence.go` (the signed licence, Ed25519, its format kept apart from a backend licence's) and `packetpulsego/pkg/reflectormicroservice/reflectorlicence/ReflectorSeal.go` (each address sealed with X25519, HKDF-SHA256 and AES-256-GCM; the reflector's key compiled into the release). In the daemon, `authoriseStart`, `requireLicensedAddress`, `runCheck` and the licence deadline in `run`; the refusal reasons in `packetpulsego/pkg/reflectormicroservice/reflectorapp/ReflectorRouteHandler.go`; the relay's failure log and the 403 mapping in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticReflectorRelay.go`. Issuing: `packetpulsego/cmd/packetpulsereflectorlicence/main.go` (`--init-keys`; refuses the key published in `.env.example` and any key the build does not trust). The release: `packetpulserelease/reflector/BuildReflectorRelease.sh` (garble with a recorded seed, so a panic from the field can be reversed), `scripts/reflector/uninstall-reflector.sh`, `scripts/reflector/reflector-diagnostics.sh`. Each guard was proved by breaking it: `TestAuthoriseStartAcceptsOnlyTheLicensedHost`, `TestEditingASignedLicenceIsRejected`, `TestALicensedBuildIsNodeLocked`, `TestEachSessionMustBeSentToALicensedAddress`, `TestARefusedRequestIsLoggedWithItsCause`, `TestEachWayAReflectorCanFailIsNamed`.
  - The relay: `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticReflectorRelay.go`, through the sweep's probe policy and the demo's public-only confinement; tenancy, policy and every reflector failure in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticReflectorRelay_test.go`. An endpoint whose reflector address is PacketPulse's own (`REFLECTOR_PUBLIC_IP`, or a name resolving to it, any port) is answered by the hosted reflector in-process and named the PacketPulse network: the relay would otherwise post to TCP 50000 on this host, where nothing listens (`namesHostedReflector`, `TestAnEndpointNamingPacketPulsesReflectorIsTestedToIt`). An endpoint whose reflector address is one of `REFLECTOR_NETWORK_HOSTS` is relayed to that standalone reflector as usual, but signed with `DeriveNetworkKey` (`packetpulsego/pkg/reflectormicroservice/reflectorsign/ReflectorSign.go`) rather than its endpoint key, so PacketPulse's test reflector answers every endpoint while a customer's answers only its own (`TestPacketPulsesReflectorAnswersForEveryEndpoint`). The request goes to the IP that matched, never to a second lookup of the name, so a name that changes in between cannot carry the network key's signature elsewhere (`TestANetworkKeyRequestGoesOnlyToTheAddressThatMatched`); and the key icon still gives such an endpoint only its own key, never the network key (`TestTheKeyRouteNeverRevealsTheNetworkKey`). The key is printed on the server by `packetpulseserver reflector-network-key` (`packetpulsego/cmd/packetpulseserver/ReflectorNetworkKey.go`), so it is derived only where the secret is.
  - In the app: the engine is the self-test's own, copied to `packetpulseflutter/web/engine/` (`packetpulseflutter/test/udp_probe_run_test.dart` fails if either copy differs from `packetpulseweb/assets/`). `packetpulseflutter/lib/diagnosticmicroservice/service/udp/UdpProbeRun.dart` maps its result to samples and the split, checked against `packetpulsetest/contracts/udp_run_fixture.json`, which the engine's own test regenerates; `UdpProbeRunnerWeb.dart` is the browser bridge, `UdpProbeRunnerNative.dart` says no. The screen: `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart`; the split line, worded as the PDF's, `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/UdpSplitLine.dart`.
  - Filing: `udpProblem` and `namePacketPulseNetwork` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticUdpObservation.go` refuse a split on any other method, an unknown reflector, negative counts and a split larger than the loss, and name a PacketPulse-network result from the server's own `REFLECTOR_PUBLIC_IP`. The split's columns and their CHECK are in `packetpulsego/pkg/common/dbclient/migrations/0031_2026_10_09_udp_split_and_endpoint_reflector.sql`; the PDF line is `SplitLine` in `packetpulsego/pkg/pingmicroservice/pingdomain/shared/PingResult.go`. The CSV export's columns are unchanged. Signed in, the hosted reflector counts sessions per person (`user:<id>`), not per address.
  - Coverage, in-app: `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticUdpObservation_test.go`, the `contract` suite's `udp_observation_test.go`, `packetpulseflutter/test/run_diagnostic_udp_test.dart`, `packetpulseflutter/test/udp_probe_run_test.dart`.
  - Connecting: `connect` in `packetpulseweb/assets/packetpulse-udp.js` opens one connection; `runUdpTest` tries `CONNECT_ATTEMPTS` (2) when nothing came back, and `iceDiagnosis` reads the browser's own candidate-pair counts to tell *no UDP sent* (`no-udp`) and *answered* (`handshake`) from *nothing came back* (`udp-blocked`). The server's half is `logOutcome` and `iceChecks` in `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorEngine.go`, held by `TestASessionThatNeverConnectedSaysNoCheckArrived` and `TestASessionThatConnectedIsLoggedWithWhatItReceived`.
  - The page's half: `packetpulseweb/assets/packetpulse-udp.js` (pacing that skips a missed slot rather than bursting, the per-sequence split, per-direction RFC 3550 jitter from the reflector's stamps, and the connect wait: `CONNECT_WAIT_MS` for ICE, `HANDSHAKE_WAIT_MS` once ICE has got through, kept under the hosted reflector's 15 s idle limit), tested on a virtual clock in `packetpulseweb/assets/packetpulse-udp.test.mjs` (part of `unit`).
  - The rate: `_udpRates` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart`, passed to the engine as `ratePps` by `UdpProbeRunnerWeb.dart`. `HostedLimits` in `packetpulsego/pkg/reflectormicroservice/reflectorcore/ReflectorSession.go` allow the fastest for the longest test with a fifth of headroom (`TestThePagesLongestRunFitsTheHostedCaps`).
  - Live: `REFLECTOR_UDP_PORT=50000` and `REFLECTOR_PUBLIC_IP` on the server; UDP 50000 open in the Oracle security list and the host firewall; `AF_NETLINK` in `scripts/deploy/packetpulse.service` (the reflector lists the host's interfaces through it). nginx needs nothing: the session route is under `/api/`, and the packets go straight to the UDP port. Since 16 October 2026 the built-in reflector is off (`REFLECTOR_UDP_PORT=0`, `REFLECTOR_PUBLIC_IP` unset) and the licensed standalone `packetpulsereflector` answers on 50000; since 17 October it holds the network key and its address is in `REFLECTOR_NETWORK_HOSTS`, so every endpoint naming it is answered (`UDP-044`).

---
## Group 5 — Service targets and grading

How every result is judged: against the targets an organisation set, the same way for a sweep and for a call measured from the customer's own device.

---

### 5.1 🎯 Service targets, and the grade each result gets

**Screen:** none — the History result view (§3.2) shows the grades · **Routes:** through the API only: `GET /monitor/sla/list` (`sla_view`), `POST /monitor/sla/add`, `PUT /monitor/sla/{slaPolicyId}`, `DELETE /monitor/sla/{slaPolicyId}` (`sla_manage`) · **Capabilities:** `sla_view`, `sla_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Every result is graded against targets for latency, jitter, loss and MOS: **OK / Degraded / Breached**, with Degraded warning at 80% of a threshold, before the customer notices. A call measured from the customer's own browser is graded by the same targets as a sweep, so a ticket says *Breached* the moment the line drops a call's worth of packets.
- 📖 **User Guide & Operational Flow**:
  - Targets are set through the API: a default for the organisation, and any number more, each site naming one (§3.1). A result with no site — a UDP test to the PacketPulse network — is held to the organisation's default; with no targets set, to the built-in defaults (100 ms, 30 ms jitter, 1% loss, MOS 3.6).
  - What is graded: a sweep's results, and a device's **UDP test** or **ICMP echo**, which put real packets on the wire. An **HTTPS** round-trip from a device is not graded — the targets are set for packets, and TCP turns loss into delay — and reads *Not graded* (stored as `unknown`), with the reason under the results (§3.2). Until the review of 13 October 2026 every device result read *Unknown*, the UDP test included.
  - A graded result keeps each figure against its target - the metric, what was measured, the target, where Degraded begins (80% of a ceiling, 105% of the MOS floor) and that figure's verdict - in `ping_result.sla_checks` (migration 0033; 0032 kept breaches alone, the same day). History, **View ticket**, the attach card on **Run diagnostic** and the PDF show them (§3.2, `DIAG-019`). Grading always knew and kept only the grade: a ticket read *Breached* beside 0% loss and 2.9 ms jitter, for a 117.6 ms round trip, and *Degraded* with nothing behind it; its PDF showed only *UP*.
  - A ticket counts its breaches (*SLA breaches* in History); a dual-stack site's IPv6 result is reported but never counted. The **Service level** chart draws each grade as itself: *Within SLA*, *Degraded*, *Breached*, and *Not graded* for what was not graded — and *Excluded* for a result filed under planned work before October 2026. Until 13 October 2026 it drew anything ungraded as *Breached*, so every device test read *Breached 100%* beside a tile that said no breach.
  - **Automatic monitoring was removed** after the review of 13 October 2026: scheduled sweeps, alert channels and alerts, planned work (maintenance windows), the month-end SLA report and its PDF, and the dashboard's 30-day availability trend, with the **Monitoring** screen and the `schedule_manage` capability. Their routes answer 404. What they stored stays in its tables, unread; nothing runs on a timer but the result store's housekeeping. Policies no longer take `breach_threshold_runs` or `alert_cooldown_minutes`: a request naming either is refused (400) for naming a field the request does not accept.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `MON-001` | Administrator | Grading against a target | Through the API, a target: latency 100 ms; a site averaging 85 ms; another at 120 ms | 85 ms → *Degraded* (past 80% of 100); 120 ms → *Breached*; one under 80 ms → *OK* |
  | `MON-002` | Administrator | The default target applies to new sites | Through the API, mark a target default; add a site with no target; run | The new site is graded against the default |
  | `MON-010` | Tester | The guard suite | `./packetpulsetest.sh grading` | Passes — a device's UDP test graded and its breach counted, an HTTPS round-trip never graded, and automatic monitoring's routes gone, against a live server |
  | `MON-011` | Administrator | Monitoring is gone | Open the app as the owner; then `GET /monitor/schedule/list`, `GET /monitor/slareport`, `GET /monitor/trend` | No **Monitoring** in the navigation and no availability card on the **Dashboard**; each route answers 404. 🛑 **Must NOT** offer *Monitoring*, *Schedules*, *Alert channels*, *Alert history*, *Planned work* or *SLA report* anywhere |
  | `MON-013` | Engineer | A UDP test is graded against the organisation's own targets | Through the API, a lenient default target (400 ms, 10% loss, MOS 2); file a UDP test to the PacketPulse network losing 5%; then remove the target and file the same | *OK* against the lenient default; then *Breached* against the built-in defaults, with *SLA breaches* 1 |
  | `MON-014` | Engineer | An HTTPS round-trip is never graded | `POST /diagnostic/clientobservation` with an `https-dns` observation answering 2 of 10 at 950 ms | `sla_status` *unknown*, `breach_count` 0; History's chart *Not graded 1 100%*. 🛑 **Must NOT** draw it as a breach |
  | `MON-015` | Administrator | Damping is no longer taken | `POST /monitor/sla/add` with `breach_threshold_runs: 2` | 400 naming `breach_threshold_runs`; the same body without it saves |
- ⚙️ **Developer Guide & Release Confidence**:
  - Grading: `EvaluateSla` in `packetpulsego/pkg/pingmicroservice/pingservice/PingSlaEvaluator.go`; device results in `PingRunRecord` (`packetpulsego/pkg/pingmicroservice/pingservice/PingService.go`), graded when `ObservesPackets` says the method put packets on the wire; the ticket's count is `breachCount` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go`. Targets: `SlaTargetsForSite` in `packetpulsego/pkg/monitormicroservice/monitorservice/MonitorService.go` (an empty site is the organisation's default).
  - The chart: `DiagnosticOutcomeDonut` in `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticCharts.dart`, held by `packetpulseflutter/test/diagnostic_charts_test.dart`.
  - Results are partitioned by month (`packetpulsego/pkg/common/dbclient/migrations/0012_2026_10_01_result_partitioning_and_retention.sql`); the retention runner (`packetpulsego/pkg/monitormicroservice/monitorservice/MonitorRetentionRunner.go`) keeps partitions three months ahead and drops what no organisation's retention still wants. It no longer writes daily rollups: only the removed report and trend read them.
  - Retired with automatic monitoring: Monitoring cases 003 to 009 and 012 (schedules, damping, escalation, webhooks, planned work, the SLA report and its PDF, and the screen's capability), and the fifth journey (an outage and its recovery).
  - Coverage: `packetpulsetest/golang/gradingconformance/`, `packetpulsego/pkg/monitormicroservice/**`, `packetpulsego/pkg/pingmicroservice/pingservice/PingRunRecord_test.go`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticUdpObservation_test.go`, `packetpulseflutter/test/shell_test.dart`, `packetpulseflutter/test/dashboard_screen_test.dart`.

---
## Group 6 — Integrations

How results reach the operator's own systems, and how its directory signs people in.

---

### 6.1 🔌 Results API keys

**Screen:** Configure → **Results API keys** · **Routes:** `GET /apikey/list`, `POST /apikey/add`, `DELETE /apikey/{credentialId}`; with a key: `GET /result/bytt/{ttNumber}`, `GET /result/export.csv` · **Capability:** `apikey_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: The ticketing system pulls the evidence itself — one ticket as JSON, or a month of results as CSV — with a key scoped to the operator's own organisation, readable even if a renewal is late. The key is shown once and stored only as a hash.
- 📖 **User Guide & Operational Flow**: **Issue key** with a label and an optional expiry; copy it — it is never shown again. Your system sends it as `Authorization: Bearer <key>`. `GET /api/v1/result/export.csv?from=…&to=…` takes RFC 3339 times, at most 31 days apart; with no `from`, the day before `to` (default now).
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `API-001` | Administrator | A key is shown once | Issue a key; reload the screen | The full key appeared once; afterwards only its prefix, label, expiry and last use |
  | `API-002` | Integrator | One ticket as JSON | `curl -H "Authorization: Bearer <key>" https://…/api/v1/result/bytt/TT-MAN-001` | The ticket with every result, including `ip_version` and `report_only` |
  | `API-003` | Integrator | A period as CSV | `…/result/export.csv?from=2026-10-01T00:00:00Z&to=2026-10-02T00:00:00Z` | `text/csv`, the contracted header, one row per result of the diagnostics that finished in that window |
  | `API-004` | Integrator | Bad periods are refused, not guessed | `from` 40 days before `to`; `to=tomorrow`; `from` after `to` | 422 (*at most 31 days*), 400 (unreadable time), 422. 🛑 **Must NOT** answer with a partial file |
  | `API-005` | Integrator | The wrong credential is refused | No header; a person's session token; a revoked key | 401 each |
  | `API-006` | RIVAL integrator | A key reads only its own organisation | RIVAL's key for ACME's TT, and a period covering ACME's tickets | 404; a CSV with none of ACME's rows |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/apikeyauth/`; the organisation comes from the key, never from the request.
  - The bulk pull **streams**; a failure after rows have gone out aborts the response rather than ending a short file a system would take as whole.
  - Coverage: `packetpulsetest/golang/apicontract/results_api_test.go`, `packetpulsetest/golang/apicontract/result_export_test.go`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticapp/`.

---

### 6.2 📤 Result export

**Screen:** Configure → **Result export** · **Routes:** `GET/PUT /export/target`, `POST /export/target/test`, `GET/PUT /report/settings` · **Capability:** `apikey_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Results delivered to the operator's own SFTP or FTPS server as CSV, every interval it chooses — no integration project. Nothing is lost: a failed delivery is retried with the same period, and each row carries a `result_id` so a file that arrives twice is recognised. The server's key is confirmed before anything is sent, and a saved password is only ever sent to the server it was entered for.
- 📖 **User Guide & Operational Flow**:
  - **Protocol:** SFTP (recommended), FTP over TLS, or FTP — with a clear warning that plain FTP is unencrypted.
  - **Server, Port, Username, Password** (or for SFTP a private key). Saved credentials are never shown again; leave the field empty to keep them, or tick *Remove the saved password*.
  - **Test connection** signs in, writes and removes a small file. For SFTP it shows the server's key: check it with whoever runs the server, then **Trust this key** and **Save**.
  - **Folder** and **Deliver every (minutes)** — 15 to 1440.
  - The status card says how the last delivery went: *Sent <file> with N results*, *Nothing new to send*, or *Not delivered: <why>*.
  - **Scheduled reports**, below, sends the network reports (§3.6) to the same server: the **Time zone** every report is cut in and every time in a result file is written in — the organisation's country's until one is chosen — and *Send each day's / week's / month's / year's reports*, each with how it last went and the period it sends next. It saves on its own **Save**; the result export need not be switched on.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `EXP-001` | Administrator | A target nobody set up | Open the screen | *Not delivered yet. The first delivery is one interval after saving.*; SFTP, port 22, 60 minutes; switched off |
  | `EXP-002` | Administrator | Confirm the server's key | Fill in an SFTP server; **Test connection** | *Trust this server?* showing `SHA256:…`. **Cancel** → *Not trusted yet*; test again → **Trust this key** → *Trusted: SHA256:…*; **Save** |
  | `EXP-003` | Administrator | Switching on needs everything a delivery needs | Switch on with no server, no username, no password and no trusted key | Refused against each field |
  | `EXP-004` | Administrator | A changed key is a warning, not a detail | Re-key the SFTP server (or point the name at another); **Test connection** | *The server's key has changed* with the new key to check; the trusted key stays until you trust the new one. 🛑 **Must NOT** deliver to a server showing a different key |
  | `EXP-005` | Administrator | A saved password goes nowhere new | With a saved password, change the server (or port, username or protocol) and test or save | The password field becomes required again (*Enter the password again…*). 🛑 **Must NOT** sign in to the new server with the saved password |
  | `EXP-006` | Administrator | A delivery lands whole | Switched on, every 15 minutes, a diagnostic run after saving; wait for the interval | `packetpulse-results-<from>-<to>.csv` in the folder (written as `.part`, renamed when whole), named in UTC, its `checked_on` in the organisation's zone (`…+05:30` in India); the card says *Sent … with N results* |
  | `EXP-007` | Administrator | A quiet period sends nothing | No diagnostics in an interval | No file; *Nothing new to send* |
  | `EXP-008` | Administrator | A failure is retried, nothing lost | Change the server's password; wait for a delivery; restore it; wait again | First: *Not delivered: The server refused the sign-in…*; the next delivery covers the same results. 🛑 **Must NOT** skip the failed period |
  | `EXP-009` | Administrator | PacketPulse's own host is not a destination | Server `127.0.0.1`, `localhost`, `169.254.169.254` or `::1`; **Test connection** | Refused against **Server**: *PacketPulse may not connect to that address.* |
  | `EXP-010` | Administrator | Plain FTP is warned about | Choose FTP | The unencrypted warning; port moves to 21; the private key and server key disappear |
  | `EXP-011` | Administrator | FTPS with a private authority | FTPS to a server whose certificate your own CA signed, without and then with that CA | Without: *certificate is not trusted*; with it pasted: test passes |
  | `EXP-012` | Tester | Credentials never come back | `GET /api/v1/export/target` | `has_password`, `has_private_key` — never the password, key or sealed text |
  | `EXP-013` | Administrator | Scheduled reports need a server | With no server saved, switch on *Send each day's reports*; **Save** in *Scheduled reports* | Refused: *Set up the file server on this page, test it and trust its key before switching scheduled reports on.* |
  | `EXP-014` | Administrator | Each closed day is four files | A trusted SFTP server; time zone `Asia/Kolkata`; daily on; wait past 00:30 IST | `packetpulse-city-daily-<day>.csv/.pdf` and `packetpulse-pan-india-daily-<day>.csv/.pdf` in the folder (a year: `packetpulse-city-yearly-<year>.csv`, sent after midnight on 1 January); the schedule reads *Last sent: …* and *Next to send: the period from <the next day>.* Switched on, the last closed day is sent at once |
  | `EXP-015` | Administrator | A quiet day is still sent | A day with no tests | All four files, the PDF saying *No test…*. 🛑 **Must NOT** skip a day: a missing file means a failure |
  | `EXP-016` | Administrator | A failed sending is retried, nothing skipped | Change the server's password; let a period close; restore it | *Not sent: The server refused the sign-in…*; within fifteen minutes the same period is sent. 🛑 **Must NOT** move on past the failed period |
  | `EXP-017` | Administrator | The zone is kept as chosen | Choose `Asia/Dubai`; **Save**; reload | `Asia/Dubai`; every report now cut in it (§3.6) |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/exportmicroservice/` (target, test, runner each minute); transport: `packetpulsego/pkg/common/filedrop/FileDrop.go` (SFTP via `pkg/sftp`, FTP/FTPS via `jlaffaye/ftp`).
  - 🔒 Every connection — the FTP **data** connection included — is made by `probeguard.Dialer`, whose `Control` hook checks the resolved address before connecting. A dial *function* would have made the FTP library send FTPS data unencrypted.
  - The watermark (`exported_through`) moves only on a delivery, and only from where the run found it; a run catches up a day per file; the last two minutes are left to settle.
  - `EXPORT_ALLOW_LOOPBACK` permits a server on PacketPulse's own host for development; production refuses to boot with it.
  - Scheduled reports: `report_settings` and `report_delivery` (migration 0035), run each minute by `packetpulsego/pkg/reportmicroservice/reportservice/ReportService.go`, which calls the export's `Drop` — the same account, key pin and dialer. A period is sent 30 minutes after it closes; `next_period` moves only when all four files land, and only from where the run found it.
  - Coverage: `packetpulsego/pkg/common/filedrop/` (in-process SFTP and FTP/FTPS servers), `packetpulsego/pkg/exportmicroservice/**`, `packetpulseflutter/test/result_export_screen_test.dart`, `packetpulsetest/golang/apicontract/result_export_test.go`.

---

### 6.3 🏢 Directory

**Screen:** Configure → **Directory** · **Routes:** `GET/PUT /ldap/config`, `POST /ldap/config/test`, `/ldap/groupmap/*` · **Capability:** `ldap_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Staff sign in with their existing network password over LDAP or Active Directory, and their directory group decides their PacketPulse role — after the second step, so a directory password alone still opens nothing. The owner always keeps a local password, so a directory outage cannot lock out the person who fixes it.
- 📖 **User Guide & Operational Flow**: Host and port (389 LDAP/StartTLS, 636 LDAPS), encryption, the CA certificate of your own authority, a read-only bind account, the base DN; then **Test connection** before enabling. **Group mappings** give each directory group a role.
  - **Who is looked up:** the person's PacketPulse email address up to the `@` is what replaces each `%s` in the **User filter**. Where directory account names match email addresses, `(sAMAccountName=%s)` (Active Directory) or `(uid=%s)` (OpenLDAP, FreeIPA) works; where they do not (staff numbers), search by mail: `(mail=%s@corp.example.com)`.
  - **Group filter:** each `%s` is the person's DN; for groups that list members by user name, `%u` is their user name — `(&(objectClass=posixGroup)(memberUid=%u))`. One of the two is required — without it every person would be in every group. Active Directory's `(&(objectClass=group)(member:1.2.840.113556.1.4.1941:=%s))` also follows nested groups. A filter may use a placeholder more than once: `(|(uid=%s)(mail=%s@corp.example.com))`.
  - Settings for Active Directory, OpenLDAP, OpenLDAP with `groupOfUniqueNames` or `posixGroup`, and FreeIPA are each proven end to end, using only the screen's fields, by `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapDirectoryProducts_test.go`.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `LDAP-001` | Administrator | Connect and test | Point at a test directory with its CA certificate; **Test connection** | Success; a wrong CA → certificate error; skipping verification is warned against |
  | `LDAP-002` | Directory user | The group decides the role — after the second step | Map group `noc` → *NOC Engineer*; a directory user in `noc` signs in | Password checked by the directory; second step asked; only then the role applied |
  | `LDAP-003` | Owner | The owner keeps a local password | Directory switched on and unreachable; owner signs in with their local password | Signed in. Others: *Your organisation's directory server could not be reached. Please retry.* |
  | `LDAP-004` | Tester | An outage is not a failed sign-in | Directory down; an administrator signs in | No *Sign-in failed* entry in Activity — nobody got the password wrong |
  | `LDAP-005` | Administrator | A group filter must name the person | Group filter `(objectClass=group)`; **Save** | Refused against **Group filter**: *The group filter needs a %s where the person's DN goes, or a %u where their user name goes.* 🛑 **Must NOT** save a filter that puts everyone in every group |
  | `LDAP-006` | Directory user | Nested Active Directory groups grant the role | Map `PacketPulse NOC` → *NOC Engineer*; the person is in `NOC Shift A`, itself a member of `PacketPulse NOC`; group filter `(&(objectClass=group)(member:1.2.840.113556.1.4.1941:=%s))`; sign in | *NOC Engineer* after the second step |
  | `LDAP-007` | Directory user | Account names that are not email addresses | AD account `E10442`, PacketPulse email `asha.rao@corp.example.com`; user filter `(&(objectCategory=person)(objectClass=user)(mail=%s@corp.example.com))` | Signed in with the directory password. With `(sAMAccountName=%s)` instead: *incorrect email address or password* |
  | `LDAP-008` | Administrator | A silent firewall does not hang the screen | Point **Host/Port** at an address that accepts the connection and never answers (a firewall that drops after the handshake, or `nc -lk 3890` on a test host); StartTLS; **Test connection** | Within about the connection timeout (10 s): *Could not reach a directory server at that host and port. Check the hostname, the port, and the encryption mode.* 🛑 **Must NOT** spin indefinitely, and **must NOT** say *The directory refused the service account* |
  | `LDAP-009` | Directory user | A person in hundreds of groups keeps them all | In a large AD, a person in several hundred groups, one of them mapped; sign in | The mapped role. 🛑 **Must NOT** fall back to the default role because the person has many groups |
  | `LDAP-010` | Tester | A filter that matches several people is named, not guessed | User filter `(uid=a*%s)`; a person signs in | The person is told *Your organisation's directory server could not be reached. Please retry.*; the server log reads *the user filter matched more than one entry*. 🛑 **Must NOT** sign in as whichever entry came first |
  | `LDAP-011` | Directory user | Groups that list members by user name | OpenLDAP `posixGroup` `netops` with `memberUid: asha.rao`, mapped to *NOC Engineer*; group filter `(&(objectClass=posixGroup)(memberUid=%u))`; `asha.rao` signs in | *NOC Engineer*. With `%s` instead of `%u`: no group, the role unchanged |
  | `LDAP-012` | Directory user | A filter that names the person twice | User filter `(&(objectClass=person)(\|(uid=%s)(mail=%s@corp.example.com)))`; a person whose `uid` is a staff number but whose mail is their PacketPulse address signs in | Signed in. 🛑 **Must NOT** fail because only the first `%s` was filled |
  | `LDAP-013` | Tester | The recorded test result says nothing about the network | **Test connection** against a closed port, then against a host that accepts and never answers; after each, `GET /api/v1/ldap/config` | `last_test_result` is the same sentence both times — *Could not reach a directory server at that host and port…* 🛑 **Must NOT** contain *refused*, *timed out*, *dial* or the port — that difference is a port scan |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/`; the bind password is sealed like every secret; the directory is dialled through the directory policy (private yes; loopback only with `LDAP_ALLOW_LOOPBACK`, refused in production).
  - The connection is set up — LDAPS handshake, or StartTLS and its handshake — inside one deadline, the connection timeout; go-ldap puts none on a StartTLS handshake, so a middlebox that accepted the connection and went quiet used to hold a sign-in open indefinitely.
  - Only `invalidCredentials` (49) is a wrong password. A timeout, a dropped connection, or a domain controller answering `busy` / `unavailable` is *could not be reached* — for the person signing in and for **Test connection** — so an outage is never recorded as a failed sign-in.
  - 🔒 **Test connection** answers, and records as `last_test_result`, one authored sentence per kind of failure (`ConnectionTestMessage` in `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapService.go`), never the raw error. Until October 2026 the raw error was recorded and `GET /ldap/config` returned it, so refused, timed out and the port could still be read. A result stored before then is shown as the generic failure.
  - Filters are filled by `fillFilter` in `packetpulsego/pkg/common/ldapclient/LdapClient.go`: every `%s` (not only the first, as `fmt.Sprintf` did), `%u` in a group filter, `%%` as a percent sign; each value escaped.
  - A user search answered with `sizeLimitExceeded` is the ambiguity (`ErrAmbiguousUser`), not an outage. Groups are read in pages of 500 with no cap on the count; a cap of 100 used to lose every group past it. Directory failures at sign-in are logged with their reason (*directory sign-in could not be completed*).
  - Coverage: `packetpulsego/pkg/common/ldaptest/` (an in-process directory that answers as OpenLDAP or Active Directory: DN matching, AD's matching rules and sign-in names, size limits and paging, per-connection identity and access rules, unavailable binds, stalled StartTLS, silent hosts), `packetpulsego/pkg/common/ldapclient/LdapCustomerDirectory_test.go`, `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapDirectoryProducts_test.go`, `packetpulsego/pkg/usermicroservice/userservice/UserDirectorySignIn_test.go`, `packetpulseflutter/test/ldap_settings_screen_test.dart`.

---
## Group 7 — Platform and Settings

The operator of PacketPulse itself, each person's own settings, and the public face.

---

### 7.1 🏛️ Platform console and licences

**Screen:** Administer → **Platform** (superuser only) · **Routes:** `/platform/*` · **Access:** platform superuser

- 🌟 **Commercial Presentation & Sales Pitch**: Every customer runs on a licence PacketPulse signs, so nobody can grant themselves an organisation, more people or a longer term. Licences carry the people covered, sites, a period and a price, scale in proportion when sold for an unusual term, and keep a history of every change with who made it and why. A customer's server needs nothing from PacketPulse to check one: the signature is checked against a key built into the server.
- 📖 **User Guide & Operational Flow**: The console is PacketPulse's own server; a customer's server has none. **Organisations → Add**; **Issue licence** with plan, people, sites, months and currency. Per licence: **Suspend**, **Resume**, **Revoke**, **Renew**, change the people and sites it covers. **Download licence file** asks for the **Owner's email** and saves the signed file; copy it into `LICENCE_DIR` on the customer's server. It is read at every sign-in, so no restart is needed. The first person to sign up there with that address, proved by an emailed code, becomes the Administrator. Several files for one organisation do not add up: the one that runs longest is the licence, and of two that end the same day, the one issued later. An old file can stay beside its renewal, and what the files are called never matters.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `PLAT-001` | Superuser | Onboard a customer | Create `ACME`; issue a 12-month licence for 10 people; **Download licence file** naming `owner@acme.example`; copy it into a test server's `LICENCE_DIR`; sign up there with that address | The emailed code completes the sign-up, and the owner lands in ACME with every Administrator capability |
  | `PLAT-002` | Superuser | Suspend and resume | Suspend ACME's licence with a reason; download its file again; resume | While suspended: running a diagnostic is refused, reading is not. History shows both, with the reason |
  | `PLAT-003` | Superuser | A smaller licence keeps everyone but the owner out | ACME has 4 people; change the licence to 3; replace the file; an engineer signs in, then the owner | The engineer is refused (403, *Your organisation has more users than its licence covers…*). The owner signs in, because the owner is who can fix it |
  | `PLAT-004` | Superuser | Renewal scales the price | Renew a 12-month licence for 36 months | Three times the period price; the new end runs from the current end (or today if lapsed) |
  | `PLAT-005` | Superuser | A superuser is not a tenant | Call `GET /dnssite/list` as the superuser | 400 — no organisation. 🛑 **Must NOT** answer with every organisation's sites |
  | `PLAT-006` | Stranger | Signing up never makes or joins an organisation | On a customer's server, sign up with an address its licence does not name | Refused: *Only the owner named in this server's licence can sign up. Ask your administrator to add you.* 🛑 **Must NOT** create an account |
  | `PLAT-007` | Superuser, then Administrator | A licence for 20 refuses the 21st person | Issue ACME a licence for 20 people and install its file; the owner adds 19 colleagues; add a 21st. Then change the licence to 21, download the file again, replace it, and add the 21st again | 20th added; 21st refused (409 `seat_limit_reached`): *Your licence covers no more users…*, and **People on the licence** reads 20 of 20. After the new file, with no restart, the 21st is added and a 22nd is refused. 🛑 **Must NOT** add the 21st under the licence for 20, or leave an account for the refused person |
  | `PLAT-008` | Tester | An edited licence file admits nobody | Open ACME's installed file in a text editor; change `seat_limit` (or `valid_until`) and save; sign in, and add a person. Put the original file back | Both refused as no licence (403 `licence_expired`, *renew*), never as a full licence. With the original back, both work again. 🛑 **Must NOT** honour the edited number |
  | `PLAT-009` | Superuser, then Administrator | Licence files do not add up, and the later issue wins | ACME has a licence for 3 people, full. Change it to 5 people (its end date stays), download the file and save it **beside** the installed one under a name that sorts after it, e.g. `zz-acme.licence`. Add people until one is refused | The 4th and 5th are added and the 6th is refused: of two files ending the same day the later issue is the licence, and the two files' people are not added together. 🛑 **Must NOT** let the file names decide, or allow 8
  | `PLAT-010` | Superuser | A price with no digits is not free | Issue a licence with the amount typed as `.`, `-` or `+` | Refused on the amount. 🛑 **Must NOT** issue the licence at zero without *Complimentary* ticked |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/platformmicroservice/`; `packetpulseaccess.RequireSuperUser` on the console group (Appendix A); licence changes recorded in the licence history and the activity trail.
  - 🔒 Licence files are Ed25519-signed and verified by `packetpulsego/pkg/common/licence/Store.go` at every sign-in, against a key compiled into the server. `LICENCE_PUBLIC_KEY` may replace it only on a development build outside production, so a customer cannot swap in a key of their own.
  - Of one organisation's genuine files, `supersedes` in `Store.go` picks the licence: the latest `valid_until`, then the latest `issued_on`. Before October 2026 a same-day tie went to the first file name in sort order, so a customer could keep the bigger file of a downgrade just by naming it to sort first.
  - The console's own demo licence is dated from the UTC day (`packetpulsego/cmd/packetpulseserver/PacketPulseBootstrap.go`). Dated from the local day, a demo licence issued just after midnight in India did not start until UTC caught up, so the demo refused every sign-in for five and a half hours.
  - Coverage: `packetpulsetest/golang/tenancyassignment/` (the 20-user licence, and additions racing for the last place), `packetpulsego/pkg/usermicroservice/userservice/UserLicenceFile_test.go` (every state a licence file can be in, renewals beside and over the old file, one organisation's licence lending nothing to another), `packetpulsego/pkg/common/licence/`, `packetpulsego/pkg/platformmicroservice/**`, `packetpulseflutter/test/platform_console_screen_test.dart`.

---

### 7.2 ⚙️ Settings

**Screen:** Administer → **Settings** · **Routes:** `POST /user/appearance`, `POST /user/password`

- 🌟 **Commercial Presentation & Sales Pitch**: Five themes in light and dark — including a warm, low-blue-light theme for night shifts and a high-contrast one — every one contrast-checked, with status colours that never change meaning. They follow the person, not the device.
- 📖 **User Guide & Operational Flow**: Light, dark or follow the device; theme; density; reduce motion. **Change password** asks for the current password and the new one twice: 10 to 72 characters with a letter and a number, checked before anything is sent. It changes this account's password and nothing else: other devices stay signed in until they sign out or their 12 hours end, and the second step still follows the new password. Someone who signs in with their organisation's directory password is told to change it there, and offered no form; the owner, who keeps a local password under directory sign-in, keeps the form (`SET-004`, `SET-006`). PacketPulse is in English; there is no language choice. The same light / dark and appearance controls are on the sign-in screen, where they are remembered on that device until someone signs in.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SET-001` | Engineer | Appearance follows the person | Choose a dark theme on a laptop; sign in on another browser | The same theme there |
  | `SET-002` | Engineer | Status colours keep their meaning | Switch through every theme on a result with OK, Degraded and Breached rows | Green, amber and red in every theme, each with its icon |
  | `SET-003` | Engineer | English only | Look for a language picker in the bar and in Settings | None. Every label is English, served by the server. 🛑 **Must NOT** show a key such as `s812` in place of a label |
  | `SET-004` | Engineer | Change password | **Settings → Change password**: a wrong current password with a good new one; then the right one with a new one of nine characters (`नमस्ते 1a` too); then two new ones that differ; then a good change. Sign out; sign in with the old, then the new | Wrong current: *Your current password is incorrect.* under **Current password**, and still signed in. Nine characters: *at least 10 characters…* under **New password**; differing: *The two new passwords are not the same.*, neither sent. A good change: *Your password was changed…* and the fields cleared; the old password refused at sign-in, the new accepted (then the second step). 🛑 **Must NOT** sign the person out on a wrong current password |
  | `SET-006` | Engineer | A directory password is changed in the directory | With directory sign-in on, sign in as mapped staff, not the owner, and open **Settings**; try `POST /user/password` with their directory password as the current one; then do the same as the owner | Staff: *You sign in with your organisation's directory password, so change it in your directory…* and no form; the route answers 409 `password_from_directory`. The owner: the form, which changes their local password. 🛑 **Must NOT** accept a change that no sign-in would check |
  | `SET-005` | Visitor | Appearance before signing in | Signed out, use the sun/moon button and the palette on the sign-in screen (desk and phone width); reload; then sign in to an account set to another mode, and sign out | Each choice applies at once and survives the reload, with no request to `/user/appearance`. After sign-in the **account's** appearance is worn; after sign-out the sign-in screen keeps it. 🛑 **Must NOT** snap back to dark on sign-out, or save a signed-out choice to anyone's account |
- ⚙️ **Developer Guide & Release Confidence**:
  - Strings: `scripts/intellicodegen/packetpulsestrings.py` generates the Go catalogue and the Dart index; entries are positional, so retired ones stay in `DEPRECATED`. Help: `packetpulsego/pkg/initmicroservice/initconstants/PacketPulseHelp.go`.
  - Change password: `_PasswordForm` in `packetpulseflutter/lib/common/presentation/PacketPulseSettingsScreen.dart`; the session's `password_from_directory` (also on `/user/me`) decides whether it is shown. Server: `UserChangePassword` in `packetpulsego/pkg/usermicroservice/userservice/UserService.go`, held by `UserDirectorySignIn_test.go` against a real directory and by `UserRouteHandler_test.go` (422 and 409, never 401).
  - Coverage: `packetpulseflutter/test/settings_screen_test.dart`, the `translation` suite. Signed-out appearance: `packetpulseflutter/lib/common/services/PacketPulseAppearanceStore.dart`, tested in `packetpulseflutter/test/app_test.dart` and `packetpulseflutter/test/sign_in_test.dart`.

---

### 7.3 🌐 Public site and self-test

**Pages:** `https://packetpulse.rummaan53.com/` and `/selftest.html` (`packetpulseweb/`)

- 🌟 **Commercial Presentation & Sales Pitch**: The product site says only what the product does today, and leads straight to **Sign in**. The connection self-test still exists at `/selftest.html`, for a link sent by hand, but the site no longer advertises it (since 17 October 2026).
- 📖 **User Guide & Operational Flow**: **Sign in** opens the app at `/app/`. The site has no **Test my connection** link, in the header or the footer: both were removed on 17 October 2026, when the live host's built-in reflector was switched off for good. The self-test is reached only at `/selftest.html`; it measures DNS round trips to a public resolver, or the reach of any HTTPS host, and no longer offers the UDP test (real packet loss is measured in **Run diagnostic**, §4.2). The sun, moon and screen buttons in the header choose light, dark or match this device; the app's sign-in screen opens in the same choice. On a phone the header keeps the theme switch and **Sign in**.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `WEB-001` | Visitor | The site is current | Read the capabilities | Emailed or texted sign-in codes, IPv4 and IPv6, CSV into your systems, check-ins, testing from the engineer's own device. 🛑 **Must NOT** promise 23 languages, a hop-by-hop path or a line-speed test |
  | `WEB-002` | Visitor | Site and app are different documents | Open `/` and `/app/` | The marketing page and the app respectively — never the same document |
  | `WEB-003` | Visitor | The self-test runs in the browser | Open `/selftest.html`; open **Target**; run a resolver | **Target** lists *Cloudflare DNS (1.1.1.1)*, *Google Public DNS (8.8.8.8)* and **Another HTTPS host…**; round trips, jitter and loss for the resolver, measured from the visitor's connection. 🛑 **Must NOT** offer *PacketPulse network — UDP, like a call* (removed 17 October 2026; the UDP test is §4.2) |
  | `WEB-004` | Visitor | Light, dark or match this device | Choose **Light** on `/`; reload; open `/selftest.html`; then **Sign in**. Choose **Match this device** and switch the computer's own mode | Light at once, on both pages and on the app's sign-in screen; *Match this device* follows the computer as it changes. 🛑 **Must NOT** flash dark before a light page paints |
  | `WEB-005` | Visitor | The site fits a phone | Open `/` and `/selftest.html` at 360, 390 and 520 px wide and scroll to the foot of each | Logo, theme switch and **Sign in** on one line; the cards in one column; the footer links wrap. 🛑 **Must NOT** scroll sideways at any of them (the footer links and the cards' minimum width once pushed both pages wider than a phone) |
  | `WEB-006` | Visitor | Only the fields the method uses | Open `/selftest.html`; switch **Target** between a resolver and **Another HTTPS host…** | **Host or URL** only for another host; Samples, Interval and Timeout for both. 🛑 **Must NOT** show a hidden field (the page's CSS once overrode `hidden`) |
  | `WEB-007` | Visitor | The site does not advertise the self-test | Open `/` at full width and at 390 px; read the header, then scroll to the footer | The section links, the theme switch and **Sign in**; no *Test my connection* in the header or the footer. 🛑 **Must NOT** link `/selftest.html` from `/` (removed 17 October 2026) |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulseweb/index.html`, `packetpulseweb/selftest.html`; the probe engine's tests: `packetpulseweb/assets/packetpulse-probe.test.mjs` (part of `unit`); the UDP engine and its tests, `packetpulseweb/assets/packetpulse-udp.js` and `packetpulseweb/assets/packetpulse-udp.test.mjs` (§4.2), which no site page loads since 17 October 2026 but which stay, because the app runs a byte-identical copy (`packetpulseflutter/web/engine/packetpulse-udp.js`); change both together. The theme switch: `packetpulseweb/assets/packetpulse-theme.js`, which keeps the mode under the app's own appearance key; its tests, `packetpulseweb/assets/packetpulse-theme.test.mjs`, are part of `unit` too.
  - The deploy asserts `/` and `/app/` differ, so the nginx misroute that once served a blank app is caught by the deploy.

---

### 7.4 🎨 Branding

**Screen:** Configure → **Branding** · **Routes:** `GET /organisation/brand`, `PUT/DELETE /organisation/brand/{part}` · **Capability:** `staff_manage` to change; any member reads

- 🌟 **Commercial Presentation & Sales Pitch**: The customer's own logo heads every report their engineers attach to a ticket — the ticket's PDF — and their own icon sits in every one of their people's browser tabs. Another organisation on the same service never sees either.
- 📖 **User Guide & Operational Flow**: **Report logo**: **Upload image** (PNG, JPEG or GIF, up to 2400 by 1200 pixels). It is printed on the report's dark band before PacketPulse's name: white lettering suits it best, and a dark logo is printed on a white panel. **Browser icon**: up to 512 by 512, square works best; it replaces PacketPulse's tab icon for everyone in the organisation once they sign in, and PacketPulse's returns when they sign out. **Replace** and **Remove** each part on its own. Images are uploaded from PacketPulse in a web browser. Use only a logo your organisation has the right to use.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `BRAND-001` | Administrator | A logo heads every report | Upload a white-lettered PNG logo; export a ticket's PDF, a run's PDF and the month's availability PDF | The logo before *PacketPulse* on every page of all three, straight on the dark band |
  | `BRAND-002` | Administrator | A dark logo still shows | Upload a dark logo on a clear background; export a PDF | The logo on a white panel, readable. 🛑 **Must NOT** vanish into the band |
  | `BRAND-003` | Engineer | The tab wears the organisation's icon | Administrator uploads an icon; an engineer signs in; then signs out | The tab shows the organisation's icon after sign-in and PacketPulse's after sign-out |
  | `BRAND-004` | Administrator | What is not an image is refused | Upload an SVG, a text file renamed `.png`, and a 5000-pixel-wide PNG | Each refused under its own part with the reason; the stored logo unchanged |
  | `BRAND-005` | RIVAL engineer | A brand is the organisation's own | Sign in to RIVAL; export a PDF | No ACME logo or icon anywhere. 🛑 **Must NOT** show another organisation's brand |
  | `BRAND-006` | Engineer | Changing the brand needs `staff_manage` | As an engineer, `PUT /organisation/brand/logo` | 403; no **Branding** in the rail |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/orgbrand/`: `Normalise` checks the declared size before decoding (so a small file claiming a vast canvas costs nothing) and redraws every upload as a plain RGBA PNG; `DrawLogo` prints it, and a logo the PDF writer cannot read costs the report its logo, never the report. Stored in `organisation_brand` (migration 0029), one row per organisation.
  - Reports get the logo through `SetReportBranding` on the diagnostic and ping services. The tab icon: `packetpulseflutter/lib/common/services/PacketPulseFavicon.dart`, applied by the shell after sign-in and removed when it closes.
  - Coverage: `packetpulsego/pkg/common/orgbrand/`, `packetpulsego/pkg/staffmicroservice/staffapp/`, `packetpulsego/pkg/pingmicroservice/pingreport/`, `packetpulseflutter/test/brand_screen_test.dart`, `packetpulseflutter/test/shell_branding_test.dart`.

---

### 7.5 🧑‍💻 PacketPulse support sign-in

**Screen:** Administer → **Platform** → a row's support button (superuser only) · **Routes:** `POST /platform/organisation/{organisationId}/support` · **Access:** platform superuser

- 🌟 **Commercial Presentation & Sales Pitch**: When a customer's people cannot get in — an engineer locked out, every seat held, a setting that stops sign-in — PacketPulse can help from inside their organisation, in minutes, without asking for anyone's password. Every visit is for one hour, says why, and is written into the customer's own activity trail; their administrators see it while it lasts and can end it. Support takes no place on the licence.
- 📖 **User Guide & Operational Flow**: On **Platform**, a row's support button opens *Sign in to <organisation> as PacketPulse support*; **Reason \*** is required, up to 200 characters. **Sign in as support** switches the app to that organisation as its Administrator, with a banner on every screen: *PacketPulse support in <organisation>, until <time>* and the reason. **End support** in the banner — or the hour running out, or the organisation's administrator signing the session out — returns to the console. Support is the organisation's own **PacketPulse support** account: it is not in the organisation's people list, takes no place on its licence, has no password, and cannot be edited, switched off or deleted by the organisation. Its session shows under **Who is signed in** as *PacketPulse support*.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `PLAT-011` | Superuser | Sign in to a customer as support | On **Platform**, press ACME's support button; give *An engineer cannot sign in*; **Sign in as support** | The app shows ACME's screens as its Administrator, with the banner, the end time an hour away and the reason. ACME's **Activity** has a sign-in by `superuser@rummaan53.com` with the reason. 🛑 **Must NOT** read another organisation's data, or sign in as any of ACME's people |
  | `PLAT-012` | Superuser | A reason is required, in characters | Press **Sign in as support** with the reason empty; then with 201 Devanagari letters; then 200 | *Say why you are signing in to this organisation.*; then *Keep the reason to 200 characters or fewer.*; the 200 are accepted. 🛑 **Must NOT** open support without a reason |
  | `PLAT-013` | ACME Administrator | ACME sees support and can end it | While `PLAT-011` runs, open **Who is signed in** as ACME's administrator; sign the *PacketPulse support* session out | The session is listed; signing it out ends it at once — the operator's next action returns to the console. 🛑 **Must NOT** hide support's session from the organisation |
  | `PLAT-014` | Tester | Support is no one in the organisation | With support signed in to ACME: **Staff**, the people count on **Who is signed in**, and the console's ACME row; then `PATCH /user/{supportUserId}` and `PUT /staff/{supportStaffId}` as ACME's administrator, with the ids from **Activity** | Support appears in none of the lists or counts; both calls answer 404. 🛑 **Must NOT** count support on the licence, or let the organisation edit, switch off or delete it |
  | `PLAT-015` | Tester | The support account cannot be signed into | Sign in on the sign-in page as `support+<ACME's organisation id>@packetpulse.invalid` with any password, in ACME with and without a directory configured | *Incorrect email address or password.*, no code sent, and the account is not locked. 🛑 **Must NOT** ask the directory about it, or send a code |
  | `PLAT-016` | Tester | Support in one organisation reaches only that one | With support signed in to ACME, `GET /dnssite/list`, then `PUT` and `DELETE /dnssite/{id}` with RIVAL's endpoint id | ACME's endpoints only; both RIVAL calls 404. 🛑 **Must NOT** reach another organisation's rows |
  | `PLAT-017` | Superuser | Support ends back in the console | Press **End support**; then sign in as support again and wait an hour | Back on **Platform** as the operator each time, without signing in again. 🛑 **Must NOT** leave the operator signed in to the organisation, or send them to the sign-in page while their own session lasts |
  | `PLAT-018` | ACME Administrator | No customer can open support | `POST /platform/organisation/{RIVAL's id}/support` with a reason, as ACME's administrator | 403. 🛑 **Must NOT** open a session in any organisation |
- ⚙️ **Developer Guide & Release Confidence**:
  - The superuser never enters an organisation: `RequireOrganisationScope` still refuses it (`PLAT-005`). `SupportSignIn` in `packetpulsego/pkg/usermicroservice/userservice/UserSupport.go` opens a session for the organisation's own support account instead — made or repaired on demand by `SupportAccountEnsure` in `packetpulsego/pkg/staffmicroservice/staffdomain/repository/StaffSupportRepositoryPostgres.go` as the built-in Administrator — so every tenant check, gate and audit entry applies to it unchanged. The token lasts `SupportSessionLifetime` (one hour, `IssueFor` in `packetpulsego/pkg/common/security/TokenManager.go`); the session row keeps `support_by_user_id` and `support_reason`.
  - Migration `packetpulsego/pkg/common/dbclient/migrations/0034_2026_10_17_support_sign_in.sql`: `staff.is_support` (one per organisation), `auth_source = 'support'` on its credential, and the session's support columns. The people count, the people and user lists, every edit of a person, `PATCH /user/{userId}` and the demo's concurrency cap leave support out; the organisation's signed-in sessions do not.
  - 🔒 The password step refuses an `auth_source = 'support'` credential before any password or directory is consulted: in a directory-governed organisation the directory would otherwise be asked about the address's local part, a name anyone with rights in that directory could create (`TestTheSupportAccountCannotBeSignedInto`).
  - The organisation's trail records the sign-in under the operator, with the reason (`recordSupportSignIn`); the platform's own entry comes from the audit registry. The app keeps the operator's own session aside (`keepConsole` in `packetpulseflutter/lib/common/services/PacketPulseSessionStore.dart`) and the auth bloc's `PacketPulseAuthSupportEnded`, or the support session ending, puts it back.
  - Coverage: `packetpulsego/pkg/usermicroservice/userservice/UserSupport_test.go`, `packetpulsego/pkg/usermicroservice/userapp/UserSupport_test.go`, `packetpulsego/pkg/staffmicroservice/staffdomain/repository/StaffSupportRepositoryPostgres_test.go`, `packetpulsego/cmd/packetpulseserver/main_test.go` (through the real middleware chain), the `tenancy` suite (`TestSupportInOneOrganisationCannotReachAnother`), `packetpulseflutter/test/platform_console_screen_test.dart` and `packetpulseflutter/test/app_test.dart`.
## Feature → Coverage Matrix

What guards each area automatically, so a manual pass can spend its time where automation cannot reach: real devices, real browsers, real file servers, and judgement.

| Area | Manual IDs | Go unit and repository | Integration suites | Flutter |
|---|---|---|---|---|
| Sign-in, second step | `AUTH-*` | `packetpulsego/pkg/usermicroservice/userservice/` | `assignment`, `audit` | `packetpulseflutter/test/sign_in_test.dart`, `packetpulseflutter/test/app_test.dart` |
| SMS gateway, location rule | `SMS-*`, `LOC-*` | `packetpulsego/pkg/smsmicroservice/`, `packetpulsego/pkg/common/smsprovider/` | `acl` | `packetpulseflutter/test/sign_in_security_screen_test.dart` |
| Staff, roles | `STF-*`, `ACL-*` | `packetpulsego/pkg/staffmicroservice/` | `acl`, `tenancy` | `packetpulseflutter/test/staff_screen_test.dart` |
| Sessions and licence | `WHO-*` | `packetpulsego/pkg/staffmicroservice/` | `assignment`, `tenancy` | `packetpulseflutter/test/session_screen_test.dart` |
| Check-ins | `CHK-*` | `packetpulsego/pkg/common/geocode/` | `tenancy`, `audit` | `packetpulseflutter/test/checkin_screen_test.dart` |
| Activity | `AUD-*` | `packetpulsego/pkg/common/auditlog/` | `audit` | `packetpulseflutter/test/audit_log_screen_test.dart` |
| Sites, policy | `SITE-*` | `packetpulsego/pkg/dnssitemicroservice/`, `packetpulsego/pkg/common/probeguard/` | `tenancy` | `packetpulseflutter/test/dns_site_screen_test.dart`, `packetpulseflutter/test/shell_test.dart` |
| Diagnostics, IPv6 | `DIAG-*`, `V6-*` | `packetpulsego/pkg/pingmicroservice/`, `packetpulsego/pkg/diagnosticmicroservice/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_result_selection_test.dart` |
| CSV | `CSV-*`, `API-003` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_history_screen_test.dart` |
| Network reports | `RPT-*`, `EXP-013`–`EXP-017` | `packetpulsego/pkg/reportmicroservice/`, `packetpulsego/pkg/common/orgzone/`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract`, `tenancy`, `acl`, `audit` | `packetpulseflutter/test/network_report_screen_test.dart`, `packetpulseflutter/test/scheduled_reports_card_test.dart` |
| Run diagnostic, device test | `DEV-*` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract` | `packetpulseflutter/test/run_diagnostic_screen_test.dart`, `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/client_probe_live_panels_test.dart`, `packetpulseflutter/test/dns_site_nearest_test.dart` |
| UDP test, reflectors | `UDP-*`, `WEB-006` | `packetpulsego/pkg/reflectormicroservice/`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract`, `tenancy`, `acl`, `audit` | `packetpulseflutter/test/run_diagnostic_udp_test.dart`, `packetpulseflutter/test/udp_probe_run_test.dart` (the page: `packetpulseweb/assets/packetpulse-udp.test.mjs`, in `unit`) |
| Service targets and grading | `MON-*` | `packetpulsego/pkg/monitormicroservice/` | `grading` | `packetpulseflutter/test/diagnostic_charts_test.dart` |
| Results API | `API-*` | `packetpulsego/pkg/common/apikeyauth/` | `contract`, `tenancy` | `packetpulseflutter/test/api_key_screen_test.dart` |
| Result export | `EXP-*` | `packetpulsego/pkg/common/filedrop/`, `packetpulsego/pkg/exportmicroservice/` | `contract`, `acl`, `tenancy`, `audit` | `packetpulseflutter/test/result_export_screen_test.dart` |
| Directory | `LDAP-*` | `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/`, `packetpulsego/pkg/usermicroservice/userservice/UserDirectorySignIn_test.go` | — | `packetpulseflutter/test/ldap_settings_screen_test.dart` |
| Platform | `PLAT-*` | `packetpulsego/pkg/platformmicroservice/` | `assignment` | `packetpulseflutter/test/platform_console_screen_test.dart` |
| Support sign-in | `PLAT-011`–`PLAT-018` | `packetpulsego/pkg/usermicroservice/userservice/`, `packetpulsego/pkg/staffmicroservice/staffdomain/repository/`, `packetpulsego/cmd/packetpulseserver/` | `tenancy` | `packetpulseflutter/test/platform_console_screen_test.dart`, `packetpulseflutter/test/app_test.dart` |
| Settings, catalogue | `SET-*` | `packetpulsego/pkg/initmicroservice/` | `translation` | `packetpulseflutter/test/settings_screen_test.dart` |
| Branding | `BRAND-*` | `packetpulsego/pkg/common/orgbrand/` | `acl`, `tenancy` | `packetpulseflutter/test/brand_screen_test.dart` |

> [!NOTE]
> What no suite covers, and why manual cases exist for it: a **real SMS** arriving on a phone (`AUTH-009`), a **real emailed code** arriving in an inbox (`AUTH-001`), **browser location prompts** (`CHK-001`, `LOC-002`), **real SFTP and FTPS servers** (`EXP-006`, `EXP-011`), **spreadsheet applications** opening the CSV (`CSV-002`), and **IPv6 on the live host** (`V6-001`).

---
# Part III — End-to-End Journeys

Each journey crosses several chapters the way a real customer does. Run them on a disposable stack after the chapter cases; every step's result is the precondition of the next, so stop at the first failure and file it against the step's case ID.

---

### JRN-001 · A new customer, from contract to first evidence

**Personas:** platform superuser, ACME owner, NOC engineer · **Covers:** `PLAT-001`, `AUTH-008`, `AUTH-001`, `STF-001`, `SITE-002`, `DIAG-001`, `DIAG-004`, `CSV-001`

1. On the console, the superuser creates **ACME**, issues a licence for 10 people and downloads its licence file, naming the owner's address.
2. The file goes into `LICENCE_DIR` on ACME's server. The owner signs up there with that address and enters the emailed code.
3. The owner adds an engineer (*NOC Engineer*) and pastes twenty sites with **Bulk import**.
4. The engineer signs in with their emailed code and runs a diagnostic **from the server** with TT `TT-JRN-001`.
5. The engineer exports the **PDF** and the **CSV**.

**Expected:** the PDF and CSV carry `TT-JRN-001`; Activity shows the owner's sign-in, the staff and site changes, both exports. 🛑 The engineer's sign-in is **not** in Activity.

---

### JRN-002 · A NOC engineer works a ticket

**Covers:** `DIAG-001`–`DIAG-005`, `HIST-001`, `V6-001`

1. Ticket `TT-JRN-002` arrives: Customer `CUST-88412`, slow calls.
2. Run against the customer's sites. Read loss and jitter on each row.
3. Re-run after the carrier's fix.
4. In **History**, compare both attempts.

**Expected:** both attempts kept; the second's average loss and jitter lower; the dual-stack site shows its IPv6 row *not counted*.

---

### JRN-003 · A field engineer's day

**Covers:** `STF-004`, `LOC-002`, `CHK-001`–`CHK-002`, `DEV-004`–`DEV-006`, `DEV-017`

1. The administrator sets the engineer to **Record location** and turns on **Require location to sign in**.
2. On a phone, the engineer signs in, allowing location.
3. At the customer's site: **Run diagnostic**, type a Customer ID and `TT-JRN-003` at the top, keep the endpoint chosen for the site, **Start test**, then attach it.
4. Signs out, allowing location.

**Expected:** **Check-ins** shows the sign-in and sign-out with places and a map link; the ticket shows the device run, *measured on a device*, named after the endpoint nearest the site; the administrator's Activity shows no entry for the engineer's sign-in.

---

### JRN-004 · The IT system takes the results

**Covers:** `API-001`–`API-003`, `EXP-002`, `EXP-006`–`EXP-008`

1. The administrator issues a Results API key; the IT system pulls yesterday's CSV.
2. The administrator sets up **Result export** to the IT team's SFTP server: test, trust the key, save, every 15 minutes.
3. Engineers run diagnostics for an hour.
4. The SFTP password is changed on the server for one interval, then restored.

**Expected:** a file per interval with results; the interval with the wrong password shows *Not delivered*, and the next file holds that period's results too; every `result_id` appears exactly once across all the files — a failed delivery sent nothing, so nothing is sent twice.

---

### JRN-006 · An attack on an administrator's account

**Covers:** `AUD-003`–`AUD-006`, `AUTH-006`, `AUTH-007`

1. From an unfamiliar browser, try the owner's address with ten wrong passwords.
2. Then, with the right password (a leaked one), ten wrong codes.

**Expected:** Activity shows the wrong passwords and wrong codes as *Sign-in failed* with the address and browser, ending in *Locked after too many wrong codes*; the owner, from their own browser with an emailed code, is locked for 15 minutes and then signs in. 🛑 The wrong passwords alone did not lock the account.

---

### JRN-007 · A lost phone

**Covers:** `AUTH-012`, `WHO-003`, `STF-004`

1. An engineer set to text loses their phone, still signed in to the app on it.
2. Their administrator signs that device out (**Who is signed in → Sign device out**) and switches them to **Email**.
3. The engineer signs in on a laptop with the code emailed to them.

**Expected:** the lost phone's next request returns it to sign-in; the sign-out and the change of method are in Activity, naming who made them; the engineer's code arrives by email, not by text.

---

### JRN-008 · A licence lapses

**Covers:** `DIAG-006`, `PLAT-002`, `API-002`

1. The superuser suspends ACME's licence.
2. Engineers try to run a diagnostic; read and export old ones; the IT system pulls through the API.
3. The superuser resumes it.

**Expected:** runs refused with who can renew; reading, exports and the Results API keep working throughout.

---
# Part IV — Non-Functional Matrix

What every screen owes every person, whatever the feature.

## 4.1 Form factors

The app runs on the web (the live build), macOS, Windows, Android and iOS from one codebase. Below 600 px a list row reads as two lines; below 1000 px the navigation rail drops its group names.

| ID | Check | Expected |
|---|---|---|
| `NFR-001` | Every screen at 390 × 844 (a phone) | No horizontal page scroll; no overflow stripes; tables scroll within their card |
| `NFR-002` | Long values: a 39-character IPv6 address, a 60-character site name, a Devanagari name | One line, ellipsised, the whole value in the tooltip and selectable. 🛑 **Must NOT** break the row or truncate silently |
| `NFR-003` | The rail at 1600, 1000 and 600 px | Grouped with names; grouped with dividers; the drawer |

## 4.2 States every screen has

| ID | State | Expected |
|---|---|---|
| `NFR-004` | Loading | A skeleton, never a blank or a spinner over an empty frame |
| `NFR-005` | Empty | The designed empty state with what to do next |
| `NFR-006` | Failed | The failure with **Try again** when the server says it is worth retrying. 🛑 **Must NOT** show the empty state — "No sites yet" under a refusal reads as data loss |
| `NFR-007` | 403 | What the server says to do: a lapsed licence names who can renew; a missing capability says to ask an administrator |
| `NFR-008` | Forms | Required fields end with ` *`; a refused save names each field; a form whose load failed offers no **Save** |

## 4.3 Appearance and accessibility

| ID | Check | Expected |
|---|---|---|
| `NFR-009` | Each of the five themes in light and dark | Text contrast passes; status colours unchanged in meaning |
| `NFR-020` | A chosen chip or segment in each theme, light and dark — *SFTP (recommended)* on **Result export**, an endpoint chip, **Packets per second** | The label is white (dark navy in dark mode) on the blue fill, as legible as its tick; an unchosen chip keeps the page's text colour. Held by `packetpulseflutter/test/packetpulse_theme_test.dart` at 4.5:1 (review of 14 October 2026). 🛑 **Must NOT** draw dark text on the dark fill |
| `NFR-010` | Reduce motion on (app setting or the device's) | No transitions |
| `NFR-011` | Status by colour alone | Every pill carries an icon as well as a colour |
| `NFR-012` | Every user-visible sentence | From the catalogue: no English hard-coded in a widget, no raw key shown |

## 4.4 Security

| ID | Check | Expected |
|---|---|---|
| `NFR-013` | Tenancy, for every id-taking route | RIVAL's id answers 404 or an empty list — `./packetpulsetest.sh tenancy` |
| `NFR-014` | The destination policy, everywhere PacketPulse connects out | Sweeps, the relay to a customer's UDP reflector, the directory and the result export all refuse loopback, link-local and cloud metadata |
| `NFR-015` | Error responses | No SQLSTATE, driver names, file paths or a remote server's raw reply in any error envelope |
| `NFR-016` | Secrets at rest | SMS tokens, bind passwords and export credentials sealed with the server's secret box; API keys stored as hashes; never returned |
| `NFR-017` | CSV injection | Cells starting `=`, `+`, `-`, `@`, tab or carriage return are written as text |
| `NFR-018` | Production configuration | Production refuses `OTP_OUTBOX_FILE`, a `LICENCE_PUBLIC_KEY` override, `LDAP_ALLOW_LOOPBACK`, `EXPORT_ALLOW_LOOPBACK` and open CORS, and will not boot without an SMTP account |
| `NFR-019` | Text the database cannot hold | A NUL character (`\u0000`) in any field of a JSON request - a site name, a note, a password - is refused with *Text in the request cannot contain a NUL character* (400). 🛑 **Must NOT** answer 500 *try again* for input that can never be stored |

## 4.5 Performance — the load suite

`./packetpulsetest.sh load` provisions one licensed organisation — the owner and nineteen engineers, each signing in with a code read from the server's outbox, and 200 sites on TEST-NET-3 — and measures it. It also runs nightly (`.github/workflows/packetpulse-load.yml`). Budgets are 95th percentiles; the measured figures are from a developer laptop against a local server.

| Scenario | Measured p95 | Budget |
|---|---|---|
| Sign-in, password step — all 20 at once | 750 ms | 3 s |
| Sign-in, code step — all 20 at once | 16 ms | 1 s |
| Reads — list, sites, dashboard, me; 20 people × 10 | 9 ms | 750 ms |
| Sweep of 200 sites (measured when failures were also traced, before path analysis was removed) | 91 s | 120 s |
| Diagnostic of 5 sites — 10 people at once | 4.1 s | 20 s |
| Results API by ticket (200 results) — 200 pulls, 10 at a time | 16 ms | 1 s |
| CSV export of the whole run (250 rows) | 2 ms | 10 s |

🔒 Absolute whatever the machine: no 5xx; the 200-site sweep stores all 200 results; the CSV holds every stored result once. With the old bug restored (saving under the sweep's deadline) the sweep answers 500 and the suite goes red.

---
# Part V — Release Governance

How a release is judged ready, shipped, checked and — if it must be — rolled back.

## 5.1 The automated gate

A release candidate passes all of these, with **no skips**, before anything ships:

```bash
PACKETPULSE_TEST_URL=http://localhost:18080 PACKETPULSE_TEST_REQUIRE_SUPERUSER=1 ./packetpulsetest.sh
./packetpulsetest.sh load          # before a release that touches sweeps, sign-in or exports
```

`./packetpulsetest.sh` runs `tenancy assignment acl audit translation contract grading unit client backup docs`. Coverage floors are a ratchet (`scripts/PacketPulseCoverageFloor.py`, `scripts/PacketPulseFlutterCoverageFloor.py`): raised when real coverage rises, never lowered to make a build pass.

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
| `POST` | `/diagnostic/endpoint/{dnsSiteId}/reflector-key` | member | `dns_site_manage` | required | ✅ |
| `POST` | `/diagnostic/endpoint/{dnsSiteId}/webrtc-offer` | member | `diagnostic_run` | required | — |
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
| `POST` | `/monitor/sla/add` | member | `sla_manage` | required | ✅ |
| `GET` | `/monitor/sla/list` | member | `sla_view` | — | — |
| `DELETE` | `/monitor/sla/{slaPolicyId}` | member | `sla_manage` | — | ✅ |
| `PUT` | `/monitor/sla/{slaPolicyId}` | member | `sla_manage` | required | ✅ |
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
| `POST` | `/platform/organisation/{organisationId}/support` | superuser | — | — | ✅ |
| `GET` | `/platform/plan/list` | superuser | — | — | — |
| `POST` | `/reflector/session` | public | — | — | — |
| `GET` | `/result/bytt/{ttNumber}` | api key | — | — | — |
| `GET` | `/report/network` | member | `report_view` and `diagnostic_view_all` | — | — |
| `GET` | `/report/network.csv` | member | `report_export` and `diagnostic_view_all` | — | ✅ |
| `GET` | `/report/network.pdf` | member | `report_export` and `diagnostic_view_all` | — | ✅ |
| `GET` | `/report/settings` | member | `apikey_manage` | — | — |
| `PUT` | `/report/settings` | member | `apikey_manage` | — | ✅ |
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
| `sla_manage` | Create and edit SLA policies. | ✅ | — | — |
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
| `0030_2026_10_08_webrtc_udp_reflector.sql` | A browser's UDP test is stored as its own measurement method, `udp-webrtc`. |
| `0031_2026_10_09_udp_split_and_endpoint_reflector.sql` | A UDP test filed on a ticket keeps where its loss happened - up, down or late - and its jitter each way; an endpoint can name the reflector beside it. |
| `0032_2026_10_14_sla_breach_reasons.sql` | A graded result keeps why it was graded so: each service target it missed, with what was measured and the target (`ping_result.sla_breaches`). |
| `0033_2026_10_14_sla_checks.sql` | Every figure of a graded result is kept against its target, with where Degraded begins and its own verdict: `sla_breaches` becomes `sla_checks`, its entries marked breached. |
| `0034_2026_10_17_support_sign_in.sql` | PacketPulse support can sign in to an organisation: one support account per organisation (`staff.is_support`, `auth_source = 'support'`), and who opened each support session and why. |
| `0035_2026_10_22_network_reports.sql` | Network reports: an index from a result's run to its test, the time zone reports and exports are written in (`report_settings`), the daily, weekly, monthly and yearly drops (`report_delivery`), the reflector a UDP test measured to and the IP it was reached at (`ping_result.reflector_address`, `reflector_ip`), and what an endpoint shows of itself to those who do not manage endpoints (`dns_site.report_visibility`). |

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
| `REFLECTOR_UDP_PORT` | `0` (off) | Runs the UDP reflector behind the public self-test page on this UDP port (50000 on the live host). Off, the page's UDP test says *UDP testing is not enabled on this server*. The host firewall and the cloud security list must admit the port |
| `REFLECTOR_PUBLIC_IP` | — | The address browsers are told to send UDP to. A cloud instance sees only its private address, so the live host needs its public one here |
| `REFLECTOR_SESSIONS_PER_MINUTE` | `6` | UDP tests one address may start in a minute on the public reflector |
| `REFLECTOR_KEY_SECRET` | — (32+ characters) | Derives each endpoint's reflector key (§4.2). Without it the relay to customer reflectors and the key route answer 503. Changing it changes every key, so every installed reflector must be reinstalled |
| `REFLECTOR_NETWORK_HOSTS` | — | PacketPulse's own standalone reflectors, as IP addresses separated by commas (live `140.245.28.232`). An endpoint whose **UDP reflector address** is one of them, or a name resolving to one, is signed for with the network key rather than its own endpoint key, so the one reflector answers every endpoint of every organisation. Install it with the key `packetpulseserver reflector-network-key` prints on the server. A name, not an IP, stops the server at boot |

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
