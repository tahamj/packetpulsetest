## Group 3 — Diagnostics

The core of the product: what to test, testing it against a ticket, and reading what came back.

---

### 3.1 🌐 Sites

**Screen:** Configure → **DNS sites** · **Routes:** `GET /dnssite/list`, `POST /dnssite/add`, `POST /dnssite/bulkimport`, `POST /dnssite/importfile`, `PUT/DELETE /dnssite/{dnsSiteId}` · **Capabilities:** `dns_site_view`, `dns_site_manage` — the page is offered only to whoever holds both

- 🌟 **Commercial Presentation & Sales Pitch**: Every city its own destinations, as many as it needs: give each endpoint its position and a field engineer's test sweeps the city they are standing in, without their choosing anything. Name the reflector beside an endpoint and a ticket's UDP test measures the path to it. Drop the customer's spreadsheet on the page to add or place hundreds at once — one bad row never rejects the rest — and PacketPulse refuses to be turned against its own host: loopback, link-local and cloud-metadata addresses are never probed, whatever a site says.
- 📖 **User Guide & Operational Flow**: **Add site** with a name, an IP address or hostname, an optional circuit ID, its **City or region**, its **Latitude** and **Longitude** (both or neither) and an SLA policy. The list shows each endpoint's position, or *No position*. **Import file** takes a CSV file dropped on the page or chosen, or rows pasted from a spreadsheet: the first row names the columns — *Name*, *Address*, *City*, *Latitude*, *Longitude*, only *Address* required, in any order. A row whose address is already monitored updates that endpoint with the columns given and leaves the rest; any other row adds one. Each row is reported, a refused one by its line. **Bulk import** still takes plain addresses by lines, commas or spaces; `Branch 12=10.0.0.1` names a site. There is no limit on how many endpoints an organisation or a city has.
  - **Who is given the page:** the endpoints are configuration, so **DNS sites** is in the rail only for whoever manages them (`dns_site_manage`, with `dns_site_view` to read the list). An engineer or a viewer reads the endpoints under `dns_site_view` alone, which is what lets **Run diagnostic** offer them (§4.1), and is not given the page (since 2026-10-09).
  - **UDP reflector address** (optional) is where the endpoint's PacketPulse reflector answers: a host or IP, with `:port` when its signalling port is not 50000. Often not the endpoint's own address — a router cannot run the daemon, a small host beside it can. Left blank, a ticket's UDP test measures to the PacketPulse network instead (§4.2). PacketPulse's own address there does the same: this server's reflector answers through the app, never on port 50000 (`UDP-026`). Whoever manages the endpoints gets a key icon on each row, **Reflector install key**, with the key and the install command to copy; the reflector answers only tests this organisation signs for this endpoint.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SITE-001` | Administrator | Add by address and by name | Add `Mumbai POP` = `203.0.113.10`; add `Resolver` = `one.one.one.one` | Both listed; the hostname is resolved at test time, not stored as an address |
  | `SITE-002` | Administrator | One bad line does not sink a paste | Bulk import `Pune=203.0.113.11`, `not an address`, `Pune=203.0.113.11` | First added, second invalid with a reason, third a duplicate |
  | `SITE-003` | Administrator | Labels with spaces survive | Bulk import `Branch 12=10.0.0.12` | One site named `Branch 12`. 🛑 **Must NOT** split into `Branch` and `12` |
  | `SITE-004` | Engineer | The host is never a target | Add `127.0.0.1`, `169.254.169.254` and `::1`; run a diagnostic over them | Each result is a refusal (*destination refused by probe policy*), never a measurement. 🛑 **Must NOT** send traffic to loopback or cloud metadata |
  | `SITE-005` | RIVAL administrator | The same address in two organisations | RIVAL adds `203.0.113.10` too | Allowed — the endpoint is unique per organisation, not globally |
  | `SITE-006` | Administrator | A position is both halves or neither | Add a site with only a latitude; then a longitude of `190`; then both blank | Refused beside the field each time, nothing sent; the blank pair saves with *No position*. 🛑 **Must NOT** store half a position, or 0, 0 for none |
  | `SITE-007` | Administrator | A dropped spreadsheet adds and places | In a browser, **Import file**; drag a CSV with `Name,Address,City,Latitude,Longitude` and four rows, one with no address, onto the page; **Import** | *2 added and 1 updated, of 4 rows*-style summary; the row with no address named by its line; the list shows each position. 🛑 **Must NOT** reject the whole file for one bad row |
  | `SITE-008` | Administrator | Positions for endpoints you already have | Edit a site's circuit and notes; import a file of `Address,Latitude,Longitude` naming it | The site gains its position; its name, city, circuit, notes and enabled state are unchanged. 🛑 **Must NOT** clear what the file had no column for |
  | `SITE-009` | Administrator | Rows pasted from Excel | Copy four rows (heading first) from Excel or Sheets; paste into **Import file**; **Import** | Read as tab-separated; each row reported. A Windows-saved CSV with `São Paulo` imports with the name intact |
  | `SITE-010` | RIVAL administrator | An import never reaches across | RIVAL imports a file naming ACME's addresses with new positions | RIVAL gains its own endpoints; ACME's are unchanged |
  | `SITE-011` | Administrator | A reflector address is kept | Edit a site: **UDP reflector address** `Reflector.Pune.Example:50001`; save; edit only its notes; save; then `PUT /dnssite/{id}` without `reflector_address`, as an app from before 2026-10-09 sends it | Stored as `reflector.pune.example:50001` and still there after both edits; *Optional* — no `*` on the label. Clearing the field and saving clears it. 🛑 **Must NOT** clear the address on an edit that did not touch it |
  | `SITE-012` | Administrator | A wrong reflector address is refused on its field | Enter `not an address`, then `203.0.113.20:70000` | Refused under the field, the rest of the edit not saved |
  | `SITE-013` | Administrator | The install key | Press the key icon on a site; **Copy** the install command; press it again | The key and the `install-reflector.sh --key …` line, copied with *Copied to the clipboard.*; the same key every time for that site. An engineer never reaches the key: no **DNS sites** in their rail (`SITE-014`), and the API refuses them (`UDP-010`) |
  | `SITE-014` | Engineer, then Viewer, then Administrator | The inventory is the endpoint managers' | Sign in as an engineer; look for **DNS sites** in the rail, then open **Run diagnostic**; repeat as a Viewer; then as an administrator | No **DNS sites** for the engineer or the Viewer, while **Run diagnostic** still lists the endpoints for the engineer; the administrator has **DNS sites** under Configure. 🛑 **Must NOT** offer an engineer the endpoint inventory, or take the endpoints off their **Run diagnostic** |
