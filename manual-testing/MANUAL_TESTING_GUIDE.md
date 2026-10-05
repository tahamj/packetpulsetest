<div align="center">

# 📡 PacketPulse Manual Testing Guide

### *& Master Knowledge Manual*

**The single source of truth for testing, understanding, selling, operating, and releasing PacketPulse.**

<br>

`v2026.10-PROD-v2`  ·  `Verified against source 2026-10-05`

<br>

| 🧭 Screens | 🔌 API routes | 🔐 Capabilities | 🗄️ Migrations | 💬 Catalogue strings | ❓ Help topics | 🧪 Guard suites |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **16** <br><sub>in the navigation rail</sub> | **104** <br><sub>under `/api/v1`</sub> | **17** <br><sub>3 built-in roles</sub> | **27** <br><sub>applied at boot</sub> | **881** <br><sub>English, server-served</sub> | **18** <br><sub>one per screen</sub> | **11** <br><sub>+ `load`, opt-in</sub> |

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
| **One Run diagnostic screen** | *Run diagnostic* and *Test from this device* are one destination with two modes: **From the server** and **From this device**. Someone without the right to run a sweep gets the device mode alone. | `DEV-*`, `DIAG-*` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart` |

---

## 📋 What Changed in v1 — the October 2026 Customer Release

> [!IMPORTANT]
> **Deployed to https://packetpulse.rummaan53.com on 2026-10-04.** It answers eight requirements from Northwind Telecom plus their remark that administrator logins be tracked. Every row below must be verified before a release is signed off.

| Requirement | What changed | Where to test | Source of truth |
|---|---|---|---|
| **1. Show IPv6** | Client IPs are read only from trusted proxies and shown canonically; a dual-stack site is measured over **both** families, the IPv6 result **reported, not counted**; the device test shows its IPv4 and IPv6 egress. | `V6-*`, `DEV-004` | `packetpulsego/pkg/common/probeguard/ProbeGuardPolicy.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql` |
| **2. Two-step sign-in** | Every sign-in has a second step. *v2 replaced the authenticator app and recovery codes with an emailed or texted code.* | `AUTH-*`, `SMS-*` | `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0019_2026_10_03_two_step_sign_in.sql` |
| **3. No target dropdown** | The device test measures the target the organisation chose (Cloudflare DNS by default), with one-tap presets for whoever may change it. *In v2 it is the From this device mode of Run diagnostic.* | `DEV-*` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart` |
| **4. Location at check-in** | Sign-in and sign-out record the device's position for people set to *Record location*; field engineers on a separate **Check-ins** screen with address and map link; a per-organisation *Require location* rule. | `CHK-*`, `LOC-*` | `packetpulsego/pkg/common/dbclient/migrations/0020_2026_10_03_session_checkin.sql`<br>`packetpulsego/pkg/common/geocode/Geocode.go` |
| **5. Loss and jitter** | Loss and RFC 3550 jitter on every result, ticket, history row, dashboard row, PDF and SLA report. | `DIAG-005`, `HIST-*`, `MON-*` | `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql` |
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
| `monitor` | SLA grading, schedules, alert damping, maintenance windows | server |
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
| **Loss** | Packets sent minus received, as a percentage. On a device test it is *query* loss, because a browser cannot send a packet. | 0% |
| **Round trip** | Average, fastest and slowest response time, in ms. | Under the site's target |
| **Jitter** | RFC 3550 interarrival jitter — how much *consecutive* round trips differ. Voice cares about this more than raw latency. | Low single digits of ms |
| **MOS** | Mean Opinion Score (ITU-T G.107 E-model): how a phone call would sound over this path, 1–5. | Above 4.0; below 3.6 is noticed on a call |
| **SLA** | **OK**, **Degraded** (within 80% of a threshold) or **Breached**, against the site's targets. | OK |

## L.4 IPv4, IPv6 and "not counted"

A site whose name resolves to both an IPv4 and an IPv6 address is measured over **both**. The IPv6 result is shown, kept and exported, but marked **not counted**: the site's totals, alerts and SLA availability are carried by IPv4, because the customer's agreement is for the service and IPv4 still carries it. A site that is IPv6 *only* counts over IPv6 — otherwise it could never fail. When the server itself has no IPv6 route, an IPv6 address is reported as *could not be tested from here* rather than as an outage.

## L.5 Two vantage points

**Run diagnostic** has two modes. **From the server** measures from PacketPulse's data centre. **From this device** measures from wherever the person is, over their own connection — the right tool for "is it slow for me?". The two will not match, and both are correct: they measure different layers from different places. A device test can be **attached** to a ticket, where it sits beside the server's figures, labelled as measured on a device.

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
| 3 · Diagnostics | Sites · Diagnostics and results · IPv6 · CSV · History and dashboard | `SITE-*`, `DIAG-*`, `V6-*`, `CSV-*`, `HIST-*` |
| 4 · The customer's side | Run diagnostic, from this device | `DEV-*` |
| 5 · Monitoring | SLA targets, schedules, alerts, maintenance, SLA report | `MON-*` |
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
  | `AUTH-014` | Engineer | An ended session returns to sign-in once | Revoke the engineer's session from another device, then use the app | One clean return to the sign-in page, and one `POST /user/signout` in the browser's network panel. 🛑 **Must NOT** loop between sign-in and an error, or send sign-outs over and over |
  | `AUTH-015` | Tester | A forged client address is ignored | `curl -H 'True-Client-IP: 6.6.6.6' -H 'X-Forwarded-For: 6.6.6.6'` a sign-in from a machine that is not a trusted proxy | Sessions and Activity show the real peer address. 🛑 **Must NOT** record `6.6.6.6` |
  | `DEMO-001` | Visitor | The public demo signs in through the real two-step flow | On the console with `SEED_DEMO` on, sign in as `demo-engineer@<domain>` with the shared demo password | The real SMS step appears with the code **shown on screen** and filled in; one tap opens the app. 🛑 **Must NOT** text or email anything |
  | `DEMO-002` | Tester | The demo code is not a master key | Use the demo code (`DEMO_SMS_CODE`, default `111111`) as the second-step code for a **non-demo** account | Refused as a wrong code. 🛑 **Must NOT** admit any account outside the demo organisation |
  | `DEMO-003` | Tester | The demo is capped | Open `DEMO_SEATS` (default 3) concurrent demo sessions, then sign in once more | The extra sign-in is refused with *the demo is busy* (`demo_busy`). 🛑 **Must NOT** let shared demo accounts open unlimited sessions |
  | `DEMO-004` | Tester | The shared demo cannot be hijacked or turned on the network | As a demo account try **Change password**; then run a diagnostic against a private/internal address added to the demo inventory | Password change refused (`demo_read_only`); the private target is refused by the probe policy. 🛑 **Must NOT** let a visitor change a shared credential or probe `10.x`/internal space |
