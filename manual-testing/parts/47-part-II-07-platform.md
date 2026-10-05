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
