## Group 7 — Platform and Settings

The operator of Pingle itself, each person's own settings, and the public face.

---

### 7.1 🏛️ Platform console and licences

**Screen:** Administer → **Platform** (superuser only) · **Routes:** `/platform/*` · **Access:** platform superuser

- 🌟 **Commercial Presentation & Sales Pitch**: Tenants are created deliberately, never self-asserted: signing up cannot make an organisation or join one. Licences carry seats, sites, a period and a price, scale in proportion when sold for an unusual term, and keep a history of every change with who made it and why.
- 📖 **User Guide & Operational Flow**: **Organisations → Add**; place an unassigned account in it (as owner, or with a role); **Issue licence** with plan, seats, sites, months and currency. Per licence: **Suspend**, **Resume**, **Revoke**, **Renew**, change seats. **Owners** lists an organisation's owners and resets one's authenticator.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `PLAT-001` | Superuser | Onboard a customer | Create `ACME`; assign a signed-up account as owner; issue a 12-month, 10-seat licence | The owner's next sign-in lands in ACME with every Administrator capability |
  | `PLAT-002` | Superuser | Suspend and resume | Suspend ACME's licence with a reason; resume it | While suspended: running a diagnostic is refused, reading is not. History shows both, with the reason |
  | `PLAT-003` | Superuser | Seats cannot drop below who is signed in | Four people signed in; set seats to 3 | Refused, naming the people signed in |
  | `PLAT-004` | Superuser | Renewal scales the price | Renew a 12-month licence for 36 months | Three times the period price; the new end runs from the current end (or today if lapsed) |
  | `PLAT-005` | Superuser | A superuser is not a tenant | Call `GET /dnssite/list` as the superuser | 400 — no organisation. 🛑 **Must NOT** answer with every organisation's sites |
  | `PLAT-006` | Organisation owner | Signing up never joins an organisation | Sign up with an address at ACME's domain | The account waits unassigned; it sees only **Test from this device** and **Settings** |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pinglego/pkg/platformmicroservice/`; `pingleaccess.RequireSuperUser` on the console group (Appendix A); licence changes recorded in the licence history and the activity trail.
  - Coverage: `pingletest/golang/tenancyassignment/`, `pinglego/pkg/platformmicroservice/**`, `pingleflutter/test/platform_console_screen_test.dart`.

---

### 7.2 ⚙️ Settings

**Screen:** Administer → **Settings** · **Routes:** `POST /user/appearance`, `POST /user/password`

- 🌟 **Commercial Presentation & Sales Pitch**: Five themes in light and dark — including a warm, low-blue-light theme for night shifts and a high-contrast one — every one contrast-checked, with status colours that never change meaning. They follow the person, not the device.
- 📖 **User Guide & Operational Flow**: Light, dark or follow the device; theme; density; reduce motion; change password. Pingle is in English; there is no language choice.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SET-001` | Engineer | Appearance follows the person | Choose a dark theme on a laptop; sign in on another browser | The same theme there |
  | `SET-002` | Engineer | Status colours keep their meaning | Switch through every theme on a result with OK, Degraded and Breached rows | Green, amber and red in every theme, each with its icon |
  | `SET-003` | Engineer | English only | Look for a language picker in the bar and in Settings | None. Every label is English, served by the server. 🛑 **Must NOT** show a key such as `s812` in place of a label |
  | `SET-004` | Engineer | Change password | Change it; sign out; sign in with the old, then the new | Old refused, new accepted (then the second step) |
- ⚙️ **Developer Guide & Release Confidence**:
  - Strings: `scripts/intellicodegen/pinglestrings.py` generates the Go catalogue and the Dart index; entries are positional, so retired ones stay in `DEPRECATED`. Help: `pinglego/pkg/initmicroservice/initconstants/PingleHelp.go`.
  - Coverage: `pingleflutter/test/settings_screen_test.dart`, the `translation` suite.

---

### 7.3 🌐 Public site and self-test

**Pages:** `https://pingle.rummaan53.com/` and `/selftest.html` (`pingleweb/`)

- 🌟 **Commercial Presentation & Sales Pitch**: The product site says only what the product does today, and the public self-test lets a prospect measure their own connection in the browser before talking to anyone.
- 📖 **User Guide & Operational Flow**: **Sign in** opens the app at `/app/`; **Test my connection** opens the self-test.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `WEB-001` | Visitor | The site is current | Read the capabilities | Two-step sign-in, IPv4 and IPv6, CSV into your systems, check-ins, the device test with line speed. 🛑 **Must NOT** promise 23 languages |
  | `WEB-002` | Visitor | Site and app are different documents | Open `/` and `/app/` | The marketing page and the app respectively — never the same document |
  | `WEB-003` | Visitor | The self-test runs in the browser | **Test my connection** → run | Round trips, jitter and loss for each resolver, measured from the visitor's connection |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pingleweb/index.html`, `pingleweb/selftest.html`; the probe engine's tests: `pingleweb/assets/pingle-probe.test.mjs` (part of `unit`).
  - The deploy asserts `/` and `/app/` differ, so the nginx misroute that once served a blank app is caught by the deploy.

---