- ⚙️ **Developer Guide & Release Confidence**:
  - Flow: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go` — the password step returns a challenge (`email` or `sms`), never a session; `newSession` runs only after verify, and checks the licence again (§7.1). Challenge ids are 32 random bytes, stored as SHA-256; codes are stored as keyed hashes bound to their challenge.
  - 🔒 Attempts are counted **before** the code is checked, in one conditional `UPDATE … RETURNING`, so parallel guesses cannot slip under the limit.
  - Limits: `maxAttempts = 5`, `lockAfter = 10`, `lockFor = 15m`, `resendAfter = 30s`, `maxSendsPerChallenge = 3`, `maxCodesPerHour = 10` by either channel; a separate rate limiter for the second step (`SECOND_STEP_ATTEMPTS_PER_MINUTE`).
  - Email: `packetpulsego/pkg/common/mailer/Mailer.go` — STARTTLS with a verified certificate (credentials are never sent otherwise), header injection refused, the load spread over up to four sending accounts (`SMTP_FROM_0…3`, `SMTP_PASSWORD_0…3`). A development or test server writes codes to `OTP_OUTBOX_FILE` instead (`packetpulsego/pkg/common/mailer/Outbox.go`); production refuses it.
  - Client IPs: `packetpulsego/pkg/common/apiratelimit/ApiRateLimit.go` walks `X-Forwarded-For` from the right past `TRUSTED_PROXIES` only, and returns canonical addresses (`::ffff:a.b.c.d` → `a.b.c.d`).
  - Coverage: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps_test.go`, `packetpulsego/pkg/common/mailer/Mailer_test.go`, the `assignment` and `audit` suites (every integration test finishes its sign-in with the code from the outbox).

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

