## Feature → Coverage Matrix

What guards each area automatically, so a manual pass can spend its time where automation cannot reach: real devices, real browsers, real file servers, and judgement.

| Area | Manual IDs | Go unit and repository | Integration suites | Flutter |
|---|---|---|---|---|
| Sign-in, second step | `AUTH-*` | `packetpulsego/pkg/usermicroservice/userservice/` | `assignment`, `audit` | `packetpulseflutter/test/sign_in_test.dart` |
| SMS gateway, location rule | `SMS-*`, `LOC-*` | `packetpulsego/pkg/smsmicroservice/`, `packetpulsego/pkg/common/smsprovider/` | `acl` | `packetpulseflutter/test/sign_in_security_screen_test.dart` |
| Staff, roles | `STF-*`, `ACL-*` | `packetpulsego/pkg/staffmicroservice/` | `acl`, `tenancy` | `packetpulseflutter/test/staff_screen_test.dart` |
| Sessions and licence | `WHO-*` | `packetpulsego/pkg/staffmicroservice/` | `assignment`, `tenancy` | `packetpulseflutter/test/session_screen_test.dart` |
| Check-ins | `CHK-*` | `packetpulsego/pkg/common/geocode/` | `tenancy`, `audit` | `packetpulseflutter/test/checkin_screen_test.dart` |
| Activity | `AUD-*` | `packetpulsego/pkg/common/auditlog/` | `audit` | `packetpulseflutter/test/audit_log_screen_test.dart` |
| Sites, policy | `SITE-*` | `packetpulsego/pkg/dnssitemicroservice/`, `packetpulsego/pkg/common/probeguard/` | `tenancy` | `packetpulseflutter/test/dns_site_screen_test.dart` |
| Diagnostics, IPv6 | `DIAG-*`, `V6-*` | `packetpulsego/pkg/pingmicroservice/`, `packetpulsego/pkg/diagnosticmicroservice/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_submit_screen_test.dart` |
| CSV | `CSV-*`, `API-003` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticexport/` | `contract`, `tenancy`, `load` | `packetpulseflutter/test/diagnostic_history_screen_test.dart` |
| Run diagnostic, device test | `DEV-*` | `packetpulsego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract` | `packetpulseflutter/test/client_probe_run_test.dart`, `packetpulseflutter/test/run_diagnostic_screen_test.dart` |
| Monitoring | `MON-*` | `packetpulsego/pkg/monitormicroservice/` | `monitor` | `packetpulseflutter/test/monitor_flows_test.dart` |
| Results API | `API-*` | `packetpulsego/pkg/common/apikeyauth/` | `contract`, `tenancy` | `packetpulseflutter/test/api_key_screen_test.dart` |
| Result export | `EXP-*` | `packetpulsego/pkg/common/filedrop/`, `packetpulsego/pkg/exportmicroservice/` | `contract`, `acl`, `tenancy`, `audit` | `packetpulseflutter/test/result_export_screen_test.dart` |
| Directory | `LDAP-*` | `packetpulsego/pkg/common/ldapclient/`, `packetpulsego/pkg/ldapmicroservice/` | — | `packetpulseflutter/test/ldap_settings_screen_test.dart` |
| Platform | `PLAT-*` | `packetpulsego/pkg/platformmicroservice/` | `assignment` | `packetpulseflutter/test/platform_console_screen_test.dart` |
| Settings, catalogue | `SET-*` | `packetpulsego/pkg/initmicroservice/` | `translation` | `packetpulseflutter/test/settings_screen_test.dart` |

> [!NOTE]
> What no suite covers, and why manual cases exist for it: a **real SMS** arriving on a phone (`AUTH-009`), a **real emailed code** arriving in an inbox (`AUTH-001`), **browser location prompts** (`CHK-001`, `LOC-002`), **real SFTP and FTPS servers** (`EXP-006`, `EXP-011`), **spreadsheet applications** opening the CSV (`CSV-002`), and **IPv6 on the live host** (`V6-001`).

---
