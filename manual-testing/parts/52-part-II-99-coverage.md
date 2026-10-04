## Feature → Coverage Matrix

What guards each area automatically, so a manual pass can spend its time where automation cannot reach: real devices, real browsers, real file servers, and judgement.

| Area | Manual IDs | Go unit and repository | Integration suites | Flutter |
|---|---|---|---|---|
| Sign-in, second step | `AUTH-*` | `pinglego/pkg/usermicroservice/userservice/` | `assignment`, `audit` | `pingleflutter/test/sign_in_test.dart` |
| SMS gateway, location rule | `SMS-*`, `LOC-*` | `pinglego/pkg/smsmicroservice/`, `pinglego/pkg/common/smsprovider/` | `acl` | `pingleflutter/test/sign_in_security_screen_test.dart` |
| Staff, roles | `STF-*`, `ACL-*` | `pinglego/pkg/staffmicroservice/` | `acl`, `tenancy` | `pingleflutter/test/staff_screen_test.dart` |
| Sessions, seats | `SEAT-*` | `pinglego/pkg/staffmicroservice/` | `assignment`, `tenancy` | `pingleflutter/test/session_screen_test.dart` |
| Check-ins | `CHK-*` | `pinglego/pkg/common/geocode/` | `tenancy`, `audit` | `pingleflutter/test/checkin_screen_test.dart` |
| Activity | `AUD-*` | `pinglego/pkg/common/auditlog/` | `audit` | `pingleflutter/test/audit_log_screen_test.dart` |
| Sites, policy | `SITE-*` | `pinglego/pkg/dnssitemicroservice/`, `pinglego/pkg/common/probeguard/` | `tenancy` | `pingleflutter/test/dns_site_screen_test.dart` |
| Diagnostics, IPv6 | `DIAG-*`, `V6-*` | `pinglego/pkg/pingmicroservice/`, `pinglego/pkg/diagnosticmicroservice/` | `contract`, `tenancy`, `load` | `pingleflutter/test/diagnostic_submit_screen_test.dart` |
| CSV | `CSV-*`, `API-003` | `pinglego/pkg/diagnosticmicroservice/diagnosticexport/` | `contract`, `tenancy`, `load` | `pingleflutter/test/diagnostic_history_screen_test.dart` |
| Device test | `DEV-*` | `pinglego/pkg/diagnosticmicroservice/diagnosticservice/` | `contract` | `pingleflutter/test/client_probe_run_test.dart` |
| Monitoring | `MON-*` | `pinglego/pkg/monitormicroservice/` | `monitor` | `pingleflutter/test/monitor_flows_test.dart` |
| Results API | `API-*` | `pinglego/pkg/common/apikeyauth/` | `contract`, `tenancy` | `pingleflutter/test/api_key_screen_test.dart` |
| Result export | `EXP-*` | `pinglego/pkg/common/filedrop/`, `pinglego/pkg/exportmicroservice/` | `contract`, `acl`, `tenancy`, `audit` | `pingleflutter/test/result_export_screen_test.dart` |
| Directory | `LDAP-*` | `pinglego/pkg/common/ldapclient/`, `pinglego/pkg/ldapmicroservice/` | — | `pingleflutter/test/ldap_settings_screen_test.dart` |
| Platform | `PLAT-*` | `pinglego/pkg/platformmicroservice/` | `assignment` | `pingleflutter/test/platform_console_screen_test.dart` |
| Settings, catalogue | `SET-*` | `pinglego/pkg/initmicroservice/` | `translation` | `pingleflutter/test/settings_screen_test.dart` |

> [!NOTE]
> What no suite covers, and why manual cases exist for it: a **real SMS** arriving on a phone (`AUTH-009`), a **real authenticator app** (`AUTH-001`), **browser location prompts** (`CHK-001`, `LOC-002`), **real SFTP and FTPS servers** (`EXP-006`, `EXP-011`), **spreadsheet applications** opening the CSV (`CSV-002`), and **IPv6 on the live host** (`V6-001`).

---
