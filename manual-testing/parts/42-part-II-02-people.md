## Group 2 — People

Who is in the organisation, what each may do, who is signed in, and the record of what they did.

---

### 2.1 👥 Staff

**Screen:** Administer → **Staff** · **Routes:** `GET /staff/list`, `PUT/DELETE /staff/{staffId}`, `PUT /staff/{staffId}/secondfactor`, `POST /staff/{staffId}/secondfactor/reset`, `POST /user/add` · **Capability:** `staff_manage` (adding people: `user_manage`)

- 🌟 **Commercial Presentation & Sales Pitch**: One screen to add a colleague, give them a role, decide how they sign in and whether their position is recorded — and to stop them instantly. Disabling someone signs them out at once, not when a token happens to expire, which is the property an auditor asks about first.
- 📖 **User Guide & Operational Flow**: **Add** a person with email, a starting password and a role. Edit to change their role, staff code (unique in the organisation), department and designation (labels only — authority comes from the role). **How they sign in** sets the second step (authenticator or text, with their mobile number) and **Record location at sign-in and sign-out**. **Reset authenticator** is for a lost phone. Turning **Enabled** off signs them out and blocks sign-in.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `STF-001` | Administrator | Add a colleague | **Add** with role *NOC Engineer* | They appear with the role; their first sign-in sets up an authenticator (`AUTH-001`) |
  | `STF-002` | Administrator | Staff codes are unique inside the organisation | Give two people the same staff code; then give RIVAL's person the same code | First: refused (`conflict`). RIVAL: allowed — uniqueness is per organisation |
  | `STF-003` | Administrator | Disabling signs out at once | Engineer signed in on another browser; untick **Enabled** and save | The engineer's next action returns them to sign-in; signing in is refused. 🛑 **Must NOT** leave the session working until it expires |
  | `STF-004` | Administrator | How they sign in, saved whole | Set *Text message* with mobile `+919876543210` and *Record location* off; save. Then edit only the department and save | Second save keeps the method, number and location setting. 🛑 **Must NOT** reset what the department edit did not show |
  | `STF-005` | Administrator | A bad mobile number is refused | Set *Text message* with `98765 43210` | Refused: international form needed (E.164) |
  | `STF-006` | Administrator | Changing a role signs the person out | Engineer signed in; change their role to *Viewer* | Their next action returns them to sign-in; signed in again, they see only Viewer's screens |
  | `STF-007` | Viewer | Without `staff_manage` there is no Staff screen | Sign in as a Viewer; call `GET /staff/list` | No **Staff** in the rail; the API answers 403 |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pinglego/pkg/staffmicroservice/staffapp/StaffRouteHandler.go`; the second-step settings are a **separate** route (`PUT /staff/{staffId}/secondfactor`) because the whole-row staff update would otherwise wipe them; it requires `second_factor`, `phone_number` and `capture_location`.
  - 🔒 Every write is scoped by the caller's organisation; editing RIVAL's `staffId` answers 404.
  - Coverage: `pinglego/pkg/staffmicroservice/**`, `pingleflutter/test/staff_screen_test.dart`, the `acl` and `tenancy` suites.

---

### 2.2 🧩 Roles and permissions

**Screen:** Administer → **Permission matrix** · **Routes:** `GET/POST /staff/role/*`, `PUT /staff/{staffId}/access`, `GET /user/capability/list` · **Capability:** `acl_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Seventeen plain-language capabilities, three built-in roles (Administrator 17, NOC Engineer 8, Viewer 5) and as many of the operator's own as it likes — plus a per-person *allow* or *deny* for the exception that does not deserve a role. Changes take effect on the next request, not the next sign-in.
- 📖 **User Guide & Operational Flow**: The matrix lists roles across and capabilities down. **Add a role**, tick what it may do, save — the whole set is saved, so nothing is left to an invisible default. Built-in roles are read-only. A role somebody holds cannot be deleted. Per person: *Inherit*, *Allow* or *Deny* each capability.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `ACL-001` | Administrator | Built-in roles cannot be changed | Try to edit *Viewer* | Not editable; `PUT /staff/role/{viewerId}` answers 403 or 409. 🛑 **Must NOT** change a role every organisation shares |
  | `ACL-002` | Administrator | A custom role grants exactly what is ticked | Add *Shift lead* with `diagnostic_run` and `report_view` only; give it to an engineer | They can run diagnostics and see the dashboard; **Staff**, **Monitoring** edits and **API keys** are absent. Each refused route answers 403 |
  | `ACL-003` | Administrator | A role in use cannot be deleted | Delete *Shift lead* while someone holds it | Refused, naming that people hold it |
  | `ACL-004` | Administrator | A per-person deny wins over the role | Deny `report_export` to one engineer | That engineer has no **Export PDF**/**Export CSV**; `GET /diagnostic/{id}/report.csv` answers 403. Other engineers unaffected |
  | `ACL-005` | Tester | Every role, both directions | Run `./pingletest.sh acl` | Passes — each role is refused what it lacks and allowed what it holds |