- ⚙️ **Developer Guide & Release Confidence**:
  - Policy: `packetpulsego/pkg/common/probeguard/ProbeGuardPolicy.go` — every address a name resolves to is checked, and one refused address refuses them all; IPv4-mapped IPv6 is judged as IPv4.
  - Positions: `latitude`/`longitude` on `dns_site`, both or neither, in `packetpulsego/pkg/common/dbclient/migrations/0029_2026_10_07_endpoint_location_and_report_branding.sql`. The nearest city is a pure function, `ChooseNearestCity` in `packetpulsego/pkg/dnssitemicroservice/dnssitedomain/shared/DnsSite.go`: the nearest **enabled** endpoint with a position names the city (its region, ignoring case), and every enabled endpoint in that city is swept; with no city, the endpoints at the nearest one's position.
  - Import: `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteImport.go` reads the heading row, the delimiter (tab, semicolon or comma) and a spreadsheet's byte-order mark; an existing endpoint is updated with `COALESCE`, so a column the file lacks is never cleared. Up to 5000 rows a file. The drop zone and chooser are `packetpulseflutter/lib/common/services/PacketPulseFileSource.dart`, browser-only; elsewhere the paste box alone.
  - Reflector address: `dns_site.reflector_address` in `packetpulsego/pkg/common/dbclient/migrations/0031_2026_10_09_udp_split_and_endpoint_reflector.sql`, validated by `ValidateReflectorAddress` in `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteAddressValidator.go`. The edit is whole-row, so an edit that does not name the field — from an app that predates it — keeps the stored address (`DnsSiteUpdate` in `packetpulsego/pkg/dnssitemicroservice/dnssiteservice/DnsSiteService.go`); imports never touch it.
  - Who is offered the page: the `dns_sites` destination in `packetpulseflutter/lib/common/presentation/PacketPulseShell.dart` asks for `dns_site_manage` and `dns_site_view`, so `DnsSiteScreen` has no read-only mode: it always offers adding, importing, editing, the key and removal. The rail only hides it; the server still refuses every change without `dns_site_manage` (Appendix A).
  - Coverage: `packetpulsego/pkg/dnssitemicroservice/**`, `packetpulsego/pkg/common/probeguard/`, `packetpulseflutter/test/dns_site_screen_test.dart`, `packetpulseflutter/test/dns_site_position_import_test.dart`, `packetpulseflutter/test/shell_test.dart`.
  - ⚠️ **TRAP** — a licence carries a **site limit** (shown on the Platform console), but adding or importing sites does not check it yet. Do not file a site count above the limit as a regression; it is a known gap.