- 🌟 **Commercial Presentation & Sales Pitch**: Seventeen plain-language capabilities, three built-in roles (Administrator 17, NOC Engineer 7, Viewer 5) and as many of the operator's own as it likes — plus a per-person *allow* or *deny* for the exception that does not deserve a role. Changes take effect on the next request, not the next sign-in.
- 📖 **User Guide & Operational Flow**: The matrix lists roles across and capabilities down. **Add a role**, tick what it may do, save — the whole set is saved, so nothing is left to an invisible default. Built-in roles are read-only: an engineer reads the tests they ran (`diagnostic_view_own`), while Administrator and Viewer read everyone's (`diagnostic_view_all`). A role somebody holds cannot be deleted. Per person: *Inherit*, *Allow* or *Deny* each capability.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `ACL-001` | Administrator | Built-in roles cannot be changed | Try to edit *Viewer* | Not editable; `PUT /staff/role/{viewerId}` answers 403 or 409. 🛑 **Must NOT** change a role every organisation shares |
  | `ACL-002` | Administrator | A custom role grants exactly what is ticked | Add *Shift lead* with `diagnostic_run` and `report_view` only; give it to an engineer | They can run diagnostics and see the dashboard; **Staff**, **Monitoring** edits and **API keys** are absent. Each refused route answers 403 |
  | `ACL-003` | Administrator | A role in use cannot be deleted | Delete *Shift lead* while someone holds it | Refused, naming that people hold it |
  | `ACL-004` | Administrator | A per-person deny wins over the role | Deny `report_export` to one engineer | That engineer has no **Export PDF**/**Export CSV**; `GET /diagnostic/{id}/report.csv` answers 403. Other engineers unaffected |
  | `ACL-005` | Tester | Every role, both directions | Run `./packetpulsetest.sh acl` | Passes — each role is refused what it lacks and allowed what it holds |
  | `ACL-006` | Administrator | An engineer reads their own tests; an administrator everyone's | Two engineers each run a diagnostic; each opens **History**; then the administrator does | Each engineer sees only their own, with *These are the tests you ran*; the administrator sees both (`HIST-005`). 🛑 **Must NOT** show an engineer a colleague's test, in the list, by its id or by its ticket |
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
  - The count is `StaffCountUsers` — staff rows not deleted, switched on or off (`packetpulsego/pkg/staffmicroservice/staffdomain/repository/StaffRepositoryPostgres.go`). It is checked when a person is added or revived (`ensureRoomForAnotherUser` in `packetpulsego/pkg/staffmicroservice/staffservice/StaffService.go`) and at every sign-in against the installed licence file (`enforceUserLimit`, the owner excepted, §7.1).
  - Coverage: `packetpulsetest/golang/tenancyassignment/`, `packetpulsetest/golang/tenancyisolation/`, `packetpulseflutter/test/session_screen_test.dart`.

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

**Screen:** Configure → **DNS sites** · **Routes:** `GET /dnssite/list`, `POST /dnssite/add`, `POST /dnssite/bulkimport`, `PUT/DELETE /dnssite/{dnsSiteId}` · **Capabilities:** `dns_site_view`, `dns_site_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Paste a whole inventory in one go — one bad line never rejects the rest — and PacketPulse refuses to be turned against its own host: loopback, link-local and cloud-metadata addresses are never probed, whatever a site says.
- 📖 **User Guide & Operational Flow**: **Add site** with a name, an IP address or hostname, an optional circuit ID and SLA policy. **Bulk import** takes lines, commas or spaces; `Branch 12=10.0.0.1` names a site. Each entry is reported as added, a duplicate, or invalid with the reason.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SITE-001` | Administrator | Add by address and by name | Add `Mumbai POP` = `203.0.113.10`; add `Resolver` = `one.one.one.one` | Both listed; the hostname is resolved at test time, not stored as an address |
  | `SITE-002` | Administrator | One bad line does not sink a paste | Bulk import `Pune=203.0.113.11`, `not an address`, `Pune=203.0.113.11` | First added, second invalid with a reason, third a duplicate |
  | `SITE-003` | Administrator | Labels with spaces survive | Bulk import `Branch 12=10.0.0.12` | One site named `Branch 12`. 🛑 **Must NOT** split into `Branch` and `12` |
  | `SITE-004` | Engineer | The host is never a target | Add `127.0.0.1`, `169.254.169.254` and `::1`; run a diagnostic over them | Each result is a refusal (*destination refused by probe policy*), never a measurement. 🛑 **Must NOT** send traffic to loopback or cloud metadata |
  | `SITE-005` | RIVAL administrator | The same address in two organisations | RIVAL adds `203.0.113.10` too | Allowed — the endpoint is unique per organisation, not globally |
- ⚙️ **Developer Guide & Release Confidence**:
  - Policy: `packetpulsego/pkg/common/probeguard/ProbeGuardPolicy.go` — every address a name resolves to is checked, and one refused address refuses them all; IPv4-mapped IPv6 is judged as IPv4.
  - Coverage: `packetpulsego/pkg/dnssitemicroservice/**`, `packetpulsego/pkg/common/probeguard/`, `packetpulseflutter/test/dns_site_screen_test.dart`.
  - ⚠️ **TRAP** — a licence carries a **site limit** (shown on the Platform console), but adding or importing sites does not check it yet. Do not file a site count above the limit as a regression; it is a known gap.

---

### 3.2 🩺 Diagnostics and results

**Screen:** Operate → **Run diagnostic** → **From the server** · **Routes:** `POST /diagnostic/submit`, `GET /diagnostic/{requestId}`, `GET /diagnostic/{requestId}/report.pdf` · **Capabilities:** `diagnostic_run`, `diagnostic_view_own` or `diagnostic_view_all`, `report_export` · **Licence:** required to run

- 🌟 **Commercial Presentation & Sales Pitch**: One form, one sweep, one report against the ticket. Every figure a NOC argues about — loss, latency, RFC 3550 jitter, MOS — from the server and, in the same screen, from the engineer's own device. A lapsed licence stops new tests but never takes away the evidence already gathered.
- 📖 **User Guide & Operational Flow**: **Run diagnostic** opens on **From the server** for anyone who may run a sweep; **From this device** beside it is the device test (§4.1), and switching between them keeps what each holds. Enter **Customer ID** and **TT number**, pick sites (or leave empty for every enabled site), and choose packet count and timeout. The result shows headline cards (sites reachable, average loss, average jitter), then a row per site and family: reachable, the packet line verbatim (*Sent = 4, Received = 4, Lost = 0*), round trips, jitter, MOS with its band and SLA grade. **Export PDF** and **Export CSV** sit at the top.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DIAG-001` | Engineer | A sweep files against the ticket | Run against every enabled site with TT `TT-MAN-001` | Status *Completed*; one row per site (and family); the packet line exactly as specified |
  | `DIAG-002` | Engineer | A ticket keeps every attempt | Run `TT-MAN-001` again | **History** for the ticket shows both attempts, newest first |
  | `DIAG-003` | Engineer | A failure is reported, not traced | Include a black-holed address (`203.0.113.99`); run; open the result, its PDF and the **Dashboard** | The row shows the failure and its packet line. No path, hops or fault verdict anywhere, and no *Where the faults lay* card. 🛑 **Must NOT** offer a *Trace failures* switch |
  | `DIAG-004` | Engineer | The PDF is evidence | **Export PDF** | Named `packetpulse-<TT>-<UTC time>.pdf`; Customer ID, TT, UTC timestamps, Jitter and MOS columns, average loss and jitter cards; a long IPv6 address wraps onto two lines; the foot names who ran it and where — *Triggered by Asha Rao  -  at MG Road, Pune*. 🛑 **Must NOT** truncate an address |
  | `DIAG-005` | Engineer | Loss and jitter on the ticket | Open the result; open **History** | Headline cards show average loss and average jitter; the history row shows the same figures |
  | `DIAG-006` | Engineer | A lapsed licence stops new runs only | Platform suspends the licence; run a diagnostic; open an old one and export it | Run refused (`licence_suspended`, who can renew named); the old result opens and exports. 🛑 **Must NOT** hide recorded evidence |
  | `DIAG-007` | RIVAL engineer | Another organisation's ticket is not found | Open `/diagnostic/<ACME request id>` as RIVAL | 404. 🛑 **Must NOT** reveal that the ticket exists |
  | `DIAG-008` | Tester | A malformed id is refused before it is looked up | `GET /diagnostic/not-a-uuid` | 400 |
  | `DIAG-009` | Engineer | The PDF is made from the record when it is asked for | Export a ticket's PDF twice, a minute apart; look for a stored copy on the server | Both carry the same measurements, drawn from the stored results at the moment each was asked for; no PDF is kept on the server. 🛑 **Must NOT** depend on a file kept on disk |
  | `DIAG-010` | Engineer | A colleague's test cannot be exported | As an engineer, `GET /diagnostic/<colleague's request id>/report.pdf` (and `.csv`) | 404, as though it did not exist. 🛑 **Must NOT** hand an engineer someone else's evidence |
  | `DIAG-011` | Engineer, then Viewer | One screen, two modes | As an engineer: open **Run diagnostic**, type a Customer ID, switch to **From this device** and back. As a Viewer: open **Run diagnostic** | The engineer starts on **From the server**, and the Customer ID is still there after switching back. The Viewer gets the device test with no switch. 🛑 **Must NOT** offer a separate *Test from this device* entry in the rail |
- ⚙️ **Developer Guide & Release Confidence**:
  - Engine: `packetpulsego/pkg/pingmicroservice/pingprobe/PingProbeRunner.go` (pro-bing, unprivileged ICMP, TCP fallback); jitter: `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go` (`InterarrivalJitter`, RFC 3550).
  - 🔒 Results are saved under their **own** deadline, never the sweep's: a sweep that runs long still stores everything it measured (`load` suite §4.5 proves it, and goes red with the old bug restored).
  - The screen is `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart`. A mode is built when first shown and kept, so switching never loses a half-filled form or a finished device run. The server mode needs an organisation and `diagnostic_run`, as its own destination used to.
  - Path analysis was removed in `packetpulsego/pkg/common/dbclient/migrations/0027_2026_10_05_path_analysis_removed.sql`: `ping_hop` and the fault verdict columns are no longer written, and the rows already stored stay. A client built before then still sends `trace_failures`; the server accepts and ignores it, because decoding is strict and refusing it would fail every run that client made.
  - Ticket figures (`avg_loss_pct`, `max_loss_pct`, `avg_jitter_ms`) are computed in `RequestFinish` from the run's counted results (`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql`).
  - The PDF is built on the fly from the database each time it is asked for (`packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`); nothing is written to disk. Its foot names the person who ran the test and the place their session checked in from.
  - Coverage: `packetpulsego/pkg/diagnosticmicroservice/**`, `packetpulsego/pkg/pingmicroservice/**`, `packetpulseflutter/test/diagnostic_submit_screen_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart`, the `tenancy` and `contract` suites.

---

### 3.3 🌍 IPv4 and IPv6

**Where:** every result list, the PDF, the SLA report, the Results API and the CSV

- 🌟 **Commercial Presentation & Sales Pitch**: A site healthy over IPv4 and dark over IPv6 is a real, common, hard-to-prove fault. PacketPulse measures both, side by side, without letting an IPv6 problem the customer did not buy an SLA for change their availability figure.
- 📖 **User Guide & Operational Flow**: A dual-stack site has two rows. The IPv6 one carries a **not counted** tag (hover for why). The PDF marks it `icmp v6 *` with a footnote; the SLA report lists `(v6)` lines separately and `(v6*)` for report-only ones.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `V6-001` | Engineer | Both families, one counted | Add a dual-stack site (e.g. `one.one.one.one`) on a host with IPv6; run | Two rows: IPv4 counted, IPv6 tagged *not counted* with its tooltip |
  | `V6-002` | Engineer | An IPv6 failure does not breach the site | A dual-stack site whose IPv6 fails while IPv4 answers | No alert; the site counts as reachable; availability unchanged. 🛑 **Must NOT** raise an alert for report-only IPv6 |
  | `V6-003` | Engineer | IPv6-only counts | A site with only an IPv6 address that fails | Counted as a failure, alerting and grading as usual |
  | `V6-004` | Engineer | No IPv6 route is said plainly | On a server without IPv6, a dual-stack site | The IPv6 row reads *There is no IPv6 route from this vantage point, so this IPv6 address could not be tested from here.* 🛑 **Must NOT** report it as the site being down |
- ⚙️ **Developer Guide & Release Confidence**:
  - `ping_result.ip_version` (4, 6 or NULL when nothing was probed) and `report_only`; the daily rollup is keyed by family (`packetpulsego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql`).
  - `probeguard.ResolveAllAndCheck` returns one checked address per family; the runner probes them in parallel and marks IPv6 report-only only when there is more than one family.

---

### 3.4 📄 CSV download

**Where:** **Export CSV** on every result · **Route:** `GET /diagnostic/{requestId}/report.csv` · **Capability:** `report_export`

- 🌟 **Commercial Presentation & Sales Pitch**: The same results, in the columns an IT system imports, in one click — and safe to open: a site name typed as a spreadsheet formula is written as text, so an export can never run code on someone else's desk.
- 📖 **User Guide & Operational Flow**: **Export CSV** saves `packetpulse-<TT>-<UTC time>.csv` beside the PDF of the same ticket. One row per site and family; empty cells (not zeros) where a probe had no reply.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CSV-001` | Engineer | The file has the contracted header | **Export CSV**; open it | Header exactly as in `packetpulsetest/contracts/result_export_columns.json`, `result_id` first; times in UTC (`…Z`); an unreachable row has empty round-trip cells |
  | `CSV-002` | Engineer | A formula is written as text | Name a site `=HYPERLINK("http://x","y")`; run; export; open in Excel or Sheets | The cell shows the text starting `'=`. 🛑 **Must NOT** become a live formula or link |
  | `CSV-003` | Engineer | Names in any script survive | Name a site `पुणे केंद्र`; export | The name intact (UTF-8). 🛑 **Must NOT** show mojibake |
  | `CSV-004` | Viewer | No export without `report_export` | Open a result as a Viewer | No **Export CSV** or **Export PDF**; the route answers 403 |
  | `CSV-005` | Tester | The download is on the record | Export a CSV; open **Activity** | An *Exported* entry naming the diagnostic |
- ⚙️ **Developer Guide & Release Confidence**:
  - One writer for every CSV: `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`; one projection and scanner: `packetpulsego/pkg/diagnosticmicroservice/diagnosticdomain/repository/DiagnosticExportPostgres.go` (the run's organisation must match the ticket's).
  - The browser download is `text/csv` (`packetpulseflutter/lib/common/services/PacketPulseFileSaverWeb.dart`); desktop and phone use the share sheet.

---

### 3.5 🗂️ History and dashboard

**Screens:** Operate → **History**, **Dashboard** · **Routes:** `GET /diagnostic/list`, `GET /diagnostic/tt/{ttNumber}`, `GET /ping/dashboard`

- 🌟 **Commercial Presentation & Sales Pitch**: Every ticket investigated, with its loss and jitter on the row, who ran it and where — and 30 days of availability on the dashboard for the supplier review.
- 📖 **User Guide & Operational Flow**: An engineer's **History** is the tests they ran, with the note *These are the tests you ran. Administrators see everyone's.* An administrator's is everyone's, and narrows by **Person**, **Place** (*Where it was run, or a site or region*), **Status**, **From** and **To**; **Clear filters** puts them back. Both search by TT number or Customer ID. Each row shows reachable/total, breaches, loss and jitter, and who ran it and where: *Customer CUST-1 · Asha Rao · at MG Road, Pune*. **Dashboard** shows sites, the last sweep, the licence, 30 days of availability (a day with nothing measured is a gap, never zero).
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `HIST-001` | Engineer | Find a ticket | Search `TT-MAN-001` | Its attempts, with loss and jitter on each row |
  | `HIST-002` | Engineer | A failed search is not an empty result | Stop the API; search | The failure, with *Try again*. 🛑 **Must NOT** show *No diagnostics yet* |
  | `HIST-003` | Engineer | A device run reads as one | Attach a device test (§4.1); find it in History | Marked as measured on a device; its loss is *query* loss |
  | `HIST-004` | Engineer | A quiet day is a gap | A schedule paused for a day; open **Dashboard** | The availability line has a gap. 🛑 **Must NOT** plot 0% for a day with nothing measured |
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
## Group 4 — The Customer's Side

Measuring from where the person is, not from where PacketPulse is.

---

### 4.1 📱 From this device

**Screen:** Operate → **Run diagnostic** → **From this device** · **Routes:** `GET /organisation/settings`, `PUT /organisation/settings` (`staff_manage`), `POST /diagnostic/clientobservation` (`diagnostic_run`, licensed)

- 🌟 **Commercial Presentation & Sales Pitch**: "Is it slow for me?" answered from the customer's own connection, in one tap, with no target to choose: everyone in the organisation measures the same host, so results compare. It shows the device's IPv4 **and** IPv6 addresses, then files the run against the ticket beside the server's figures — labelled as measured on a device.
- 📖 **User Guide & Operational Flow**: Open **Run diagnostic** and choose **From this device**; someone who may not run a sweep, or has no organisation, gets this mode alone, with no switch. It names what it is **Testing against** — the organisation's host, or *Cloudflare DNS (1.1.1.1)* with a note when none is chosen. **Start test** runs it. Whoever manages people sees **Change target**: one tap for *Cloudflare DNS* or *Google DNS (dns.google)*, or *Your own host* to type one. A finished run can be attached to a TT number. The line-speed test is hidden.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DEV-001` | Engineer | No target to choose | Open the screen | *Testing against* names one host; there is no dropdown or free host field. With none chosen: *Cloudflare DNS (1.1.1.1)* and the default note |
  | `DEV-002` | Administrator | Presets, and your own host | **Change target** → *Google DNS*; save. Reopen → *Your own host* → type `noc.northwind.example`; save | First saves `dns.google`; second saves the typed host. While *Your own host* is chosen with the field still empty, that chip — not Cloudflare — is lit |
  | `DEV-003` | Administrator | The target saves alone | Change the target | The *Require location* rule (§1.2) is unchanged. 🛑 **Must NOT** reset another setting |
  | `DEV-004` | Engineer | Both addresses | Run from a dual-stack connection; then from one without IPv6 | *IPv4 a.b.c.d · IPv6 2401:…*; then *No IPv6 connectivity* |
  | `DEV-005` | Engineer | The speed test is hidden | Open **From this device**; then open a ticket that had a speed filed before the test was hidden | No **Measure speed** button and no ~12 MB note. The older ticket still shows its line speed. 🛑 **Must NOT** move any data to the speed-test endpoints |
  | `DEV-006` | Engineer | Attach to a ticket | Run, then attach to `TT-MAN-002` | The ticket shows the device run, *measured on a device*, with no line speed; loss is *query* loss |
  | `DEV-007` | Viewer | Only people managers change the target | Open as a Viewer | No **Change target**. `PUT /organisation/settings` answers 403 |
  | `DEV-008` | Engineer | An unreadable setting does not block the test | Stop the API after the screen loads; reopen it | *Your organisation's target could not be read, so this tests Cloudflare DNS*; the test still runs |
  | `DEV-009` | Superuser | No organisation, still a test | Sign in as the platform superuser; open **Run diagnostic** | The device test alone, with no mode switch, testing Cloudflare DNS. It needs no organisation and no licence |
- ⚙️ **Developer Guide & Release Confidence**:
  - A browser cannot send ICMP, so the web build measures DNS-over-HTTPS (Cloudflare) or an HTTPS reach to the chosen host; the app on a desktop or phone can also ping it. Jitter is RFC 3550 with the standard deviation beside it.
  - Target normalisation (pasted URL → host, lower case, canonical IPs, zones refused): `NormaliseTestTarget` in `packetpulsego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`.
  - Speed, hidden: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientThroughput.dart` (Cloudflare `__down`/`__up`, median after a warm-up) is kept and tested behind `speedTestEnabled = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`; turning it back on is that one line. Egress: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientEgress.dart`.
  - Coverage: `packetpulseflutter/test/client_probe_screen_test.dart`, `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart`, `packetpulseflutter/test/client_throughput_test.dart`, `packetpulseflutter/test/client_egress_test.dart`.

---
## Group 5 — Monitoring

Turning on-demand testing into a continuous watch, with alerts that mean something.

---

### 5.1 📈 SLA targets, schedules, alerts and maintenance

**Screen:** Configure → **Monitoring** · **Routes:** `/monitor/sla/*`, `/monitor/schedule/*`, `/monitor/channel/*`, `/monitor/maintenance/*`, `GET /monitor/alert/list`, `GET /monitor/slareport`, `GET /monitor/slareport.pdf`, `GET /monitor/trend` · **Capabilities:** `sla_view`, `sla_manage`, `schedule_manage`, `report_view`, `report_export`

- 🌟 **Commercial Presentation & Sales Pitch**: Targets per site for latency, jitter, loss and MOS, graded **OK / Degraded / Breached** — Degraded warns at 80% of a threshold, before the customer notices. Alerts that do not cry wolf: a breach must persist for the sweeps you choose, then is announced once, its recovery once, and an outage that follows a warning is escalated even inside the quiet period. Planned work is excluded from the figure a customer is measured against, and the monthly SLA report — availability, loss, jitter and MOS, IPv4 and IPv6 separately — exports as the document for the customer.
- 📖 **User Guide & Operational Flow**:
  - **Service targets:** maximum latency, jitter and loss and a minimum MOS; mark one **default** so new sites are measured against it.
  - **Schedules:** every *N* minutes (five or more), over **every enabled site** (including ones added later) or named sites. **Run now** runs one at once.
  - **Alert channels:** an email address or an HTTPS webhook. **Send a test** proves it works now.
  - **Alert history:** what fired, when, and whether it was delivered.
  - **Maintenance windows:** the whole organisation, a region or named sites, in the window's own timezone. Results inside are *excluded*: no alert, no effect on availability.
  - **SLA report:** a month per site; a month with no measurements has no figure — never 100%. **Export PDF**.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `MON-001` | Administrator | Grading against a target | Target: latency 100 ms; a site averaging 85 ms; another at 120 ms | 85 ms → *Degraded* (past 80% of 100); 120 ms → *Breached*; one under 80 ms → *OK* |
  | `MON-002` | Administrator | The default target applies to new sites | Mark a target default; add a site with no target; run | The new site is graded against the default |
  | `MON-003` | Administrator | Schedules run, and run now | Schedule every 5 minutes; **Run now** | A diagnostic `<prefix>-<UTC date-time>` appears in History at once, and again every 5 minutes. An interval under 5 is refused |
  | `MON-004` | Administrator | A breach is announced once, its recovery once | Webhook channel; a site that fails three sweeps then recovers, damping set to 2 consecutive breaches | One alert after the 2nd failing sweep, none on the 3rd, one recovery alert. 🛑 **Must NOT** alert on every failing sweep |
  | `MON-005` | Administrator | An outage after a warning is escalated | Quiet period 60 min; a site goes *Degraded* (alert), then *Breached* 5 minutes later | A second, critical alert despite the quiet period |
  | `MON-006` | Administrator | A webhook cannot reach the host | Add a channel `https://127.0.0.1/hook`, then `http://example.com/hook` | Both refused: loopback is not a destination, and a webhook must be HTTPS |
  | `MON-007` | Administrator | Maintenance excludes and silences | A window over a failing site, in `Asia/Kolkata`, viewed from a device in another timezone | Results inside are *excluded*: no alert, availability unaffected; the window's hours are Kolkata wall-clock hours |
  | `MON-008` | Engineer | The SLA report says what it measured | Open a month with data and one without | With data: availability, loss, jitter and MOS per site, IPv6 lines marked `(v6)`; without: no figure. 🛑 **Must NOT** show 100% for an unmeasured month |
  | `MON-009` | Engineer | The report exports | **Export PDF** on the SLA report | The month's document, with the Loss column and the `(v6*)` footnote where report-only lines appear |
  | `MON-010` | Tester | The guard suite | `./packetpulsetest.sh monitor` | Passes — grading, damping, recovery and maintenance against a live server |
- ⚙️ **Developer Guide & Release Confidence**:
  - Damping: `DecideAlert` in `packetpulsego/pkg/monitormicroservice/monitordomain/shared/` (consecutive breaches, cooldown, escalation); runner: `packetpulsego/pkg/monitormicroservice/monitorservice/MonitorScheduleRunner.go` (claims due schedules with `FOR UPDATE SKIP LOCKED`).
  - Results are partitioned by month and rolled up daily per site **and family** (`packetpulsego/pkg/common/dbclient/migrations/0012_2026_10_01_result_partitioning_and_retention.sql`, `0023`); the retention runner keeps partitions three months ahead.
  - 🔒 Report-only IPv6 never enters availability (`NOT report_only` in the rollup and the SLA queries).
  - Coverage: `packetpulsetest/golang/monitorconformance/`, `packetpulsego/pkg/monitormicroservice/**`, `packetpulseflutter/test/monitor_screen_test.dart`, `packetpulseflutter/test/monitor_flows_test.dart`.

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

**Screen:** Configure → **Result export** · **Routes:** `GET/PUT /export/target`, `POST /export/target/test` · **Capability:** `apikey_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Results delivered to the operator's own SFTP or FTPS server as CSV, every interval it chooses — no integration project. Nothing is lost: a failed delivery is retried with the same period, and each row carries a `result_id` so a file that arrives twice is recognised. The server's key is confirmed before anything is sent, and a saved password is only ever sent to the server it was entered for.
- 📖 **User Guide & Operational Flow**:
  - **Protocol:** SFTP (recommended), FTP over TLS, or FTP — with a clear warning that plain FTP is unencrypted.
  - **Server, Port, Username, Password** (or for SFTP a private key). Saved credentials are never shown again; leave the field empty to keep them, or tick *Remove the saved password*.
  - **Test connection** signs in, writes and removes a small file. For SFTP it shows the server's key: check it with whoever runs the server, then **Trust this key** and **Save**.
  - **Folder** and **Deliver every (minutes)** — 15 to 1440.
  - The status card says how the last delivery went: *Sent <file> with N results*, *Nothing new to send*, or *Not delivered: <why>*.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `EXP-001` | Administrator | A target nobody set up | Open the screen | *Not delivered yet. The first delivery is one interval after saving.*; SFTP, port 22, 60 minutes; switched off |
  | `EXP-002` | Administrator | Confirm the server's key | Fill in an SFTP server; **Test connection** | *Trust this server?* showing `SHA256:…`. **Cancel** → *Not trusted yet*; test again → **Trust this key** → *Trusted: SHA256:…*; **Save** |
  | `EXP-003` | Administrator | Switching on needs everything a delivery needs | Switch on with no server, no username, no password and no trusted key | Refused against each field |
  | `EXP-004` | Administrator | A changed key is a warning, not a detail | Re-key the SFTP server (or point the name at another); **Test connection** | *The server's key has changed* with the new key to check; the trusted key stays until you trust the new one. 🛑 **Must NOT** deliver to a server showing a different key |
  | `EXP-005` | Administrator | A saved password goes nowhere new | With a saved password, change the server (or port, username or protocol) and test or save | The password field becomes required again (*Enter the password again…*). 🛑 **Must NOT** sign in to the new server with the saved password |
  | `EXP-006` | Administrator | A delivery lands whole | Switched on, every 15 minutes, a diagnostic run after saving; wait for the interval | `packetpulse-results-<from>-<to>.csv` in the folder (written as `.part`, renamed when whole); the card says *Sent … with N results* |
  | `EXP-007` | Administrator | A quiet period sends nothing | No diagnostics in an interval | No file; *Nothing new to send* |
  | `EXP-008` | Administrator | A failure is retried, nothing lost | Change the server's password; wait for a delivery; restore it; wait again | First: *Not delivered: The server refused the sign-in…*; the next delivery covers the same results. 🛑 **Must NOT** skip the failed period |
  | `EXP-009` | Administrator | PacketPulse's own host is not a destination | Server `127.0.0.1`, `localhost`, `169.254.169.254` or `::1`; **Test connection** | Refused against **Server**: *PacketPulse may not connect to that address.* |
  | `EXP-010` | Administrator | Plain FTP is warned about | Choose FTP | The unencrypted warning; port moves to 21; the private key and server key disappear |
  | `EXP-011` | Administrator | FTPS with a private authority | FTPS to a server whose certificate your own CA signed, without and then with that CA | Without: *certificate is not trusted*; with it pasted: test passes |
  | `EXP-012` | Tester | Credentials never come back | `GET /api/v1/export/target` | `has_password`, `has_private_key` — never the password, key or sealed text |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/exportmicroservice/` (target, test, runner each minute); transport: `packetpulsego/pkg/common/filedrop/FileDrop.go` (SFTP via `pkg/sftp`, FTP/FTPS via `jlaffaye/ftp`).
  - 🔒 Every connection — the FTP **data** connection included — is made by `probeguard.Dialer`, whose `Control` hook checks the resolved address before connecting. A dial *function* would have made the FTP library send FTPS data unencrypted.
  - The watermark (`exported_through`) moves only on a delivery, and only from where the run found it; a run catches up a day per file; the last two minutes are left to settle.
  - `EXPORT_ALLOW_LOOPBACK` permits a server on PacketPulse's own host for development; production refuses to boot with it.
  - Coverage: `packetpulsego/pkg/common/filedrop/` (in-process SFTP and FTP/FTPS servers), `packetpulsego/pkg/exportmicroservice/**`, `packetpulseflutter/test/result_export_screen_test.dart`, `packetpulsetest/golang/apicontract/result_export_test.go`.

---

### 6.3 🏢 Directory

**Screen:** Configure → **Directory** · **Routes:** `GET/PUT /ldap/config`, `POST /ldap/config/test`, `/ldap/groupmap/*` · **Capability:** `ldap_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Staff sign in with their existing network password over LDAP or Active Directory, and their directory group decides their PacketPulse role — after the second step, so a directory password alone still opens nothing. The owner always keeps a local password, so a directory outage cannot lock out the person who fixes it.
- 📖 **User Guide & Operational Flow**: Host and port (389 LDAP/StartTLS, 636 LDAPS), encryption, the CA certificate of your own authority, a read-only bind account, the base DN; then **Test connection** before enabling. **Group mappings** give each directory group a role.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `LDAP-001` | Administrator | Connect and test | Point at a test directory with its CA certificate; **Test connection** | Success; a wrong CA → certificate error; skipping verification is warned against |
  | `LDAP-002` | Directory user | The group decides the role — after the second step | Map group `noc` → *NOC Engineer*; a directory user in `noc` signs in | Password checked by the directory; second step asked; only then the role applied |
  | `LDAP-003` | Owner | The owner keeps a local password | Directory switched on and unreachable; owner signs in with their local password | Signed in. Others: *Your organisation's directory server could not be reached. Please retry.* |
  | `LDAP-004` | Tester | An outage is not a failed sign-in | Directory down; an administrator signs in | No *Sign-in failed* entry in Activity — nobody got the password wrong |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/`; the bind password is sealed like every secret; the directory is dialled through the directory policy (private yes; loopback only with `LDAP_ALLOW_LOOPBACK`, refused in production).
  - Coverage: `packetpulsego/pkg/common/ldaptest/` (an in-process directory), `packetpulseflutter/test/ldap_settings_screen_test.dart`.

---
## Group 7 — Platform and Settings

The operator of PacketPulse itself, each person's own settings, and the public face.

---

### 7.1 🏛️ Platform console and licences

**Screen:** Administer → **Platform** (superuser only) · **Routes:** `/platform/*` · **Access:** platform superuser

- 🌟 **Commercial Presentation & Sales Pitch**: Every customer runs on a licence PacketPulse signs, so nobody can grant themselves an organisation, more people or a longer term. Licences carry the people covered, sites, a period and a price, scale in proportion when sold for an unusual term, and keep a history of every change with who made it and why. A customer's server needs nothing from PacketPulse to check one: the signature is checked against a key built into the server.
- 📖 **User Guide & Operational Flow**: The console is PacketPulse's own server; a customer's server has none. **Organisations → Add**; **Issue licence** with plan, people, sites, months and currency. Per licence: **Suspend**, **Resume**, **Revoke**, **Renew**, change the people and sites it covers. **Download licence file** asks for the **Owner's email** and saves the signed file; copy it into `LICENCE_DIR` on the customer's server. It is read at every sign-in, so no restart is needed. The first person to sign up there with that address, proved by an emailed code, becomes the Administrator.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `PLAT-001` | Superuser | Onboard a customer | Create `ACME`; issue a 12-month licence for 10 people; **Download licence file** naming `owner@acme.example`; copy it into a test server's `LICENCE_DIR`; sign up there with that address | The emailed code completes the sign-up, and the owner lands in ACME with every Administrator capability |
  | `PLAT-002` | Superuser | Suspend and resume | Suspend ACME's licence with a reason; download its file again; resume | While suspended: running a diagnostic is refused, reading is not. History shows both, with the reason |
  | `PLAT-003` | Superuser | A smaller licence keeps everyone but the owner out | ACME has 4 people; change the licence to 3; replace the file; an engineer signs in, then the owner | The engineer is refused (403, *Your organisation has more users than its licence covers…*). The owner signs in, because the owner is who can fix it |
  | `PLAT-004` | Superuser | Renewal scales the price | Renew a 12-month licence for 36 months | Three times the period price; the new end runs from the current end (or today if lapsed) |
  | `PLAT-005` | Superuser | A superuser is not a tenant | Call `GET /dnssite/list` as the superuser | 400 — no organisation. 🛑 **Must NOT** answer with every organisation's sites |
  | `PLAT-006` | Stranger | Signing up never makes or joins an organisation | On a customer's server, sign up with an address its licence does not name | Refused: *Only the owner named in this server's licence can sign up. Ask your administrator to add you.* 🛑 **Must NOT** create an account |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/platformmicroservice/`; `packetpulseaccess.RequireSuperUser` on the console group (Appendix A); licence changes recorded in the licence history and the activity trail.
  - 🔒 Licence files are Ed25519-signed and verified by `packetpulsego/pkg/common/licence/Store.go` at every sign-in, against a key compiled into the server. `LICENCE_PUBLIC_KEY` may replace it only on a development build outside production, so a customer cannot swap in a key of their own.
  - Coverage: `packetpulsetest/golang/tenancyassignment/`, `packetpulsego/pkg/platformmicroservice/**`, `packetpulseflutter/test/platform_console_screen_test.dart`.

---

### 7.2 ⚙️ Settings

**Screen:** Administer → **Settings** · **Routes:** `POST /user/appearance`, `POST /user/password`

- 🌟 **Commercial Presentation & Sales Pitch**: Five themes in light and dark — including a warm, low-blue-light theme for night shifts and a high-contrast one — every one contrast-checked, with status colours that never change meaning. They follow the person, not the device.
- 📖 **User Guide & Operational Flow**: Light, dark or follow the device; theme; density; reduce motion; change password. PacketPulse is in English; there is no language choice. The same light / dark and appearance controls are on the sign-in screen, where they are remembered on that device until someone signs in.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SET-001` | Engineer | Appearance follows the person | Choose a dark theme on a laptop; sign in on another browser | The same theme there |
  | `SET-002` | Engineer | Status colours keep their meaning | Switch through every theme on a result with OK, Degraded and Breached rows | Green, amber and red in every theme, each with its icon |
  | `SET-003` | Engineer | English only | Look for a language picker in the bar and in Settings | None. Every label is English, served by the server. 🛑 **Must NOT** show a key such as `s812` in place of a label |
  | `SET-004` | Engineer | Change password | Change it; sign out; sign in with the old, then the new | Old refused, new accepted (then the second step) |
  | `SET-005` | Visitor | Appearance before signing in | Signed out, use the sun/moon button and the palette on the sign-in screen (desk and phone width); reload; then sign in to an account set to another mode, and sign out | Each choice applies at once and survives the reload, with no request to `/user/appearance`. After sign-in the **account's** appearance is worn; after sign-out the sign-in screen keeps it. 🛑 **Must NOT** snap back to dark on sign-out, or save a signed-out choice to anyone's account |
- ⚙️ **Developer Guide & Release Confidence**:
  - Strings: `scripts/intellicodegen/packetpulsestrings.py` generates the Go catalogue and the Dart index; entries are positional, so retired ones stay in `DEPRECATED`. Help: `packetpulsego/pkg/initmicroservice/initconstants/PacketPulseHelp.go`.
  - Coverage: `packetpulseflutter/test/settings_screen_test.dart`, the `translation` suite. Signed-out appearance: `packetpulseflutter/lib/common/services/PacketPulseAppearanceStore.dart`, tested in `packetpulseflutter/test/app_test.dart` and `packetpulseflutter/test/sign_in_test.dart`.

---

### 7.3 🌐 Public site and self-test

**Pages:** `https://packetpulse.rummaan53.com/` and `/selftest.html` (`packetpulseweb/`)

- 🌟 **Commercial Presentation & Sales Pitch**: The product site says only what the product does today, and the public self-test lets a prospect measure their own connection in the browser before talking to anyone.
- 📖 **User Guide & Operational Flow**: **Sign in** opens the app at `/app/`; **Test my connection** opens the self-test. The sun, moon and screen buttons in the header choose light, dark or match this device; the app's sign-in screen opens in the same choice. On a phone the header keeps the theme switch and **Sign in**, and **Test my connection** is in the footer.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `WEB-001` | Visitor | The site is current | Read the capabilities | Emailed or texted sign-in codes, IPv4 and IPv6, CSV into your systems, check-ins, testing from the engineer's own device. 🛑 **Must NOT** promise 23 languages, a hop-by-hop path or a line-speed test |
  | `WEB-002` | Visitor | Site and app are different documents | Open `/` and `/app/` | The marketing page and the app respectively — never the same document |
  | `WEB-003` | Visitor | The self-test runs in the browser | **Test my connection** → run | Round trips, jitter and loss for each resolver, measured from the visitor's connection |
  | `WEB-004` | Visitor | Light, dark or match this device | Choose **Light** on `/`; reload; open `/selftest.html`; then **Sign in**. Choose **Match this device** and switch the computer's own mode | Light at once, on both pages and on the app's sign-in screen; *Match this device* follows the computer as it changes. 🛑 **Must NOT** flash dark before a light page paints |
  | `WEB-005` | Visitor | The header fits a phone | Open `/` at 360 px wide | Logo, theme switch and **Sign in** on one line, no sideways scrolling |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulseweb/index.html`, `packetpulseweb/selftest.html`; the probe engine's tests: `packetpulseweb/assets/packetpulse-probe.test.mjs` (part of `unit`). The theme switch: `packetpulseweb/assets/packetpulse-theme.js`, which keeps the mode under the app's own appearance key; its tests, `packetpulseweb/assets/packetpulse-theme.test.mjs`, are part of `unit` too.
  - The deploy asserts `/` and `/app/` differ, so the nginx misroute that once served a blank app is caught by the deploy.

---
## Feature → Coverage Matrix

What guards each area automatically, so a manual pass can spend its time where automation cannot reach: real devices, real browsers, real file servers, and judgement.

| Area | Manual IDs | Go unit and repository | Integration suites | Flutter |
|---|---|---|---|---|
| Sign-in, second step | `AUTH-*` | `packetpulsego/pkg/usermicroservice/userservice/` | `assignment`, `audit` | `packetpulseflutter/test/sign_in_test.dart` |
| SMS gateway, location rule | `SMS-*`, `LOC-*` | `packetpulsego/pkg/smsmicroservice/`, `packetpulsego/pkg/common/smsprovider/` | `acl` | `packetpulseflutter/test/sign_in_security_screen_test.dart` |
| Staff, roles | `STF-*`, `ACL-*` | `packetpulsego/pkg/staffmicroservice/` | `acl`, `tenancy` | `packetpulseflutter/test/staff_screen_test.dart` |
| Sessions and licence | `WHO-*` | `packetpulsego/pkg/staffmicroservice/` | `assignment`, `tenancy` | `packetpulseflutter/test/session_screen_test.dart` |
| Check-ins | `CHK-*` | `packetpulsego/pkg/common/geocode/` | `tenancy`, `audit` | `packetpulseflutter/test/checkin_screen_test.dart` |
| Activity | `AUD-*` | `packetpulsego/pkg/common/auditlog/` | `audit` | `packetpulseflutter/test/audit_log_screen_test.dart` |
| Sites, policy | `SITE-*` | `packetpulsego/pkg/dnssitemicroservice/`, `packetpulsego/pkg/common/probeguard/` | `tenancy` | `packetpulseflutter/test/dns_site_screen_test.dart` |
| Diagnostics, IPv6 | `DIAG-*`, `V6-*` | `packetpulsego/pkg/pingmicroservice/`, `packetpulsego/pkg/diagnosticmicroservice/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_submit_screen_test.dart` |
| CSV | `CSV-*`, `API-003` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_history_screen_test.dart` |
| Run diagnostic, device test | `DEV-*` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract` | `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart` |
| Monitoring | `MON-*` | `packetpulsego/pkg/monitormicroservice/` | `monitor` | `packetpulseflutter/test/monitor_flows_test.dart` |
| Results API | `API-*` | `packetpulsego/pkg/common/apikeyauth/` | `contract`, `tenancy` | `packetpulseflutter/test/api_key_screen_test.dart` |
| Result export | `EXP-*` | `packetpulsego/pkg/common/filedrop/`, `packetpulsego/pkg/exportmicroservice/` | `contract`, `acl`, `tenancy`, `audit` | `packetpulseflutter/test/result_export_screen_test.dart` |
| Directory | `LDAP-*` | `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/` | — | `packetpulseflutter/test/ldap_settings_screen_test.dart` |
| Platform | `PLAT-*` | `packetpulsego/pkg/platformmicroservice/` | `assignment` | `packetpulseflutter/test/platform_console_screen_test.dart` |
| Settings, catalogue | `SET-*` | `packetpulsego/pkg/initmicroservice/` | `translation` | `packetpulseflutter/test/settings_screen_test.dart` |

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
2. Run against the customer's sites. Read loss, jitter and MOS on each row.
3. Re-run after the carrier's fix.
4. In **History**, compare both attempts.

**Expected:** both attempts kept; the second's average loss and jitter lower; the dual-stack site shows its IPv6 row *not counted*.

---

### JRN-003 · A field engineer's day

**Covers:** `STF-004`, `LOC-002`, `CHK-001`–`CHK-002`, `DEV-001`–`DEV-006`

1. The administrator sets the engineer to **Record location** and turns on **Require location to sign in**.
2. On a phone, the engineer signs in, allowing location.
3. At the customer's site: **Run diagnostic → From this device**, then attach the run to `TT-JRN-003`.
4. Signs out, allowing location.

**Expected:** **Check-ins** shows the sign-in and sign-out with places and a map link; the ticket shows the device run, *measured on a device*; the administrator's Activity shows no entry for the engineer's sign-in.

---

### JRN-004 · The IT system takes the results

**Covers:** `API-001`–`API-003`, `EXP-002`, `EXP-006`–`EXP-008`

1. The administrator issues a Results API key; the IT system pulls yesterday's CSV.
2. The administrator sets up **Result export** to the IT team's SFTP server: test, trust the key, save, every 15 minutes.
3. Engineers run diagnostics for an hour.
4. The SFTP password is changed on the server for one interval, then restored.

**Expected:** a file per interval with results; the interval with the wrong password shows *Not delivered*, and the next file holds that period's results too; every `result_id` appears exactly once across all the files — a failed delivery sent nothing, so nothing is sent twice.

---

### JRN-005 · Monitoring catches an outage and its recovery

**Covers:** `MON-001`–`MON-005`, `MON-007`

1. A default target; a schedule every 5 minutes over every enabled site; a webhook channel; damping 2.
2. Black-hole one site for 15 minutes, then restore it.
3. Declare a maintenance window over another site and break it inside the window.

**Expected:** one breach alert, one recovery; the maintenance site raises nothing and its availability is untouched.

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
| `NFR-010` | Reduce motion on (app setting or the device's) | No transitions |
| `NFR-011` | Status by colour alone | Every pill carries an icon as well as a colour |
| `NFR-012` | Every user-visible sentence | From the catalogue: no English hard-coded in a widget, no raw key shown |

## 4.4 Security

| ID | Check | Expected |
|---|---|---|
| `NFR-013` | Tenancy, for every id-taking route | RIVAL's id answers 404 or an empty list — `./packetpulsetest.sh tenancy` |
| `NFR-014` | The destination policy, everywhere PacketPulse connects out | Sweeps, webhooks, the directory and the result export all refuse loopback, link-local and cloud metadata |
| `NFR-015` | Error responses | No SQLSTATE, driver names, file paths or a remote server's raw reply in any error envelope |
| `NFR-016` | Secrets at rest | SMS tokens, bind passwords and export credentials sealed with the server's secret box; API keys stored as hashes; never returned |
| `NFR-017` | CSV injection | Cells starting `=`, `+`, `-`, `@`, tab or carriage return are written as text |
| `NFR-018` | Production configuration | Production refuses `OTP_OUTBOX_FILE`, a `LICENCE_PUBLIC_KEY` override, `LDAP_ALLOW_LOOPBACK`, `EXPORT_ALLOW_LOOPBACK` and open CORS, and will not boot without an SMTP account |

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
| `DEMO_SMS_CODE`, `DEMO_SEATS` | `111111`, `3` | The demo's on-screen code, and how many demo walkthroughs may run at once |
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
