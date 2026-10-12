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
  - `packetpulsego/pkg/common/apikeyauth/`; the organisation comes from the key, never from the request.
  - The bulk pull **streams**; a failure after rows have gone out aborts the response rather than ending a short file a system would take as whole.
  - Coverage: `packetpulsetest/golang/apicontract/results_api_test.go`, `packetpulsetest/golang/apicontract/result_export_test.go`, `packetpulsego/pkg/diagnosticmicroservice/diagnosticapp/`.

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
  | `EXP-006` | Administrator | A delivery lands whole | Switched on, every 15 minutes, a diagnostic run after saving; wait for the interval | `packetpulse-results-<from>-<to>.csv` in the folder (written as `.part`, renamed when whole); the card says *Sent … with N results* |
  | `EXP-007` | Administrator | A quiet period sends nothing | No diagnostics in an interval | No file; *Nothing new to send* |
  | `EXP-008` | Administrator | A failure is retried, nothing lost | Change the server's password; wait for a delivery; restore it; wait again | First: *Not delivered: The server refused the sign-in…*; the next delivery covers the same results. 🛑 **Must NOT** skip the failed period |
  | `EXP-009` | Administrator | PacketPulse's own host is not a destination | Server `127.0.0.1`, `localhost`, `169.254.169.254` or `::1`; **Test connection** | Refused against **Server**: *PacketPulse may not connect to that address.* |
  | `EXP-010` | Administrator | Plain FTP is warned about | Choose FTP | The unencrypted warning; port moves to 21; the private key and server key disappear |
  | `EXP-011` | Administrator | FTPS with a private authority | FTPS to a server whose certificate your own CA signed, without and then with that CA | Without: *certificate is not trusted*; with it pasted: test passes |
  | `EXP-012` | Tester | Credentials never come back | `GET /api/v1/export/target` | `has_password`, `has_private_key` — never the password, key or sealed text |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/exportmicroservice/` (target, test, runner each minute); transport: `packetpulsego/pkg/common/filedrop/FileDrop.go` (SFTP via `pkg/sftp`, FTP/FTPS via `jlaffaye/ftp`).
  - 🔒 Every connection — the FTP **data** connection included — is made by `probeguard.Dialer`, whose `Control` hook checks the resolved address before connecting. A dial *function* would have made the FTP library send FTPS data unencrypted.
  - The watermark (`exported_through`) moves only on a delivery, and only from where the run found it; a run catches up a day per file; the last two minutes are left to settle.
  - `EXPORT_ALLOW_LOOPBACK` permits a server on PacketPulse's own host for development; production refuses to boot with it.
  - Coverage: `packetpulsego/pkg/common/filedrop/` (in-process SFTP and FTP/FTPS servers), `packetpulsego/pkg/exportmicroservice/**`, `packetpulseflutter/test/result_export_screen_test.dart`, `packetpulsetest/golang/apicontract/result_export_test.go`.

---

### 6.3 🏢 Directory

**Screen:** Configure → **Directory** · **Routes:** `GET/PUT /ldap/config`, `POST /ldap/config/test`, `/ldap/groupmap/*` · **Capability:** `ldap_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: Staff sign in with their existing network password over LDAP or Active Directory, and their directory group decides their PacketPulse role — after the second step, so a directory password alone still opens nothing. The owner always keeps a local password, so a directory outage cannot lock out the person who fixes it.
- 📖 **User Guide & Operational Flow**: Host and port (389 LDAP/StartTLS, 636 LDAPS), encryption, the CA certificate of your own authority, a read-only bind account, the base DN; then **Test connection** before enabling. **Group mappings** give each directory group a role.
  - **Who is looked up:** the person's PacketPulse email address up to the `@` is what replaces each `%s` in the **User filter**. Where directory account names match email addresses, `(sAMAccountName=%s)` (Active Directory) or `(uid=%s)` (OpenLDAP, FreeIPA) works; where they do not (staff numbers), search by mail: `(mail=%s@corp.example.com)`.
  - **Group filter:** each `%s` is the person's DN; for groups that list members by user name, `%u` is their user name — `(&(objectClass=posixGroup)(memberUid=%u))`. One of the two is required — without it every person would be in every group. Active Directory's `(&(objectClass=group)(member:1.2.840.113556.1.4.1941:=%s))` also follows nested groups. A filter may use a placeholder more than once: `(|(uid=%s)(mail=%s@corp.example.com))`.
  - Settings for Active Directory, OpenLDAP, OpenLDAP with `groupOfUniqueNames` or `posixGroup`, and FreeIPA are each proven end to end, using only the screen's fields, by `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapDirectoryProducts_test.go`.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `LDAP-001` | Administrator | Connect and test | Point at a test directory with its CA certificate; **Test connection** | Success; a wrong CA → certificate error; skipping verification is warned against |
  | `LDAP-002` | Directory user | The group decides the role — after the second step | Map group `noc` → *NOC Engineer*; a directory user in `noc` signs in | Password checked by the directory; second step asked; only then the role applied |
  | `LDAP-003` | Owner | The owner keeps a local password | Directory switched on and unreachable; owner signs in with their local password | Signed in. Others: *Your organisation's directory server could not be reached. Please retry.* |
  | `LDAP-004` | Tester | An outage is not a failed sign-in | Directory down; an administrator signs in | No *Sign-in failed* entry in Activity — nobody got the password wrong |
  | `LDAP-005` | Administrator | A group filter must name the person | Group filter `(objectClass=group)`; **Save** | Refused against **Group filter**: *The group filter needs a %s where the person's DN goes, or a %u where their user name goes.* 🛑 **Must NOT** save a filter that puts everyone in every group |
  | `LDAP-006` | Directory user | Nested Active Directory groups grant the role | Map `PacketPulse NOC` → *NOC Engineer*; the person is in `NOC Shift A`, itself a member of `PacketPulse NOC`; group filter `(&(objectClass=group)(member:1.2.840.113556.1.4.1941:=%s))`; sign in | *NOC Engineer* after the second step |
  | `LDAP-007` | Directory user | Account names that are not email addresses | AD account `E10442`, PacketPulse email `asha.rao@corp.example.com`; user filter `(&(objectCategory=person)(objectClass=user)(mail=%s@corp.example.com))` | Signed in with the directory password. With `(sAMAccountName=%s)` instead: *incorrect email address or password* |
  | `LDAP-008` | Administrator | A silent firewall does not hang the screen | Point **Host/Port** at an address that accepts the connection and never answers (a firewall that drops after the handshake, or `nc -lk 3890` on a test host); StartTLS; **Test connection** | Within about the connection timeout (10 s): *Could not reach a directory server at that host and port. Check the hostname, the port, and the encryption mode.* 🛑 **Must NOT** spin indefinitely, and **must NOT** say *The directory refused the service account* |
  | `LDAP-009` | Directory user | A person in hundreds of groups keeps them all | In a large AD, a person in several hundred groups, one of them mapped; sign in | The mapped role. 🛑 **Must NOT** fall back to the default role because the person has many groups |
  | `LDAP-010` | Tester | A filter that matches several people is named, not guessed | User filter `(uid=a*%s)`; a person signs in | The person is told *Your organisation's directory server could not be reached. Please retry.*; the server log reads *the user filter matched more than one entry*. 🛑 **Must NOT** sign in as whichever entry came first |
  | `LDAP-011` | Directory user | Groups that list members by user name | OpenLDAP `posixGroup` `netops` with `memberUid: asha.rao`, mapped to *NOC Engineer*; group filter `(&(objectClass=posixGroup)(memberUid=%u))`; `asha.rao` signs in | *NOC Engineer*. With `%s` instead of `%u`: no group, the role unchanged |
  | `LDAP-012` | Directory user | A filter that names the person twice | User filter `(&(objectClass=person)(\|(uid=%s)(mail=%s@corp.example.com)))`; a person whose `uid` is a staff number but whose mail is their PacketPulse address signs in | Signed in. 🛑 **Must NOT** fail because only the first `%s` was filled |
  | `LDAP-013` | Tester | The recorded test result says nothing about the network | **Test connection** against a closed port, then against a host that accepts and never answers; after each, `GET /api/v1/ldap/config` | `last_test_result` is the same sentence both times — *Could not reach a directory server at that host and port…* 🛑 **Must NOT** contain *refused*, *timed out*, *dial* or the port — that difference is a port scan |
- ⚙️ **Developer Guide & Release Confidence**:
  - `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/`; the bind password is sealed like every secret; the directory is dialled through the directory policy (private yes; loopback only with `LDAP_ALLOW_LOOPBACK`, refused in production).
  - The connection is set up — LDAPS handshake, or StartTLS and its handshake — inside one deadline, the connection timeout; go-ldap puts none on a StartTLS handshake, so a middlebox that accepted the connection and went quiet used to hold a sign-in open indefinitely.
  - Only `invalidCredentials` (49) is a wrong password. A timeout, a dropped connection, or a domain controller answering `busy` / `unavailable` is *could not be reached* — for the person signing in and for **Test connection** — so an outage is never recorded as a failed sign-in.
  - 🔒 **Test connection** answers, and records as `last_test_result`, one authored sentence per kind of failure (`ConnectionTestMessage` in `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapService.go`), never the raw error. Until October 2026 the raw error was recorded and `GET /ldap/config` returned it, so refused, timed out and the port could still be read. A result stored before then is shown as the generic failure.
  - Filters are filled by `fillFilter` in `packetpulsego/pkg/common/ldapclient/LdapClient.go`: every `%s` (not only the first, as `fmt.Sprintf` did), `%u` in a group filter, `%%` as a percent sign; each value escaped.
  - A user search answered with `sizeLimitExceeded` is the ambiguity (`ErrAmbiguousUser`), not an outage. Groups are read in pages of 500 with no cap on the count; a cap of 100 used to lose every group past it. Directory failures at sign-in are logged with their reason (*directory sign-in could not be completed*).
  - Coverage: `packetpulsego/pkg/common/ldaptest/` (an in-process directory that answers as OpenLDAP or Active Directory: DN matching, AD's matching rules and sign-in names, size limits and paging, per-connection identity and access rules, unavailable binds, stalled StartTLS, silent hosts), `packetpulsego/pkg/common/ldapclient/LdapCustomerDirectory_test.go`, `packetpulsego/pkg/ldapmicroservice/ldapservice/LdapDirectoryProducts_test.go`, `packetpulsego/pkg/usermicroservice/userservice/UserDirectorySignIn_test.go`, `packetpulseflutter/test/ldap_settings_screen_test.dart`.

---
