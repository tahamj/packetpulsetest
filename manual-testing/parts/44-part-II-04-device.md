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
