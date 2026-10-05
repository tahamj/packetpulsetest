# Part II — Feature Chapters

One chapter per area of the product, in the order a new customer meets them. Every chapter carries all four lenses — **🌟 Commercial**, **📖 User Guide**, **🧪 Testing Playbook** and **⚙️ Developer** — and `./packetpulsetest.sh docs` fails if one is missing.

Run every case on a disposable stack (Part 0 §0.3) with **two organisations of your own**: `ACME` (yours) and `RIVAL` (someone else's). Unless a case says otherwise, "an administrator" is ACME's owner and "an engineer" holds ACME's built-in *NOC Engineer* role.

| Group | Chapters | Test IDs |
|---|---|---|
| 1 · Access | Sign-in and the second step · Sign-in security | `AUTH-*`, `SMS-*`, `LOC-*` |
| 2 · People | Staff · Roles and permissions · Sessions and seats · Check-ins · Activity | `STF-*`, `ACL-*`, `SEAT-*`, `CHK-*`, `AUD-*` |
| 3 · Diagnostics | Sites · Diagnostics and results · IPv6 · CSV · History and dashboard | `SITE-*`, `DIAG-*`, `V6-*`, `CSV-*`, `HIST-*` |
| 4 · The customer's side | Test from this device | `DEV-*` |
| 5 · Monitoring | SLA targets, schedules, alerts, maintenance, SLA report | `MON-*` |
| 6 · Integrations | Results API keys · Result export · Directory | `API-*`, `EXP-*`, `LDAP-*` |
| 7 · Platform and settings | Platform console · Settings · Public site | `PLAT-*`, `SET-*`, `WEB-*` |

---
