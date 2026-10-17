## Group 7 — Platform and Settings

The operator of PacketPulse itself, each person's own settings, and the public face.

---

### 7.1 🏛️ Platform console and licences

**Screen:** Administer → **Platform** (superuser only) · **Routes:** `/platform/*` · **Access:** platform superuser

- 🌟 **Commercial Presentation & Sales Pitch**: Every customer runs on a licence PacketPulse signs, so nobody can grant themselves an organisation, more people or a longer term. Licences carry the people covered, sites, a period and a price, scale in proportion when sold for an unusual term, and keep a history of every change with who made it and why. A customer's server needs nothing from PacketPulse to check one: the signature is checked against a key built into the server.
- 📖 **User Guide & Operational Flow**: The console is PacketPulse's own server; a customer's server has none. **Organisations → Add**; **Issue licence** with plan, people, sites, months and currency. Per licence: **Suspend**, **Resume**, **Revoke**, **Renew**, change the people and sites it covers. **Download licence file** asks for the **Owner's email** and saves the signed file; copy it into `LICENCE_DIR` on the customer's server. It is read at every sign-in, so no restart is needed. The first person to sign up there with that address, proved by an emailed code, becomes the Administrator. Several files for one organisation do not add up: the one that runs longest is the licence, and of two that end the same day, the one issued later. An old file can stay beside its renewal, and what the files are called never matters.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `PLAT-001` | Superuser | Onboard a customer | Create `ACME`; issue a 12-month licence for 10 people; **Download licence file** naming `owner@acme.example`; copy it into a test server's `LICENCE_DIR`; sign up there with that address | The emailed code completes the sign-up, and the owner lands in ACME with every Administrator capability |
  | `PLAT-002` | Superuser | Suspend and resume | Suspend ACME's licence with a reason; download its file again; resume | While suspended: running a diagnostic is refused, reading is not. History shows both, with the reason |
  | `PLAT-003` | Superuser | A smaller licence keeps everyone but the owner out | ACME has 4 people; change the licence to 3; replace the file; an engineer signs in, then the owner | The engineer is refused (403, *Your organisation has more users than its licence covers…*). The owner signs in, because the owner is who can fix it |
  | `PLAT-004` | Superuser | Renewal scales the price | Renew a 12-month licence for 36 months | Three times the period price; the new end runs from the current end (or today if lapsed) |
  | `PLAT-005` | Superuser | A superuser is not a tenant | Call `GET /dnssite/list` as the superuser | 400 — no organisation. 🛑 **Must NOT** answer with every organisation's sites |
  | `PLAT-006` | Stranger | Signing up never makes or joins an organisation | On a customer's server, sign up with an address its licence does not name | Refused: *Only the owner named in this server's licence can sign up. Ask your administrator to add you.* 🛑 **Must NOT** create an account |
  | `PLAT-007` | Superuser, then Administrator | A licence for 20 refuses the 21st person | Issue ACME a licence for 20 people and install its file; the owner adds 19 colleagues; add a 21st. Then change the licence to 21, download the file again, replace it, and add the 21st again | 20th added; 21st refused (409 `seat_limit_reached`): *Your licence covers no more users…*, and **People on the licence** reads 20 of 20. After the new file, with no restart, the 21st is added and a 22nd is refused. 🛑 **Must NOT** add the 21st under the licence for 20, or leave an account for the refused person |
  | `PLAT-008` | Tester | An edited licence file admits nobody | Open ACME's installed file in a text editor; change `seat_limit` (or `valid_until`) and save; sign in, and add a person. Put the original file back | Both refused as no licence (403 `licence_expired`, *renew*), never as a full licence. With the original back, both work again. 🛑 **Must NOT** honour the edited number |
  | `PLAT-009` | Superuser, then Administrator | Licence files do not add up, and the later issue wins | ACME has a licence for 3 people, full. Change it to 5 people (its end date stays), download the file and save it **beside** the installed one under a name that sorts after it, e.g. `zz-acme.licence`. Add people until one is refused | The 4th and 5th are added and the 6th is refused: of two files ending the same day the later issue is the licence, and the two files' people are not added together. 🛑 **Must NOT** let the file names decide, or allow 8
  | `PLAT-010` | Superuser | A price with no digits is not free | Issue a licence with the amount typed as `.`, `-` or `+` | Refused on the amount. 🛑 **Must NOT** issue the licence at zero without *Complimentary* ticked |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/platformmicroservice/`; `packetpulseaccess.RequireSuperUser` on the console group (Appendix A); licence changes recorded in the licence history and the activity trail.
  - 🔒 Licence files are Ed25519-signed and verified by `packetpulsego/pkg/common/licence/Store.go` at every sign-in, against a key compiled into the server. `LICENCE_PUBLIC_KEY` may replace it only on a development build outside production, so a customer cannot swap in a key of their own.
  - Of one organisation's genuine files, `supersedes` in `Store.go` picks the licence: the latest `valid_until`, then the latest `issued_on`. Before October 2026 a same-day tie went to the first file name in sort order, so a customer could keep the bigger file of a downgrade just by naming it to sort first.
  - The console's own demo licence is dated from the UTC day (`packetpulsego/cmd/packetpulseserver/PacketPulseBootstrap.go`). Dated from the local day, a demo licence issued just after midnight in India did not start until UTC caught up, so the demo refused every sign-in for five and a half hours.
  - Coverage: `packetpulsetest/golang/tenancyassignment/` (the 20-user licence, and additions racing for the last place), `packetpulsego/pkg/usermicroservice/userservice/UserLicenceFile_test.go` (every state a licence file can be in, renewals beside and over the old file, one organisation's licence lending nothing to another), `packetpulsego/pkg/common/licence/`, `packetpulsego/pkg/platformmicroservice/**`, `packetpulseflutter/test/platform_console_screen_test.dart`.

---

### 7.2 ⚙️ Settings

**Screen:** Administer → **Settings** · **Routes:** `POST /user/appearance`, `POST /user/password`

- 🌟 **Commercial Presentation & Sales Pitch**: Five themes in light and dark — including a warm, low-blue-light theme for night shifts and a high-contrast one — every one contrast-checked, with status colours that never change meaning. They follow the person, not the device.
- 📖 **User Guide & Operational Flow**: Light, dark or follow the device; theme; density; reduce motion. **Change password** asks for the current password and the new one twice: 10 to 72 characters with a letter and a number, checked before anything is sent. It changes this account's password and nothing else: other devices stay signed in until they sign out or their 12 hours end, and the second step still follows the new password. Someone who signs in with their organisation's directory password is told to change it there, and offered no form; the owner, who keeps a local password under directory sign-in, keeps the form (`SET-004`, `SET-006`). PacketPulse is in English; there is no language choice. The same light / dark and appearance controls are on the sign-in screen, where they are remembered on that device until someone signs in.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SET-001` | Engineer | Appearance follows the person | Choose a dark theme on a laptop; sign in on another browser | The same theme there |
  | `SET-002` | Engineer | Status colours keep their meaning | Switch through every theme on a result with OK, Degraded and Breached rows | Green, amber and red in every theme, each with its icon |
  | `SET-003` | Engineer | English only | Look for a language picker in the bar and in Settings | None. Every label is English, served by the server. 🛑 **Must NOT** show a key such as `s812` in place of a label |
  | `SET-004` | Engineer | Change password | **Settings → Change password**: a wrong current password with a good new one; then the right one with a new one of nine characters (`नमस्ते 1a` too); then two new ones that differ; then a good change. Sign out; sign in with the old, then the new | Wrong current: *Your current password is incorrect.* under **Current password**, and still signed in. Nine characters: *at least 10 characters…* under **New password**; differing: *The two new passwords are not the same.*, neither sent. A good change: *Your password was changed…* and the fields cleared; the old password refused at sign-in, the new accepted (then the second step). 🛑 **Must NOT** sign the person out on a wrong current password |
  | `SET-006` | Engineer | A directory password is changed in the directory | With directory sign-in on, sign in as mapped staff, not the owner, and open **Settings**; try `POST /user/password` with their directory password as the current one; then do the same as the owner | Staff: *You sign in with your organisation's directory password, so change it in your directory…* and no form; the route answers 409 `password_from_directory`. The owner: the form, which changes their local password. 🛑 **Must NOT** accept a change that no sign-in would check |
  | `SET-005` | Visitor | Appearance before signing in | Signed out, use the sun/moon button and the palette on the sign-in screen (desk and phone width); reload; then sign in to an account set to another mode, and sign out | Each choice applies at once and survives the reload, with no request to `/user/appearance`. After sign-in the **account's** appearance is worn; after sign-out the sign-in screen keeps it. 🛑 **Must NOT** snap back to dark on sign-out, or save a signed-out choice to anyone's account |
- ⚙️ **Developer Guide & Release Confidence**:
  - Strings: `scripts/intellicodegen/packetpulsestrings.py` generates the Go catalogue and the Dart index; entries are positional, so retired ones stay in `DEPRECATED`. Help: `packetpulsego/pkg/initmicroservice/initconstants/PacketPulseHelp.go`.
  - Change password: `_PasswordForm` in `packetpulseflutter/lib/common/presentation/PacketPulseSettingsScreen.dart`; the session's `password_from_directory` (also on `/user/me`) decides whether it is shown. Server: `UserChangePassword` in `packetpulsego/pkg/usermicroservice/userservice/UserService.go`, held by `UserDirectorySignIn_test.go` against a real directory and by `UserRouteHandler_test.go` (422 and 409, never 401).
  - Coverage: `packetpulseflutter/test/settings_screen_test.dart`, the `translation` suite. Signed-out appearance: `packetpulseflutter/lib/common/services/PacketPulseAppearanceStore.dart`, tested in `packetpulseflutter/test/app_test.dart` and `packetpulseflutter/test/sign_in_test.dart`.

---

### 7.3 🌐 Public site and self-test

**Pages:** `https://packetpulse.rummaan53.com/` and `/selftest.html` (`packetpulseweb/`)

- 🌟 **Commercial Presentation & Sales Pitch**: The product site says only what the product does today, and leads straight to **Sign in**. The connection self-test still exists at `/selftest.html`, for a link sent by hand, but the site no longer advertises it (since 17 October 2026).
- 📖 **User Guide & Operational Flow**: **Sign in** opens the app at `/app/`. The site has no **Test my connection** link, in the header or the footer: both were removed on 17 October 2026, when the live host's built-in reflector was switched off for good. The self-test is reached only at `/selftest.html`; it measures DNS round trips to a public resolver, or the reach of any HTTPS host, and no longer offers the UDP test (real packet loss is measured in **Run diagnostic**, §4.2). The sun, moon and screen buttons in the header choose light, dark or match this device; the app's sign-in screen opens in the same choice. On a phone the header keeps the theme switch and **Sign in**.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `WEB-001` | Visitor | The site is current | Read the capabilities | Emailed or texted sign-in codes, IPv4 and IPv6, CSV into your systems, check-ins, testing from the engineer's own device. 🛑 **Must NOT** promise 23 languages, a hop-by-hop path or a line-speed test |
  | `WEB-002` | Visitor | Site and app are different documents | Open `/` and `/app/` | The marketing page and the app respectively — never the same document |
  | `WEB-003` | Visitor | The self-test runs in the browser | Open `/selftest.html`; open **Target**; run a resolver | **Target** lists *Cloudflare DNS (1.1.1.1)*, *Google Public DNS (8.8.8.8)* and **Another HTTPS host…**; round trips, jitter and loss for the resolver, measured from the visitor's connection. 🛑 **Must NOT** offer *PacketPulse network — UDP, like a call* (removed 17 October 2026; the UDP test is §4.2) |
  | `WEB-004` | Visitor | Light, dark or match this device | Choose **Light** on `/`; reload; open `/selftest.html`; then **Sign in**. Choose **Match this device** and switch the computer's own mode | Light at once, on both pages and on the app's sign-in screen; *Match this device* follows the computer as it changes. 🛑 **Must NOT** flash dark before a light page paints |
  | `WEB-005` | Visitor | The site fits a phone | Open `/` and `/selftest.html` at 360, 390 and 520 px wide and scroll to the foot of each | Logo, theme switch and **Sign in** on one line; the cards in one column; the footer links wrap. 🛑 **Must NOT** scroll sideways at any of them (the footer links and the cards' minimum width once pushed both pages wider than a phone) |
  | `WEB-006` | Visitor | Only the fields the method uses | Open `/selftest.html`; switch **Target** between a resolver and **Another HTTPS host…** | **Host or URL** only for another host; Samples, Interval and Timeout for both. 🛑 **Must NOT** show a hidden field (the page's CSS once overrode `hidden`) |
  | `WEB-007` | Visitor | The site does not advertise the self-test | Open `/` at full width and at 390 px; read the header, then scroll to the footer | The section links, the theme switch and **Sign in**; no *Test my connection* in the header or the footer. 🛑 **Must NOT** link `/selftest.html` from `/` (removed 17 October 2026) |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulseweb/index.html`, `packetpulseweb/selftest.html`; the probe engine's tests: `packetpulseweb/assets/packetpulse-probe.test.mjs` (part of `unit`); the UDP engine and its tests, `packetpulseweb/assets/packetpulse-udp.js` and `packetpulseweb/assets/packetpulse-udp.test.mjs` (§4.2), which no site page loads since 17 October 2026 but which stay, because the app runs a byte-identical copy (`packetpulseflutter/web/engine/packetpulse-udp.js`); change both together. The theme switch: `packetpulseweb/assets/packetpulse-theme.js`, which keeps the mode under the app's own appearance key; its tests, `packetpulseweb/assets/packetpulse-theme.test.mjs`, are part of `unit` too.
  - The deploy asserts `/` and `/app/` differ, so the nginx misroute that once served a blank app is caught by the deploy.

---

### 7.4 🎨 Branding

**Screen:** Configure → **Branding** · **Routes:** `GET /organisation/brand`, `PUT/DELETE /organisation/brand/{part}` · **Capability:** `staff_manage` to change; any member reads

- 🌟 **Commercial Presentation & Sales Pitch**: The customer's own logo heads every report their engineers attach to a ticket — the ticket's PDF — and their own icon sits in every one of their people's browser tabs. Another organisation on the same service never sees either.
- 📖 **User Guide & Operational Flow**: **Report logo**: **Upload image** (PNG, JPEG or GIF, up to 2400 by 1200 pixels). It is printed on the report's dark band before PacketPulse's name: white lettering suits it best, and a dark logo is printed on a white panel. **Browser icon**: up to 512 by 512, square works best; it replaces PacketPulse's tab icon for everyone in the organisation once they sign in, and PacketPulse's returns when they sign out. **Replace** and **Remove** each part on its own. Images are uploaded from PacketPulse in a web browser. Use only a logo your organisation has the right to use.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `BRAND-001` | Administrator | A logo heads every report | Upload a white-lettered PNG logo; export a ticket's PDF, a run's PDF and the month's availability PDF | The logo before *PacketPulse* on every page of all three, straight on the dark band |
  | `BRAND-002` | Administrator | A dark logo still shows | Upload a dark logo on a clear background; export a PDF | The logo on a white panel, readable. 🛑 **Must NOT** vanish into the band |
  | `BRAND-003` | Engineer | The tab wears the organisation's icon | Administrator uploads an icon; an engineer signs in; then signs out | The tab shows the organisation's icon after sign-in and PacketPulse's after sign-out |
  | `BRAND-004` | Administrator | What is not an image is refused | Upload an SVG, a text file renamed `.png`, and a 5000-pixel-wide PNG | Each refused under its own part with the reason; the stored logo unchanged |
  | `BRAND-005` | RIVAL engineer | A brand is the organisation's own | Sign in to RIVAL; export a PDF | No ACME logo or icon anywhere. 🛑 **Must NOT** show another organisation's brand |
  | `BRAND-006` | Engineer | Changing the brand needs `staff_manage` | As an engineer, `PUT /organisation/brand/logo` | 403; no **Branding** in the rail |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/orgbrand/`: `Normalise` checks the declared size before decoding (so a small file claiming a vast canvas costs nothing) and redraws every upload as a plain RGBA PNG; `DrawLogo` prints it, and a logo the PDF writer cannot read costs the report its logo, never the report. Stored in `organisation_brand` (migration 0029), one row per organisation.
  - Reports get the logo through `SetReportBranding` on the diagnostic and ping services. The tab icon: `packetpulseflutter/lib/common/services/PacketPulseFavicon.dart`, applied by the shell after sign-in and removed when it closes.
  - Coverage: `packetpulsego/pkg/common/orgbrand/`, `packetpulsego/pkg/staffmicroservice/staffapp/`, `packetpulsego/pkg/pingmicroservice/pingreport/`, `packetpulseflutter/test/brand_screen_test.dart`, `packetpulseflutter/test/shell_branding_test.dart`.

---