- ⚙️ **Developer Guide & Release Confidence**:
  - Capabilities: `pinglego/pkg/common/pingleaccess/PingleAccessCategory.go` (Appendix B); guards: `pingleaccess.RequireCapability` on each route (Appendix A).
  - 🔒 Authority is read fresh on every request; a role change revokes the person's sessions.
  - Coverage: `pingletest/golang/aclconformance/acl_conformance_test.go` (the full matrix), `pingleflutter/test/permission_matrix_screen_test.dart`.

---

### 2.3 💺 Sessions and seats

**Screen:** Administer → **Sessions and seats** · **Routes:** `GET /staff/session/list`, `DELETE /staff/session/{sessionId}`, `GET /staff/session/organisation`, `DELETE /staff/session/organisation/{sessionId}`

- 🌟 **Commercial Presentation & Sales Pitch**: Floating licences: seats count people signed in *now*, not accounts. Forty engineers share twenty seats across shifts, and an administrator can free a seat left signed in on a desk without disabling anyone.
- 📖 **User Guide & Operational Flow**: Your own sessions list every device you are signed in on, with its address (IPv4 or IPv6, in full in the tooltip). Administrators also see everyone signed in, and **Free this seat** ends one session.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SEAT-001` | Engineer | The next person over the limit is refused | Licence of 2 seats (Platform → seats); two people signed in; a third signs in | Third refused (`seat_limit_reached`) with guidance to ask an administrator. 🛑 **Must NOT** sign the third person in |
  | `SEAT-002` | Engineer | One person, two devices, one seat | Sign in on a laptop and a phone | Seat count rises by one |
  | `SEAT-003` | Engineer | A second tab never locks you out | All seats full; open a second tab and sign in as yourself | Allowed. 🛑 **Must NOT** refuse you for your own seat |
  | `SEAT-004` | Administrator | Free a seat | **Free this seat** on a colleague's session | Their next action returns them to sign-in; Activity records who freed whose seat |
  | `SEAT-005` | Engineer | An IPv6 address shows whole | Sign in over IPv6 (or seed a session with `2401:4900:1c2a:8e1f::1`) | One line, ellipsised, full address in the tooltip and selectable. 🛑 **Must NOT** wrap across lines or overflow at phone width |
- ⚙️ **Developer Guide & Release Confidence**:
  - Seats are checked at the password step **and** when the session opens (`enforceSeatLimit`), counted per person over live, unrevoked sessions.
  - Coverage: `pingletest/golang/tenancyassignment/`, `pingleflutter/test/session_screen_test.dart`.

---

### 2.4 📍 Check-ins

**Screen:** Administer → **Check-ins** · **Route:** `GET /staff/checkin/list` · **Capability:** `staff_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Where every field engineer was when they signed in and out — position, accuracy, street address and a map link — on a list of its own, separate from the administrators' audit trail. Proof of attendance without a separate app.
- 📖 **User Guide & Operational Flow**: One row per session of a person whose location is recorded and who is **not** an administrator: who, signed in (time, place, *Open in Maps*), signed out (time, place, or *still signed in* / *expired* / *ended*), address and device. Filter by date range; the default is the last seven days.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `CHK-001` | Engineer | A sign-in records where | *Record location* on; sign in and allow location | **Check-ins** shows the time, the position ± accuracy and — a moment later — the address; *Open in Maps* opens the spot |
  | `CHK-002` | Engineer | A sign-out records where it ended | Sign out, allowing location | The same row gains the sign-out time and place |
  | `CHK-003` | Engineer | A session left to expire says so | Sign in; let the session expire (or shorten `JWT_TTL`) | *Expired*, not a sign-out time |
  | `CHK-004` | Engineer | A refusal is recorded as one | Deny the location prompt (rule off) | *Location refused* — no position, no map link |
  | `CHK-005` | Administrator | Administrators are not on Check-ins | The owner signs in with location | No Check-ins row; the sign-in is in **Activity** (`AUD-001`) |
  | `CHK-006` | RIVAL administrator | Another organisation's check-ins are never shown | RIVAL opens Check-ins | Only RIVAL's people. 🛑 **Must NOT** show ACME's |
