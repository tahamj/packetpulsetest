# Part IV — Non-Functional Matrix

What every screen owes every person, whatever the feature.

## 4.1 Form factors

The app runs on the web (the live build), macOS, Windows, Android and iOS from one codebase. Below 600 px a list row reads as two lines; below 1000 px the navigation rail drops its group names.

| ID | Check | Expected |
|---|---|---|
| `NFR-001` | Every screen at 390 × 844 (a phone) | No horizontal page scroll; no overflow stripes; tables scroll within their card |
| `NFR-002` | Long values: a 39-character IPv6 address, a 60-character site name, a Devanagari name | One line, ellipsised, the whole value in the tooltip and selectable. 🛑 **Must NOT** break the row or truncate silently |
| `NFR-003` | The rail at 1600, 1000 and 600 px | Grouped with names; grouped with dividers; the drawer |

## 4.2 States every screen has

| ID | State | Expected |
|---|---|---|
| `NFR-004` | Loading | A skeleton, never a blank or a spinner over an empty frame |
| `NFR-005` | Empty | The designed empty state with what to do next |
| `NFR-006` | Failed | The failure with **Try again** when the server says it is worth retrying. 🛑 **Must NOT** show the empty state — "No sites yet" under a refusal reads as data loss |
| `NFR-007` | 403 | What the server says to do: a lapsed licence names who can renew; a missing capability says to ask an administrator |
| `NFR-008` | Forms | Required fields end with ` *`; a refused save names each field; a form whose load failed offers no **Save** |

## 4.3 Appearance and accessibility

| ID | Check | Expected |
|---|---|---|
| `NFR-009` | Each of the five themes in light and dark | Text contrast passes; status colours unchanged in meaning |
| `NFR-010` | Reduce motion on (app setting or the device's) | No transitions |
| `NFR-011` | Status by colour alone | Every pill carries an icon as well as a colour |
| `NFR-012` | Every user-visible sentence | From the catalogue: no English hard-coded in a widget, no raw key shown |

## 4.4 Security

| ID | Check | Expected |
|---|---|---|
| `NFR-013` | Tenancy, for every id-taking route | RIVAL's id answers 404 or an empty list — `./packetpulsetest.sh tenancy` |
| `NFR-014` | The destination policy, everywhere PacketPulse connects out | Sweeps, webhooks, the directory and the result export all refuse loopback, link-local and cloud metadata |
| `NFR-015` | Error responses | No SQLSTATE, driver names, file paths or a remote server's raw reply in any error envelope |
| `NFR-016` | Secrets at rest | SMS tokens, bind passwords and export credentials sealed with the server's secret box; API keys stored as hashes; never returned |
| `NFR-017` | CSV injection | Cells starting `=`, `+`, `-`, `@`, tab or carriage return are written as text |
| `NFR-018` | Production configuration | Production refuses `OTP_OUTBOX_FILE`, a `LICENCE_PUBLIC_KEY` override, `LDAP_ALLOW_LOOPBACK`, `EXPORT_ALLOW_LOOPBACK` and open CORS, and will not boot without an SMTP account |
| `NFR-019` | Text the database cannot hold | A NUL character (`\u0000`) in any field of a JSON request - a site name, a note, a password - is refused with *Text in the request cannot contain a NUL character* (400). 🛑 **Must NOT** answer 500 *try again* for input that can never be stored |

## 4.5 Performance — the load suite

`./packetpulsetest.sh load` provisions one licensed organisation — the owner and nineteen engineers, each signing in with a code read from the server's outbox, and 200 sites on TEST-NET-3 — and measures it. It also runs nightly (`.github/workflows/packetpulse-load.yml`). Budgets are 95th percentiles; the measured figures are from a developer laptop against a local server.

| Scenario | Measured p95 | Budget |
|---|---|---|
| Sign-in, password step — all 20 at once | 750 ms | 3 s |
| Sign-in, code step — all 20 at once | 16 ms | 1 s |
| Reads — list, sites, dashboard, me; 20 people × 10 | 9 ms | 750 ms |
| Sweep of 200 sites (measured when failures were also traced, before path analysis was removed) | 91 s | 120 s |
| Diagnostic of 5 sites — 10 people at once | 4.1 s | 20 s |
| Results API by ticket (200 results) — 200 pulls, 10 at a time | 16 ms | 1 s |
| CSV export of the whole run (250 rows) | 2 ms | 10 s |

🔒 Absolute whatever the machine: no 5xx; the 200-site sweep stores all 200 results; the CSV holds every stored result once. With the old bug restored (saving under the sweep's deadline) the sweep answers 500 and the suite goes red.

---
