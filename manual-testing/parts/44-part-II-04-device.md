## Group 4 — The Customer's Side

Measuring from where the person is, not from where PacketPulse is.

---

### 4.1 📱 From this device

**Screen:** Operate → **Run diagnostic**, bottom half (**From this device**) · **Routes:** `GET /organisation/settings`, `POST /diagnostic/clientobservation` (`diagnostic_run`, licensed); the target is set with `PUT /organisation/settings` (`staff_manage`), not on this screen

- 🌟 **Commercial Presentation & Sales Pitch**: "Is it slow for me?" answered from the customer's own connection, in one tap, with no target to choose or explain: everyone in the organisation measures the same host, so results compare. It shows the device's IPv4 **and** IPv6 addresses, then files the run against the ticket beside the server's figures — labelled as measured on a device.
- 📖 **User Guide & Operational Flow**: The bottom half of **Run diagnostic**; someone who may not run a sweep, or has no organisation, gets this half alone, with no ticket fields. **Start test** runs it against the organisation's host, or Cloudflare DNS when none is chosen. The target is internal: the screen does not name it, and nobody changes it here. A finished run can be attached to the ticket typed at the top of the page. The line-speed test is hidden.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `DEV-001` | Administrator | The target is internal | Open **Run diagnostic**; run a device test | No *Testing against* panel, no **Change target**, no dropdown or free host field, and no *Target* line in *What this run found*. The address card says the address came from *a public address lookup over the same connection*. 🛑 **Must NOT** name a host anywhere on the page: no *1.1.1.1*, *Cloudflare DNS* or the organisation's own |
  | `DEV-002` | Administrator | The internal target is still what is measured | `PUT /organisation/settings` with `device_test_target` `dns.google`; open **Run diagnostic**; run a device test with the browser's network panel open | Every probe goes to `dns.google`, and the screen names no host. 🛑 **Must NOT** fall back to Cloudflare while a target is set |
  | `DEV-003` | Administrator | The target saves alone | `PUT /organisation/settings` with a new `device_test_target` and the current `require_checkin_location` | The *Require location* rule (§1.2) is unchanged. 🛑 **Must NOT** reset another setting |
  | `DEV-004` | Engineer | Both addresses | Run from a dual-stack connection; then from one without IPv6 | *IPv4 a.b.c.d · IPv6 2401:…*; then *No IPv6 connectivity* |
  | `DEV-005` | Engineer | The speed test is hidden | Open **Run diagnostic**; then open a ticket that had a speed filed before the test was hidden | No **Measure speed** button and no ~12 MB note. The older ticket still shows its line speed. 🛑 **Must NOT** move any data to the speed-test endpoints |
  | `DEV-006` | Engineer | Attach to a ticket | Type TT `TT-MAN-002` at the top of the page; run a device test; **Attach to a ticket** | The ticket shows the device run, *measured on a device*, named *Device test* with no address, and no line speed; loss is *query* loss. Its **Export PDF** and **Export CSV** name it the same way; an error that quoted the host says *the target*. 🛑 **Must NOT** show the host the device measured (stored with the result, never served) |
  | `DEV-007` | Viewer | Only people managers change the target | As a Viewer, `PUT /organisation/settings` | 403. No **Change target** on the screen, for anyone |
  | `DEV-008` | Engineer | An unreadable setting does not block the test | Stop the API after the screen loads; reopen it | The test still runs, against Cloudflare DNS, and no error about the target is shown |
  | `DEV-009` | Superuser | No organisation, still a test | Sign in as the platform superuser; open **Run diagnostic** | The device test alone, with no ticket fields, testing Cloudflare DNS. It needs no organisation and no licence |
- ⚙️ **Developer Guide & Release Confidence**:
  - A browser cannot send ICMP, so the web build measures DNS-over-HTTPS (Cloudflare) or an HTTPS reach to the chosen host; the app on a desktop or phone can also ping it. Jitter is RFC 3550 with the standard deviation beside it.
  - Target hidden: the *Testing against* panel, the *Target* line and the **Change target** dialog are kept and tested behind `testTargetShown = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`, as the speed test is.
  - A filed device run keeps its host in `ping_result`. It is replaced when read back — the diagnostic detail, the Results API's ticket, the PDF, and every CSV (ticket, pulled, scheduled push) — by `hideDeviceTargets` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/DiagnosticService.go` and `Row.shown` in `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/DiagnosticExport.go`. The raw `/ping/run/{runId}` routes, which the app does not call, still show it. The page's wording is held to naming no host by `TestTheRunDiagnosticPageNamesNoDeviceTarget` in `packetpulsego/pkg/initmicroservice/initconstants/PacketPulseStrings_test.go`.
  - Target normalisation (pasted URL → host, lower case, canonical IPs, zones refused): `NormaliseTestTarget` in `packetpulsego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`.
  - Speed, hidden: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientThroughput.dart` (Cloudflare `__down`/`__up`, median after a warm-up) is kept and tested behind `speedTestEnabled = false` in `packetpulseflutter/lib/common/config/PacketPulseConfig.dart`; turning it back on is that one line. Egress: `packetpulseflutter/lib/diagnosticmicroservice/service/ClientEgress.dart`.
  - Coverage: `packetpulseflutter/test/client_probe_screen_test.dart`, `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart`, `packetpulseflutter/test/client_throughput_test.dart`, `packetpulseflutter/test/client_egress_test.dart`.

---