- ⚙️ **Developer Guide & Release Confidence**:
  - `session_checkin`, one row per session, keyed by `staff_session.session_id` (`pinglego/pkg/common/dbclient/migrations/0020_2026_10_03_session_checkin.sql`); coordinates only with status `captured`, checked by `CHECK`s.
  - Addresses: a background worker (`pinglego/pkg/common/geocode/Geocode.go`) — Google with `GOOGLE_MAPS_API_KEY`, else OpenStreetMap Nominatim at one request a second. `REVERSE_GEOCODING=off` disables it.
  - Coverage: `pinglego/pkg/staffmicroservice/staffservice/`, `pingleflutter/test/checkin_screen_test.dart`, `pingletest/golang/tenancyisolation/`.

---

### 2.5 🧾 Activity

**Screen:** Administer → **Activity** · **Routes:** `GET /auditlog/list`, `GET /auditlog/verify` · **Capability:** `auditlog_view`

- 🌟 **Commercial Presentation & Sales Pitch**: A tamper-evident record of every change and every administrator sign-in — device, browser, address, second step and where they were — and of every **refused** attempt on an administrator's account, with why. Hash-chained: altering or deleting one entry breaks every entry after it, and *Verify chain* names the first. Positions sit beside entries, never inside their hash, so the record stays verifiable *and* erasable under a retention policy.
- 📖 **User Guide & Operational Flow**: Newest first. Select a row for its details: role, second step, device, **browser** ("Chrome 128 on macOS", with the raw string beneath), place with *Open in Maps*, session id and entry hash. **Sign-ins and sign-outs only** narrows the list. A *Sign-in failed* row shows **Why**: wrong password, wrong code, or locked after too many wrong codes.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `AUD-001` | Administrator | An administrator's sign-in is recorded in full | Sign in as the owner from Chrome, allowing location; open **Activity** | *Signed in* by the owner: role, second step, device, *Chrome NN on <system>*, the IP (v4 or v6), the place with a map link, the session id |
  | `AUD-002` | Administrator | And their sign-out | Sign out with location; sign back in | A *Signed out* row with its own place |
  | `AUD-003` | Tester | A wrong password on an administrator's account is recorded | From another browser, sign in as the owner with a wrong password | *Sign-in failed* — Why: *Wrong password*, from that address. 🛑 **Must NOT** carry a position in the entry |
  | `AUD-004` | Tester | A wrong code, and the lock | Owner's right password, then wrong codes until locked | *Sign-in failed* rows: *Wrong code* (naming the authenticator) until the 10th, which says *Locked after too many wrong codes*; the next right password is recorded as locked too |
  | `AUD-005` | Tester | Not everyone's failures are activity | An engineer's wrong code; an unknown email's wrong password | Neither appears. 🛑 **Must NOT** record a field engineer's mistype, or list addresses people guessed |
  | `AUD-006` | Tester | The record cannot be flooded | 25 wrong passwords at the owner's address within an hour | At most **20** *Sign-in failed* rows for that account in the hour; the rest go to the server log |
  | `AUD-007` | Administrator | Verify the chain | *Verify chain* | *Intact*, with how many entries were checked and the head hash |
  | `AUD-008` | Tester | Tampering is caught (disposable stack only) | `UPDATE auditlog_activity SET actor_email='x' WHERE activity_id = <some id>`; *Verify chain* | *Broken* at exactly that entry, with the reason |
  | `AUD-009` | Engineer | A refused change leaves no entry | As a Viewer, try to add a site (403) | No entry. Only what happened is recorded |
  | `AUD-010` | Administrator | Exports and integrations are on the record | Download a ticket's CSV; save the result export; test its connection | Rows: *Exported* (diagnostic), *Changed* and *Tested* (result_export) |
  | `AUD-011` | RIVAL administrator | Another organisation's trail is never shown | RIVAL opens Activity | Only RIVAL's entries |
- ⚙️ **Developer Guide & Release Confidence**:
  - Registry: `pinglego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go` — every mutating route is there or excluded with a reason, and `TestEveryMutationRouteIsRegistered` fails a route added without either.
  - Sign-ins are recorded by UserMS itself (`recordSignIn`, `recordSignOut`, `recordFailedSignIn` in `pinglego/pkg/usermicroservice/userservice/UserSignInSteps.go`) because the routes are public; failed ones are capped by `maxFailedEntriesPerHour = 20` per account.
  - 🔒 The list joins positions from `session_checkin` by session id; the chain's `details` never hold coordinates.
  - Coverage: `pingletest/golang/auditconformance/`, `pinglego/pkg/common/auditlog/`, `pingleflutter/test/audit_log_screen_test.dart`, `pingleflutter/test/user_agent_test.dart`.

---
