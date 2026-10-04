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
  | `MON-010` | Tester | The guard suite | `./pingletest.sh monitor` | Passes — grading, damping, recovery and maintenance against a live server |
- ⚙️ **Developer Guide & Release Confidence**:
  - Damping: `DecideAlert` in `pinglego/pkg/monitormicroservice/monitordomain/shared/` (consecutive breaches, cooldown, escalation); runner: `pinglego/pkg/monitormicroservice/monitorservice/MonitorScheduleRunner.go` (claims due schedules with `FOR UPDATE SKIP LOCKED`).
  - Results are partitioned by month and rolled up daily per site **and family** (`pinglego/pkg/common/dbclient/migrations/0012_2026_10_01_result_partitioning_and_retention.sql`, `0023`); the retention runner keeps partitions three months ahead.
  - 🔒 Report-only IPv6 never enters availability (`NOT report_only` in the rollup and the SLA queries).
  - Coverage: `pingletest/golang/monitorconformance/`, `pinglego/pkg/monitormicroservice/**`, `pingleflutter/test/monitor_screen_test.dart`, `pingleflutter/test/monitor_flows_test.dart`.

---
