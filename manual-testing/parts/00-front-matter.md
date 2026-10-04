<div align="center">

# 📡 Pingle Manual Testing Guide

### *& Master Knowledge Manual*

**The single source of truth for testing, understanding, selling, operating, and releasing Pingle.**

<br>

`v2026.10-PROD-v1`  ·  `Verified against source 2026-10-04`

<br>

| 🧭 Screens | 🔌 API routes | 🔐 Capabilities | 🗄️ Migrations | 💬 Catalogue strings | ❓ Help topics | 🧪 Guard suites |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **17** <br><sub>in the navigation rail</sub> | **110** <br><sub>under `/api/v1`</sub> | **17** <br><sub>3 built-in roles</sub> | **24** <br><sub>applied at boot</sub> | **835** <br><sub>English, server-served</sub> | **18** <br><sub>one per screen</sub> | **11** <br><sub>+ `load`, opt-in</sub> |

<br>

*Quad-Lens Architecture — an **executable QA playbook**, an **onboarding & user manual**, a **commercial & sales pitch**, and a **developer release confidence guide**.*

<sub>🔒 Every number above is asserted against source by **`./pingletest.sh docs`**. This guide is **generated** from `pingletest/manual-testing/parts/` by `pingletest/manual-testing/build/build_guide.py` — edit the parts, not the outputs.</sub>

</div>

---

## 🧭 How To Read This Manual

Pingle is a multi-tenant network-diagnostics service for telecom operators: a Go API, a Flutter client (web, desktop, phone) and PostgreSQL, live at **https://pingle.rummaan53.com**. It is structurally a server-side request forgery primitive with a user interface — an authenticated person names an address and the server sends traffic to it — so much of what follows is about what must **not** happen.

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
> Every manual test case carries a stable **Test ID** (`AUTH-004`, `JRN-003`, …). Quote it in bug reports and run records. IDs are **never reused** — a retired case leaves its number vacant — and `./pingletest.sh docs` fails if two cases share one.

> [!CAUTION]
> **Never test against live customer data.** Use a local server with a disposable database (Part 0 §0.2), or the live demo organisation for read-only walkthroughs. Tenancy, ACL and audit cases need **two organisations** of your own — testing them from a single account proves nothing.

---

## 📋 What Changed in v1 — the October 2026 Customer Release

> [!IMPORTANT]
> **Deployed to https://pingle.rummaan53.com on 2026-10-04.** It answers eight requirements from Northwind Telecom plus their remark that administrator logins be tracked. Every row below must be verified before a release is signed off.

| Requirement | What changed | Where to test | Source of truth |
|---|---|---|---|
| **1. Show IPv6** | Client IPs are read only from trusted proxies and shown canonically; a dual-stack site is measured over **both** families, the IPv6 result **reported, not counted**; the device test shows its IPv4 and IPv6 egress. | `V6-*`, `DEV-004` | `pinglego/pkg/common/probeguard/ProbeGuardPolicy.go`<br>`pinglego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql` |
| **2. Two-step sign-in** | Every sign-in has a second step: an authenticator app (TOTP) or a texted code through the organisation's **own Twilio account**. Owners and superusers always use an authenticator; recovery codes for a lost phone. | `AUTH-*`, `SMS-*` | `pinglego/pkg/usermicroservice/userservice/UserSignInSteps.go`<br>`pinglego/pkg/common/dbclient/migrations/0019_2026_10_03_two_step_sign_in.sql` |
| **3. No target dropdown** | *Test from this device* tests the target the organisation chose (Cloudflare DNS by default), with one-tap presets for whoever may change it. | `DEV-*` | `pingleflutter/lib/diagnosticmicroservice/presentation/screens/ClientProbeScreen.dart` |
| **4. Location at check-in** | Sign-in and sign-out record the device's position for people set to *Record location*; field engineers on a separate **Check-ins** screen with address and map link; a per-organisation *Require location* rule. | `CHK-*`, `LOC-*` | `pinglego/pkg/common/dbclient/migrations/0020_2026_10_03_session_checkin.sql`<br>`pinglego/pkg/common/geocode/Geocode.go` |
| **5. Loss and jitter** | Loss and RFC 3550 jitter on every result, ticket, history row, dashboard row, PDF and SLA report. | `DIAG-005`, `HIST-*`, `MON-*` | `pinglego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go`<br>`pinglego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql` |
| **6. CSV, FTP and download** | **Export CSV** on every ticket; a Results API bulk pull; a scheduled push to the organisation's own **SFTP / FTPS / FTP** server with the SSH host key confirmed first. | `CSV-*`, `EXP-*` | `pinglego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`<br>`pinglego/pkg/exportmicroservice/exportservice/ExportService.go` |
| **7. English only** | The other 22 languages were removed; wording is still served by the server, in English. | `SET-003` | `scripts/intellicodegen/pinglestrings.py` |
| **8. Load testing** | A device line-speed test (Cloudflare), and a platform load suite: 20 people at once, a traced 200-site sweep, API and CSV pulls against p95 budgets. | `DEV-005`, Part IV §4.5 | `pingletest/golang/loadtest/load_test.go`<br>`.github/workflows/pingle-load.yml` |
| **Admin logins tracked** | Administrators' sign-ins and sign-outs in the **Activity** trail with device, browser, address and place; **failed** attempts on their accounts too, with the reason. | `AUD-*` | `pinglego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go` |

---
