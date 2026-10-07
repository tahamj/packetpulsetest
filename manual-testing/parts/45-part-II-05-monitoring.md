## Group 5 — Monitoring

Turning on-demand testing into a continuous watch, with alerts that mean something.

---

### 5.1 📈 The SLA report, and the monitoring behind it

**Screen:** Configure → **Monitoring** (the SLA report) · **Routes:** `GET /monitor/slareport`, `GET /monitor/slareport.pdf`, `GET /monitor/trend`; through the API only: `/monitor/sla/*`, `/monitor/schedule/*`, `/monitor/channel/*`, `/monitor/maintenance/*`, `GET /monitor/alert/list` · **Capabilities:** `report_view`, `report_export`; through the API: `sla_view`, `sla_manage`, `schedule_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: The monthly SLA report — availability, loss, jitter and MOS per site, IPv4 and IPv6 separately — exports as the document for the customer. Behind it, every result is graded against targets for latency, jitter, loss and MOS: **OK / Degraded / Breached**, with Degraded warning at 80% of a threshold, before the customer notices. Alerts do not cry wolf: a breach must persist for the sweeps chosen, then is announced once, its recovery once, and an outage that follows a warning is escalated even inside the quiet period. Planned work is excluded from the figure a customer is measured against.
- 📖 **User Guide & Operational Flow**:
  - **Monitoring** is the SLA report, for anyone who may read reports (`report_view`): last month first, or any of the eleven before it under **Month**.
  - **Availability**, **Sites on target** and **Excluded by maintenance** head the month; each site has a line with round-trip time, loss, jitter and MOS, and *On target* or *Missed target*. A month with no measurements has no figure — never 100%.
  - **Export PDF**, for those who may export reports (`report_export`).
  - Service targets, schedules, alert channels, alert history and planned work are no longer on the screen (removed October 2026). The server still grades, runs schedules, sends alerts and excludes planned work, for what was set up before and for what is set up through the API routes above.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `MON-001` | Administrator | Grading against a target | Through the API, a target: latency 100 ms; a site averaging 85 ms; another at 120 ms | 85 ms → *Degraded* (past 80% of 100); 120 ms → *Breached*; one under 80 ms → *OK* |
  | `MON-002` | Administrator | The default target applies to new sites | Through the API, mark a target default; add a site with no target; run | The new site is graded against the default |
  | `MON-003` | Administrator | Schedules run, and run now | Through the API, a schedule every 5 minutes; `POST /monitor/schedule/{scheduleId}/run` | A diagnostic `<prefix>-<UTC date-time>` appears in History at once, and again every 5 minutes. An interval under 5 is refused |
  | `MON-004` | Administrator | A breach is announced once, its recovery once | Through the API, a webhook channel; a site that fails three sweeps then recovers, damping set to 2 consecutive breaches | One alert after the 2nd failing sweep, none on the 3rd, one recovery alert. 🛑 **Must NOT** alert on every failing sweep |
  | `MON-005` | Administrator | An outage after a warning is escalated | Quiet period 60 min; a site goes *Degraded* (alert), then *Breached* 5 minutes later | A second, critical alert despite the quiet period |
  | `MON-006` | Administrator | A webhook cannot reach the host | `POST /monitor/channel/add` with `https://127.0.0.1/hook`, then `http://example.com/hook` | Both refused: loopback is not a destination, and a webhook must be HTTPS |
  | `MON-007` | Administrator | Maintenance excludes and silences | Through the API, a window over a failing site, in `Asia/Kolkata` | Results inside are *excluded*: no alert, availability unaffected; the window's hours are Kolkata wall-clock hours |
  | `MON-008` | Engineer | The SLA report says what it measured | Open a month with data and one without | With data: availability, loss, jitter and MOS per site, IPv6 lines marked `IPv6 · not counted`; without: no figure. 🛑 **Must NOT** show 100% for an unmeasured month |
  | `MON-009` | Engineer | The report exports | **Export PDF** on Monitoring | The month's document, with the Loss column and the `(v6*)` footnote where report-only lines appear |
  | `MON-010` | Tester | The guard suite | `./packetpulsetest.sh monitor` | Passes — grading, damping, recovery and maintenance against a live server |
  | `MON-011` | Administrator | Monitoring is the report alone | Open **Monitoring** | The month's report, and nothing else. 🛑 **Must NOT** offer *Service targets*, *Schedules*, *Alert channels*, *Alert history* or *Planned work* |
  | `MON-012` | Administrator | Monitoring follows the right to read reports | A role with `sla_view` but not `report_view`; then one with `report_view` only | The first has no **Monitoring** in the navigation; the second has it, and the report opens. Without `report_export`, no **Export PDF** |
- ⚙️ **Developer Guide & Release Confidence**:
  - Damping: `DecideAlert` in `packetpulsego/pkg/monitormicroservice/monitordomain/shared/` (consecutive breaches, cooldown, escalation); runner: `packetpulsego/pkg/monitormicroservice/monitorservice/MonitorScheduleRunner.go` (claims due schedules with `FOR UPDATE SKIP LOCKED`).
  - Results are partitioned by month and rolled up daily per site **and family** (`packetpulsego/pkg/common/dbclient/migrations/0012_2026_10_01_result_partitioning_and_retention.sql`, `0023`); the retention runner keeps partitions three months ahead.
  - 🔒 Report-only IPv6 never enters availability (`NOT report_only` in the rollup and the SLA queries).
  - The screen: `packetpulseflutter/lib/monitormicroservice/presentation/screens/MonitorScreen.dart`; its navigation entry is gated on `report_view` in `packetpulseflutter/lib/common/presentation/PacketPulseShell.dart`.
  - Coverage: `packetpulsetest/golang/monitorconformance/`, `packetpulsego/pkg/monitormicroservice/**`, `packetpulseflutter/test/monitor_screen_test.dart`, `packetpulseflutter/test/shell_test.dart`.

---
