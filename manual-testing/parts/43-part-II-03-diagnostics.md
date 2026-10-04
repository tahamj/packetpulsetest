## Group 3 — Diagnostics

The core of the product: what to test, testing it against a ticket, and reading what came back.

---

### 3.1 🌐 Sites

**Screen:** Configure → **DNS sites** · **Routes:** `GET /dnssite/list`, `POST /dnssite/add`, `POST /dnssite/bulkimport`, `PUT/DELETE /dnssite/{dnsSiteId}` · **Capabilities:** `dns_site_view`, `dns_site_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Paste a whole inventory in one go — one bad line never rejects the rest — and Pingle refuses to be turned against its own host: loopback, link-local and cloud-metadata addresses are never probed, whatever a site says.
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
  - Policy: `pinglego/pkg/common/probeguard/ProbeGuardPolicy.go` — every address a name resolves to is checked, and one refused address refuses them all; IPv4-mapped IPv6 is judged as IPv4.
  - Coverage: `pinglego/pkg/dnssitemicroservice/**`, `pinglego/pkg/common/probeguard/`, `pingleflutter/test/dns_site_screen_test.dart`.
  - ⚠️ **TRAP** — a licence carries a **site limit** (shown on the Platform console), but adding or importing sites does not check it yet. Do not file a site count above the limit as a regression; it is a known gap.

---

### 3.2 🩺 Diagnostics and results

**Screen:** Operate → **Run diagnostic** · **Routes:** `POST /diagnostic/submit`, `GET /diagnostic/{requestId}`, `GET /diagnostic/{requestId}/report.pdf` · **Capabilities:** `diagnostic_run`, `diagnostic_view_all`, `report_export` · **Licence:** required to run

- 🌟 **Commercial Presentation & Sales Pitch**: One form, one sweep, one report against the ticket. Every figure a NOC argues about — loss, latency, RFC 3550 jitter, MOS — and a verdict on *where* the fault lies, with the evidence. A lapsed licence stops new tests but never takes away the evidence already gathered.
- 📖 **User Guide & Operational Flow**: Enter **Customer ID** and **TT number**, pick sites (or leave empty for every enabled site), choose packet count and timeout, tick **Trace failures** for the path to anything that fails. The result shows headline cards (sites reachable, average loss, average jitter), then a row per site and family: reachable, the packet line verbatim (*Sent = 4, Received = 4, Lost = 0*), round trips, jitter, MOS with its band, SLA grade, verdict, and — for a failure — the hops with the first lossy one marked. **Export PDF** and **Export CSV** sit at the top.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DIAG-001` | Engineer | A sweep files against the ticket | Run against every enabled site with TT `TT-MAN-001` | Status *Completed*; one row per site (and family); the packet line exactly as specified |
  | `DIAG-002` | Engineer | A ticket keeps every attempt | Run `TT-MAN-001` again | **History** for the ticket shows both attempts, newest first |
  | `DIAG-003` | Engineer | Tracing marks where loss begins | Include a black-holed address (`203.0.113.99`) with **Trace failures** | Its row lists hops; the first hop with loss is marked. A trace that runs out of time leaves the result intact without hops |
  | `DIAG-004` | Engineer | The PDF is evidence | **Export PDF** | Named `pingle-<TT>-<UTC time>.pdf`; Customer ID, TT, UTC timestamps, Jitter and MOS columns, average loss and jitter cards; a long IPv6 address wraps onto two lines. 🛑 **Must NOT** truncate an address |
  | `DIAG-005` | Engineer | Loss and jitter on the ticket | Open the result; open **History** | Headline cards show average loss and average jitter; the history row shows the same figures |
  | `DIAG-006` | Engineer | A lapsed licence stops new runs only | Platform suspends the licence; run a diagnostic; open an old one and export it | Run refused (`licence_suspended`, who can renew named); the old result opens and exports. 🛑 **Must NOT** hide recorded evidence |
  | `DIAG-007` | RIVAL engineer | Another organisation's ticket is not found | Open `/diagnostic/<ACME request id>` as RIVAL | 404. 🛑 **Must NOT** reveal that the ticket exists |
  | `DIAG-008` | Tester | A malformed id is refused before it is looked up | `GET /diagnostic/not-a-uuid` | 400 |
- ⚙️ **Developer Guide & Release Confidence**:
  - Engine: `pinglego/pkg/pingmicroservice/pingprobe/PingProbeRunner.go` (pro-bing, unprivileged ICMP, TCP fallback); jitter: `pinglego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go` (`InterarrivalJitter`, RFC 3550).
  - 🔒 Results are saved under their **own** deadline, never the sweep's or the trace's: a traced sweep that runs long still stores everything it measured (`load` suite §4.5 proves it, and goes red with the old bug restored).
  - Ticket figures (`avg_loss_pct`, `max_loss_pct`, `avg_jitter_ms`) are computed in `RequestFinish` from the run's counted results (`pinglego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql`).
  - Coverage: `pinglego/pkg/diagnosticmicroservice/**`, `pinglego/pkg/pingmicroservice/**`, `pingleflutter/test/diagnostic_submit_screen_test.dart`, the `tenancy` and `contract` suites.

---

### 3.3 🌍 IPv4 and IPv6

**Where:** every result list, the PDF, the SLA report, the Results API and the CSV

- 🌟 **Commercial Presentation & Sales Pitch**: A site healthy over IPv4 and dark over IPv6 is a real, common, hard-to-prove fault. Pingle measures both, side by side, without letting an IPv6 problem the customer did not buy an SLA for change their availability figure.
- 📖 **User Guide & Operational Flow**: A dual-stack site has two rows. The IPv6 one carries a **not counted** tag (hover for why). The PDF marks it `icmp v6 *` with a footnote; the SLA report lists `(v6)` lines separately and `(v6*)` for report-only ones.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `V6-001` | Engineer | Both families, one counted | Add a dual-stack site (e.g. `one.one.one.one`) on a host with IPv6; run | Two rows: IPv4 counted, IPv6 tagged *not counted* with its tooltip |
  | `V6-002` | Engineer | An IPv6 failure does not breach the site | A dual-stack site whose IPv6 fails while IPv4 answers | No alert; the site counts as reachable; availability unchanged. 🛑 **Must NOT** raise an alert for report-only IPv6 |
  | `V6-003` | Engineer | IPv6-only counts | A site with only an IPv6 address that fails | Counted as a failure, alerting and grading as usual |
  | `V6-004` | Engineer | No IPv6 route is said plainly | On a server without IPv6, a dual-stack site | The IPv6 row reads *There is no IPv6 route from this vantage point, so this IPv6 address could not be tested from here.* 🛑 **Must NOT** report it as the site being down |
  | `V6-005` | Engineer | Traceroute over IPv6 shows its hops | Trace a failing IPv6 address on a host with IPv6 | Intermediate hops listed (not only the destination) |
- ⚙️ **Developer Guide & Release Confidence**:
  - `ping_result.ip_version` (4, 6 or NULL when nothing was probed) and `report_only`; the daily rollup is keyed by family (`pinglego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql`).
  - `probeguard.ResolveAllAndCheck` returns one checked address per family; the runner probes them in parallel and marks IPv6 report-only only when there is more than one family.
  - ICMPv6 Time Exceeded parsing walks extension headers (`quotedSequenceIPv6` in `pinglego/pkg/pingmicroservice/pingprobe/PingTraceRunner.go`).

---

### 3.4 📄 CSV download

**Where:** **Export CSV** on every result · **Route:** `GET /diagnostic/{requestId}/report.csv` · **Capability:** `report_export`

- 🌟 **Commercial Presentation & Sales Pitch**: The same results, in the columns an IT system imports, in one click — and safe to open: a site name typed as a spreadsheet formula is written as text, so an export can never run code on someone else's desk.
- 📖 **User Guide & Operational Flow**: **Export CSV** saves `pingle-<TT>-<UTC time>.csv` beside the PDF of the same ticket. One row per site and family; empty cells (not zeros) where a probe had no reply.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CSV-001` | Engineer | The file has the contracted header | **Export CSV**; open it | Header exactly as in `pingletest/contracts/result_export_columns.json`, `result_id` first; times in UTC (`…Z`); an unreachable row has empty round-trip cells |
  | `CSV-002` | Engineer | A formula is written as text | Name a site `=HYPERLINK("http://x","y")`; run; export; open in Excel or Sheets | The cell shows the text starting `'=`. 🛑 **Must NOT** become a live formula or link |
  | `CSV-003` | Engineer | Names in any script survive | Name a site `पुणे केंद्र`; export | The name intact (UTF-8). 🛑 **Must NOT** show mojibake |
  | `CSV-004` | Viewer | No export without `report_export` | Open a result as a Viewer | No **Export CSV** or **Export PDF**; the route answers 403 |
  | `CSV-005` | Tester | The download is on the record | Export a CSV; open **Activity** | An *Exported* entry naming the diagnostic |
- ⚙️ **Developer Guide & Release Confidence**:
  - One writer for every CSV: `pinglego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`; one projection and scanner: `pinglego/pkg/diagnosticmicroservice/diagnosticdomain/repository/DiagnosticExportPostgres.go` (the run's organisation must match the ticket's).
  - The browser download is `text/csv` (`pingleflutter/lib/common/services/PingleFileSaverWeb.dart`); desktop and phone use the share sheet.

---

### 3.5 🗂️ History and dashboard

**Screens:** Operate → **History**, **Dashboard** · **Routes:** `GET /diagnostic/list`, `GET /diagnostic/tt/{ttNumber}`, `GET /ping/dashboard`, `GET /diagnostic/faults`

- 🌟 **Commercial Presentation & Sales Pitch**: Every ticket investigated, with its loss and jitter on the row, and a management view of where faults lay over the last 90 days — the customer's own network, the access circuit or the carrier — for the supplier review.
- 📖 **User Guide & Operational Flow**: **History** searches by TT number or Customer ID; each row shows reachable/total, breaches, loss and jitter. **Dashboard** shows sites, the last sweep, the licence, 30 days of availability (a day with nothing measured is a gap, never zero) and the fault breakdown.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `HIST-001` | Engineer | Find a ticket | Search `TT-MAN-001` | Its attempts, with loss and jitter on each row |
  | `HIST-002` | Engineer | A failed search is not an empty result | Stop the API; search | The failure, with *Try again*. 🛑 **Must NOT** show *No diagnostics yet* |
  | `HIST-003` | Engineer | A device run reads as one | Attach a device test (§4.1); find it in History | Marked as measured on a device; its loss is *query* loss |
  | `HIST-004` | Engineer | A quiet day is a gap | A schedule paused for a day; open **Dashboard** | The availability line has a gap. 🛑 **Must NOT** plot 0% for a day with nothing measured |
- ⚙️ **Developer Guide & Release Confidence**:
  - Row figures: `pingleflutter/lib/diagnosticmicroservice/presentation/widgets/DiagnosticFigures.dart`.
  - Coverage: `pingleflutter/test/diagnostic_history_screen_test.dart`, `pingleflutter/test/dashboard_screen_test.dart`.

---
