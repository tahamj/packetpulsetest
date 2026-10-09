<div align="center">

# 📡 PacketPulse Manual Testing Guide

### *& Master Knowledge Manual*

**The single source of truth for testing, understanding, selling, operating, and releasing PacketPulse.**

<br>

`v2026.10-PROD-v2`  ·  `Verified against source 2026-10-06`

<br>

| 🧭 Screens | 🔌 API routes | 🔐 Capabilities | 🗄️ Migrations | 💬 Catalogue strings | ❓ Help topics | 🧪 Guard suites |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **17** <br><sub>in the navigation rail</sub> | **111** <br><sub>under `/api/v1`</sub> | **17** <br><sub>3 built-in roles</sub> | **31** <br><sub>applied at boot</sub> | **1014** <br><sub>English, server-served</sub> | **19** <br><sub>one per screen</sub> | **11** <br><sub>+ `load`, opt-in</sub> |

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
| **One Run diagnostic page** | *Run diagnostic* and *Test from this device* are one page: the sweep **From the server** on top, the test **From this device** below, and one Customer ID, TT number and note for both. There is no mode switch (since 2026-10-06; v2 shipped two modes). Superseded on 2026-10-07: see the next row. | `DEV-*` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/screens/RunDiagnosticScreen.dart` |
| **Device-test target kept internal** | The device test no longer names what it measures and offers no **Change target**; it still measures the organisation's host, or Cloudflare DNS when none is chosen. A filed device run reads *Device test*, with no address, in History, its PDF and every CSV; the stored row keeps the host. The target is set through `PUT /organisation/settings` (since 2026-10-06). | `DEV-001`, `DEV-002`, `DEV-006`, `DEV-007` | `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticdomain/shared/DeviceTarget.go` |
| **Run diagnostic from this device, to an endpoint** | The sweep **From the server** left the page; **Run diagnostic** is the test from this device. Customer ID, TT number and **Endpoint** are required before **Start test**; the endpoint nearest the device is chosen from each endpoint's latitude and longitude, an engineer cannot change it, and a manager can. The run measures that endpoint and is filed, and named, under it; the server refuses another organisation's (since 2026-10-07). | `DEV-017`–`DEV-022`, `DEV-006` | `packetpulseflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticEndpointField.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go` |
| **UDP test in Run diagnostic** | In a browser, **Run diagnostic** measures the ticket's endpoint with **UDP, like a call** by default (or **HTTPS to the endpoint**), for 10, 30 or 60 seconds, and files where each lost packet went - up, down or late - with jitter each way, on the ticket, its page and its PDF. It measures to the reflector beside the endpoint when the endpoint names one (**UDP reflector address**, §3.1), and otherwise to the PacketPulse network, filed as *PacketPulse network*: the customer's access line, never the endpoint. The desktop and phone apps keep ICMP (since 2026-10-09). | `UDP-016`–`UDP-026`, `SITE-011`–`SITE-013` | `packetpulseflutter/lib/diagnosticmicroservice/service/udp/UdpProbeRun.dart`<br>`packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticUdpObservation.go` |
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
| **5. Loss and jitter** | Loss and RFC 3550 jitter on every result, ticket, history row, dashboard row, PDF and SLA report. | `DIAG-005`, `HIST-*`, `MON-*` | `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go`<br>`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql` |
| **6. CSV, FTP and download** | **Export CSV** on every ticket; a Results API bulk pull; a scheduled push to the organisation's own **SFTP / FTPS / FTP** server with the SSH host key confirmed first. | `CSV-*`, `EXP-*` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`<br>`packetpulsego/pkg/exportmicroservice/exportservice/ExportService.go` |
| **7. English only** | The other 22 languages were removed; wording is still served by the server, in English. | `SET-003` | `scripts/intellicodegen/packetpulsestrings.py` |
| **8. Load testing** | A platform load suite: 20 people at once, a 200-site sweep, API and CSV pulls against p95 budgets. *The device line-speed test of v1 is hidden in v2.* | Part IV §4.5 | `packetpulsetest/golang/loadtest/load_test.go`<br>`.github/workflows/packetpulse-load.yml` |
| **Admin logins tracked** | Administrators' sign-ins and sign-outs in the **Activity** trail with device, browser, address and place; **failed** attempts on their accounts too, with the reason. | `AUD-*` | `packetpulsego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go` |

---