---

### 3.2 🩺 Diagnostics and results

**Screen:** a ticket's result, from **History**; no screen starts a sweep (since 2026-10-07: **Run diagnostic** is the test from this device, §4.1) · **Routes:** `POST /diagnostic/submit`, `GET /diagnostic/{requestId}`, `GET /diagnostic/{requestId}/report.pdf` · **Capabilities:** `diagnostic_run`, `diagnostic_view_own` or `diagnostic_view_all`, `report_export` · **Licence:** required to run

- 🌟 **Commercial Presentation & Sales Pitch**: One ticket, one report. Every figure a NOC argues about — loss, latency, RFC 3550 jitter, MOS — from PacketPulse's server on its schedule, and from the engineer's own device on **Run diagnostic**. A lapsed licence stops new tests but never takes away the evidence already gathered.
- 📖 **User Guide & Operational Flow**: A sweep runs on PacketPulse's server: on a schedule (§5.1), from the Results API, or with `POST /diagnostic/submit`. **Run diagnostic** no longer starts one (since 2026-10-07); it is the test from this device against one endpoint (§4.1). A sweep given a position tests the endpoints of the city nearest it (§3.1); with no position — location refused, no fix, or no endpoint placed yet — every enabled endpoint is swept. Only whoever manages the endpoints (`dns_site_manage`) may name endpoints instead; an engineer's list is ignored. The result says **Tested from** — the device's position as the test ran, beside the sign-in place — and **Endpoints**: the nearest city and how far away it was, every endpoint, or chosen by hand. The result shows headline cards (sites reachable, average loss, average jitter), then a row per site and family: reachable, the packet line verbatim (*Sent = 4, Received = 4, Lost = 0*), round trips, jitter, MOS with its band and SLA grade. **Export PDF** and **Export CSV** sit at the top.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DIAG-001` | Tester | A sweep files against the ticket | `POST /diagnostic/submit` against every enabled site with TT `TT-MAN-001` | Status *Completed*; one row per site (and family); the packet line exactly as specified |
  | `DIAG-002` | Tester | A ticket keeps every attempt | Submit `TT-MAN-001` again | **History** for the ticket shows both attempts, newest first |
  | `DIAG-003` | Engineer | A failure is reported, not traced | Include a black-holed address (`203.0.113.99`); submit; open the result, its PDF and the **Dashboard** | The row shows the failure and its packet line. No path, hops or fault verdict anywhere, and no *Where the faults lay* card. 🛑 **Must NOT** offer a *Trace failures* switch |
  | `DIAG-004` | Engineer | The PDF is evidence | **Export PDF** | Named `packetpulse-<TT>-<UTC time>.pdf`; Customer ID, TT, UTC timestamps, Jitter and MOS columns, average loss and jitter cards; a long IPv6 address wraps onto two lines; the foot names who ran it and where — *Triggered by Asha Rao  -  at MG Road, Pune*. 🛑 **Must NOT** truncate an address |
  | `DIAG-005` | Engineer | Loss and jitter on the ticket | Open the result; open **History** | Headline cards show average loss and average jitter; the history row shows the same figures |
  | `DIAG-006` | Engineer | A lapsed licence stops new runs only | Platform suspends the licence; submit a sweep, and attach a device run (§4.1); open an old one and export it | Both refused (`licence_suspended`, who can renew named); the old result opens and exports. 🛑 **Must NOT** hide recorded evidence |
  | `DIAG-007` | RIVAL engineer | Another organisation's ticket is not found | Open `/diagnostic/<ACME request id>` as RIVAL | 404. 🛑 **Must NOT** reveal that the ticket exists |
  | `DIAG-008` | Tester | A malformed id is refused before it is looked up | `GET /diagnostic/not-a-uuid` | 400 |
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
  - No screen starts a sweep: the app's sweep form was removed when **Run diagnostic** became the test from this device (2026-10-07, §4.1), and the client no longer calls `POST /diagnostic/submit`. The route stays, for the scheduler's runs, the Results API and scripts. The case for one page holding the sweep and the device test is retired, its number left vacant.
  - Path analysis was removed in `packetpulsego/pkg/common/dbclient/migrations/0027_2026_10_05_path_analysis_removed.sql`: `ping_hop` and the fault verdict columns are no longer written, and the rows already stored stay. A client built before then still sends `trace_failures`; the server accepts and ignores it, because decoding is strict and refusing it would fail every run that client made.
  - Endpoint choice: `chooseEndpoints` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go`; the handler empties `dns_site_ids` for anyone without `dns_site_manage`. The request records `test_location_*`, `endpoint_selection` (`chosen`, `nearest`, `all`), `selection_city` and `selection_distance_km` (migration 0029).
  - Ticket figures (`avg_loss_pct`, `max_loss_pct`, `avg_jitter_ms`) are computed in `RequestFinish` from the run's counted results (`packetpulsego/pkg/common/dbclient/migrations/0022_2026_10_03_loss_and_jitter.sql`).
  - The PDF is built on the fly from the database each time it is asked for (`packetpulsego/pkg/pingmicroservice/pingreport/PingReportPdfBuilder.go`); nothing is written to disk. Its foot names the person who ran the test and the place their session checked in from.
  - Coverage: `packetpulsego/pkg/diagnosticmicroservice/**`, `packetpulsego/pkg/pingmicroservice/**`, `packetpulseflutter/test/diagnostic_result_selection_test.dart`, `packetpulseflutter/test/diagnostic_history_screen_test.dart`, the `tenancy` and `contract` suites.

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
- 📖 **User Guide & Operational Flow**: **Export CSV** saves `packetpulse-<TT>-<UTC time>.csv` beside the PDF of the same ticket. One row per site and family; empty cells (not zeros) where a probe had no reply. The last two columns, `test_latitude` and `test_longitude`, say where the device was when the test ran; empty when it was not known.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CSV-001` | Engineer | The file has the contracted header | **Export CSV**; open it | Header exactly as in `packetpulsetest/contracts/result_export_columns.json`, `result_id` first; times in UTC (`…Z`); an unreachable row has empty round-trip cells |
  | `CSV-002` | Engineer | A formula is written as text | Name a site `=HYPERLINK("http://x","y")`; run; export; open in Excel or Sheets | The cell shows the text starting `'=`. 🛑 **Must NOT** become a live formula or link |
  | `CSV-003` | Engineer | Names in any script survive | Name a site `पुणे केंद्र`; export | The name intact (UTF-8). 🛑 **Must NOT** show mojibake |
  | `CSV-004` | Viewer | No export without `report_export` | Open a result as a Viewer | No **Export CSV** or **Export PDF**; the route answers 403 |
  | `CSV-005` | Tester | The download is on the record | Export a CSV; open **Activity** | An *Exported* entry naming the diagnostic |
  | `CSV-006` | Engineer | Where the test ran, as numbers | Export `DIAG-012`'s CSV; export a scheduled sweep's | Six-place `test_latitude`/`test_longitude` on every row of the first; empty on the second. A longitude west of Greenwich is a plain negative number. 🛑 **Must NOT** be written as text starting `'-`, or as 0 |
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
