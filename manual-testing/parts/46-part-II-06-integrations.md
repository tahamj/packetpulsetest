## Group 6 — Integrations

How results reach the operator's own systems, and how its directory signs people in.

---

### 6.1 🔌 Results API keys

**Screen:** Configure → **Results API keys** · **Routes:** `GET /apikey/list`, `POST /apikey/add`, `DELETE /apikey/{credentialId}`; with a key: `GET /result/bytt/{ttNumber}`, `GET /result/export.csv` · **Capability:** `apikey_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: The ticketing system pulls the evidence itself — one ticket as JSON, or a month of results as CSV — with a key scoped to the operator's own organisation, readable even if a renewal is late. The key is shown once and stored only as a hash.
- 📖 **User Guide & Operational Flow**: **Issue key** with a label and an optional expiry; copy it — it is never shown again. Your system sends it as `Authorization: Bearer <key>`. `GET /api/v1/result/export.csv?from=…&to=…` takes RFC 3339 times, at most 31 days apart; with no `from`, the day before `to` (default now).
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `API-001` | Administrator | A key is shown once | Issue a key; reload the screen | The full key appeared once; afterwards only its prefix, label, expiry and last use |
  | `API-002` | Integrator | One ticket as JSON | `curl -H "Authorization: Bearer <key>" https://…/api/v1/result/bytt/TT-MAN-001` | The ticket with every result, including `ip_version` and `report_only` |
  | `API-003` | Integrator | A period as CSV | `…/result/export.csv?from=2026-10-01T00:00:00Z&to=2026-10-02T00:00:00Z` | `text/csv`, the contracted header, one row per result of the diagnostics that finished in that window |
  | `API-004` | Integrator | Bad periods are refused, not guessed | `from` 40 days before `to`; `to=tomorrow`; `from` after `to` | 422 (*at most 31 days*), 400 (unreadable time), 422. 🛑 **Must NOT** answer with a partial file |
  | `API-005` | Integrator | The wrong credential is refused | No header; a person's session token; a revoked key | 401 each |
  | `API-006` | RIVAL integrator | A key reads only its own organisation | RIVAL's key for ACME's TT, and a period covering ACME's tickets | 404; a CSV with none of ACME's rows |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pinglego/pkg/common/apikeyauth/`; the organisation comes from the key, never from the request.
  - The bulk pull **streams**; a failure after rows have gone out aborts the response rather than ending a short file a system would take as whole.
  - Coverage: `pingletest/golang/apicontract/results_api_test.go`, `pingletest/golang/apicontract/result_export_test.go`, `pinglego/pkg/diagnosticmicroservice/diagnosticapp/`.

---

### 6.2 📤 Result export

**Screen:** Configure → **Result export** · **Routes:** `GET/PUT /export/target`, `POST /export/target/test` · **Capability:** `apikey_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Results delivered to the operator's own SFTP or FTPS server as CSV, every interval it chooses — no integration project. Nothing is lost: a failed delivery is retried with the same period, and each row carries a `result_id` so a file that arrives twice is recognised. The server's key is confirmed before anything is sent, and a saved password is only ever sent to the server it was entered for.
- 📖 **User Guide & Operational Flow**:
  - **Protocol:** SFTP (recommended), FTP over TLS, or FTP — with a clear warning that plain FTP is unencrypted.
  - **Server, Port, Username, Password** (or for SFTP a private key). Saved credentials are never shown again; leave the field empty to keep them, or tick *Remove the saved password*.
  - **Test connection** signs in, writes and removes a small file. For SFTP it shows the server's key: check it with whoever runs the server, then **Trust this key** and **Save**.
  - **Folder** and **Deliver every (minutes)** — 15 to 1440.
  - The status card says how the last delivery went: *Sent <file> with N results*, *Nothing new to send*, or *Not delivered: <why>*.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `EXP-001` | Administrator | A target nobody set up | Open the screen | *Not delivered yet. The first delivery is one interval after saving.*; SFTP, port 22, 60 minutes; switched off |
  | `EXP-002` | Administrator | Confirm the server's key | Fill in an SFTP server; **Test connection** | *Trust this server?* showing `SHA256:…`. **Cancel** → *Not trusted yet*; test again → **Trust this key** → *Trusted: SHA256:…*; **Save** |
  | `EXP-003` | Administrator | Switching on needs everything a delivery needs | Switch on with no server, no username, no password and no trusted key | Refused against each field |
  | `EXP-004` | Administrator | A changed key is a warning, not a detail | Re-key the SFTP server (or point the name at another); **Test connection** | *The server's key has changed* with the new key to check; the trusted key stays until you trust the new one. 🛑 **Must NOT** deliver to a server showing a different key |
  | `EXP-005` | Administrator | A saved password goes nowhere new | With a saved password, change the server (or port, username or protocol) and test or save | The password field becomes required again (*Enter the password again…*). 🛑 **Must NOT** sign in to the new server with the saved password |
  | `EXP-006` | Administrator | A delivery lands whole | Switched on, every 15 minutes, a diagnostic run after saving; wait for the interval | `pingle-results-<from>-<to>.csv` in the folder (written as `.part`, renamed when whole); the card says *Sent … with N results* |
  | `EXP-007` | Administrator | A quiet period sends nothing | No diagnostics in an interval | No file; *Nothing new to send* |
  | `EXP-008` | Administrator | A failure is retried, nothing lost | Change the server's password; wait for a delivery; restore it; wait again | First: *Not delivered: The server refused the sign-in…*; the next delivery covers the same results. 🛑 **Must NOT** skip the failed period |
  | `EXP-009` | Administrator | Pingle's own host is not a destination | Server `127.0.0.1`, `localhost`, `169.254.169.254` or `::1`; **Test connection** | Refused against **Server**: *Pingle may not connect to that address.* |
  | `EXP-010` | Administrator | Plain FTP is warned about | Choose FTP | The unencrypted warning; port moves to 21; the private key and server key disappear |
  | `EXP-011` | Administrator | FTPS with a private authority | FTPS to a server whose certificate your own CA signed, without and then with that CA | Without: *certificate is not trusted*; with it pasted: test passes |
  | `EXP-012` | Tester | Credentials never come back | `GET /api/v1/export/target` | `has_password`, `has_private_key` — never the password, key or sealed text |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pinglego/pkg/exportmicroservice/` (target, test, runner each minute); transport: `pinglego/pkg/common/filedrop/FileDrop.go` (SFTP via `pkg/sftp`, FTP/FTPS via `jlaffaye/ftp`).
  - 🔒 Every connection — the FTP **data** connection included — is made by `probeguard.Dialer`, whose `Control` hook checks the resolved address before connecting. A dial *function* would have made the FTP library send FTPS data unencrypted.
  - The watermark (`exported_through`) moves only on a delivery, and only from where the run found it; a run catches up a day per file; the last two minutes are left to settle.
  - `EXPORT_ALLOW_LOOPBACK` permits a server on Pingle's own host for development; production refuses to boot with it.
  - Coverage: `pinglego/pkg/common/filedrop/` (in-process SFTP and FTP/FTPS servers), `pinglego/pkg/exportmicroservice/**`, `pingleflutter/test/result_export_screen_test.dart`, `pingletest/golang/apicontract/result_export_test.go`.

---

### 6.3 🏢 Directory

**Screen:** Configure → **Directory** · **Routes:** `GET/PUT /ldap/config`, `POST /ldap/config/test`, `/ldap/groupmap/*` · **Capability:** `ldap_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Staff sign in with their existing network password over LDAP or Active Directory, and their directory group decides their Pingle role — after the second step, so a directory password alone still opens nothing. The owner always keeps a local password, so a directory outage cannot lock out the person who fixes it.
- 📖 **User Guide & Operational Flow**: Host and port (389 LDAP/StartTLS, 636 LDAPS), encryption, the CA certificate of your own authority, a read-only bind account, the base DN; then **Test connection** before enabling. **Group mappings** give each directory group a role.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `LDAP-001` | Administrator | Connect and test | Point at a test directory with its CA certificate; **Test connection** | Success; a wrong CA → certificate error; skipping verification is warned against |
  | `LDAP-002` | Directory user | The group decides the role — after the second step | Map group `noc` → *NOC Engineer*; a directory user in `noc` signs in | Password checked by the directory; second step asked; only then the role applied |
  | `LDAP-003` | Owner | The owner keeps a local password | Directory switched on and unreachable; owner signs in with their local password | Signed in. Others: *Your organisation's directory server could not be reached. Please retry.* |
  | `LDAP-004` | Tester | An outage is not a failed sign-in | Directory down; an administrator signs in | No *Sign-in failed* entry in Activity — nobody got the password wrong |
- ⚙️ **Developer Guide & Release Confidence**:
  - `pinglego/pkg/common/ldapclient/`, `pinglego/pkg/ldapmicroservice/`; the bind password is sealed like every secret; the directory is dialled through the directory policy (private yes; loopback only with `LDAP_ALLOW_LOOPBACK`, refused in production).
  - Coverage: `pinglego/pkg/common/ldaptest/` (an in-process directory), `pingleflutter/test/ldap_settings_screen_test.dart`.

---
