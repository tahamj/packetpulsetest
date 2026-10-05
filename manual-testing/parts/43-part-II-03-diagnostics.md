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

**Screen:** Operate → **Run diagnostic** · **Routes:** `POST /diagnostic/submit`, `GET /diagnostic/{requestId}`, `GET /diagnostic/{requestId}/report.pdf` · **Capabilities:** `diagnostic_run`, `diagnostic_view_own` or `diagnostic_view_all`, `report_export` · **Licence:** required to run

- 🌟 **Commercial Presentation & Sales Pitch**: One form, one sweep, one report against the ticket. Every figure a NOC argues about — loss, latency, RFC 3550 jitter, MOS — and a verdict on *where* the fault lies, with the evidence. A lapsed licence stops new tests but never takes away the evidence already gathered.
- 📖 **User Guide & Operational Flow**: Enter **Customer ID** and **TT number**, pick sites (or leave empty for every enabled site), choose packet count and timeout, tick **Trace failures** for the path to anything that fails. The result shows headline cards (sites reachable, average loss, average jitter), then a row per site and family: reachable, the packet line verbatim (*Sent = 4, Received = 4, Lost = 0*), round trips, jitter, MOS with its band, SLA grade, verdict, and — for a failure — the hops with the first lossy one marked. **Export PDF** and **Export CSV** sit at the top.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DIAG-001` | Engineer | A sweep files against the ticket | Run against every enabled site with TT `TT-MAN-001` | Status *Completed*; one row per site (and family); the packet line exactly as specified |
  | `DIAG-002` | Engineer | A ticket keeps every attempt | Run `TT-MAN-001` again | **History** for the ticket shows both attempts, newest first |
  | `DIAG-003` | Engineer | Tracing marks where loss begins | Include a black-holed address (`203.0.113.99`) with **Trace failures** | Its row lists hops; the first hop with loss is marked. A trace that runs out of time leaves the result intact without hops |
  | `DIAG-004` | Engineer | The PDF is evidence | **Export PDF** | Named `packetpulse-<TT>-<UTC time>.pdf`; Customer ID, TT, UTC timestamps, Jitter and MOS columns, average loss and jitter cards; a long IPv6 address wraps onto two lines; the foot names who ran it and where — *Triggered by Asha Rao  -  at MG Road, Pune*. 🛑 **Must NOT** truncate an address |
  | `DIAG-005` | Engineer | Loss and jitter on the ticket | Open the result; open **History** | Headline cards show average loss and average jitter; the history row shows the same figures |
  | `DIAG-006` | Engineer | A lapsed licence stops new runs only | Platform suspends the licence; run a diagnostic; open an old one and export it | Run refused (`licence_suspended`, who can renew named); the old result opens and exports. 🛑 **Must NOT** hide recorded evidence |
  | `DIAG-007` | RIVAL engineer | Another organisation's ticket is not found | Open `/diagnostic/<ACME request id>` as RIVAL | 404. 🛑 **Must NOT** reveal that the ticket exists |
  | `DIAG-008` | Tester | A malformed id is refused before it is looked up | `GET /diagnostic/not-a-uuid` | 400 |
  | `DIAG-009` | Engineer | The PDF is made from the record when it is asked for | Export a ticket's PDF twice, a minute apart; look for a stored copy on the server | Both carry the same measurements, drawn from the stored results at the moment each was asked for; no PDF is kept on the server. 🛑 **Must NOT** depend on a file kept on disk |
  | `DIAG-010` | Engineer | A colleague's test cannot be exported | As an engineer, `GET /diagnostic/<colleague's request id>/report.pdf` (and `.csv`) | 404, as though it did not exist. 🛑 **Must NOT** hand an engineer someone else's evidence |
- ⚙️ **Developer Guide & Release Confidence**:
  - Engine: `packetpulsego/pkg/pingmicroservice/pingprobe/PingProbeRunner.go` (pro-bing, unprivileged ICMP, TCP fallback); jitter: `packetpulsego/pkg/pingmicroservice/pingprobe/PingVoiceQuality.go` (`InterarrivalJitter`, RFC 3550).
  - 🔒 Results are saved under their **own** deadline, never the sweep's or the trace's: a traced sweep that runs long still stores everything it measured (`load` suite §4.5 proves it, and goes red with the old bug restored).
  - Ticket figures (`avg_loss_pct`, `max_loss_pct`, `avg_jitter_ms`) are computed in `RequestFinish` from the run's counted results (`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql`).
  - The PDF is built on the fly from the database each time it is asked for (`packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`); nothing is written to disk. Its foot names the person who ran the test and the place their session checked in from.
  - Coverage: `packetpulsego/pkg/diagnosticmicroservice/**`, `packetpulsego/pkg/pingmicroservice/**`, `packetpulseflutter/test/diagnostic_submit_screen_test.dart`, the `tenancy` and `contract` suites.

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
  | `V6-005` | Engineer | Traceroute over IPv6 shows its hops | Trace a failing IPv6 address on a host with IPv6 | Intermediate hops listed (not only the destination) |
- ⚙️ **Developer Guide & Release Confidence**:
  - `ping_result.ip_version` (4, 6 or NULL when nothing was probed) and `report_only`; the daily rollup is keyed by family (`packetpulsego/pkg/common/dbclient/migrations/0023_2026_10_04_dual_stack_results.sql`).
  - `probeguard.ResolveAllAndCheck` returns one checked address per family; the runner probes them in parallel and marks IPv6 report-only only when there is more than one family.
  - ICMPv6 Time Exceeded parsing walks extension headers (`quotedSequenceIPv6` in `packetpulsego/pkg/pingmicroservice/pingprobe/PingTraceRunner.go`).

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

**Screens:** Operate → **History**, **Dashboard** · **Routes:** `GET /diagnostic/list`, `GET /diagnostic/tt/{ttNumber}`, `GET /ping/dashboard`, `GET /diagnostic/faults`

- 🌟 **Commercial Presentation & Sales Pitch**: Every ticket investigated, with its loss and jitter on the row, and a management view of where faults lay over the last 90 days — the customer's own network, the access circuit or the carrier — for the supplier review.
- 📖 **User Guide & Operational Flow**: An engineer's **History** is the tests they ran, with the note *These are the tests you ran. Administrators see everyone's.* An administrator's is everyone's, and narrows by **Person**, **Place** (*Where it was run, or a site or region*), **Status**, **From** and **To**; **Clear filters** puts them back. Both search by TT number or Customer ID. Each row shows reachable/total, breaches, loss and jitter, and who ran it and where: *Customer CUST-1 · Asha Rao · at MG Road, Pune*. **Dashboard** shows sites, the last sweep, the licence, 30 days of availability (a day with nothing measured is a gap, never zero) and the fault breakdown.
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
