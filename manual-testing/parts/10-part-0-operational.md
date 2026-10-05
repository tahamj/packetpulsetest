# Part 0 — Operational Prerequisites

Everything you need before the first test case: where to run PacketPulse, how to sign in to it, and how the automated gate relates to the manual one.

## 0.1 Environments

| Environment | Address | Database | Use it for | Never use it for |
|---|---|---|---|---|
| **Local development** | `http://localhost:8080` (API) · `flutter run -d chrome` (app) | `packetpulsedb` on the local PostgreSQL | Building and trying features | Destructive cases you cannot undo by hand |
| **Disposable test** | any free port, e.g. `http://localhost:18080` | a database you create for the run (e.g. `packetpulse_manual_test`) | **Every manual test pass** and the integration suites | — |
| **Live** | **https://packetpulse.rummaan53.com** (site `/`, app `/app/`, API `/api/`) | `packetpulsedb` on the Oracle host, behind PgBouncer | Read-only walkthroughs in the demo organisation; the post-deploy checks of Part V §5.4 | Tenancy, ACL, lockout or delete cases; anything that changes a real customer's data |

> [!WARNING]
> ⚠️ **TRAP — the `.env` fallback.** The Go database tests read `DATABASE_URL`, and when it is unset they fall back to the repository's `.env` — your development database — and **apply every migration to it on first connect**. Always export a disposable `DATABASE_URL` before `go test`, or run tests only through `./packetpulsetest.sh`, which does.

## 0.2 Toolchain

| Tool | Version | Why |
|---|---|---|
| Go | 1.26 | `packetpulsego/` and the integration suites in `packetpulsetest/golang/` |
| Flutter | 3.35 | `packetpulseflutter/` — web, macOS, Windows, Android, iOS |
| PostgreSQL | 18 (client tools too) | The database, and `pg_dump` for the backup suite |
| Python | 3.9+ (3.13 recommended) | `scripts/intellicodegen/packetpulsestrings.py`, the coverage floors, this guide's generator |
| gpg, rsync | any recent | The backup suite (`scripts/deploy/backup/`) |
| govulncheck | latest | Part of the `unit` suite |
| Node | 18+ | The public self-test page's probe engine tests |

## 0.3 Starting a disposable stack

```bash
# 1. a database of your own
createdb packetpulse_manual_test

# 2. the API, migrating that database at boot (migrations are embedded in the binary)
cd packetpulsego
DATABASE_URL=postgres://packetpulse:…@localhost:5432/packetpulse_manual_test?sslmode=disable \
PORT=18080 APP_ENV=test REVERSE_GEOCODING=off go run ./cmd/packetpulseserver

# 3. the app, pointed at it
cd ../packetpulseflutter
flutter run -d chrome --dart-define=PACKETPULSE_API_BASE_URL=http://localhost:18080
```

`REVERSE_GEOCODING=off` stops sign-in positions being sent to a map service from a test run. Every other setting comes from `.env` (copy `.env.example`); Appendix D lists them all.

With the development key pair in `.env` (`LICENCE_PUBLIC_KEY` and `LICENCE_SIGNING_KEY`), the stack is a **console**: the superuser can sign in and issue licence files (**Platform → Download licence file**). Copy a file into `LICENCE_DIR` and its organisation can sign in; the server verifies it at every sign-in. Only the demo's own licence is written there for you. A customer's server has no signing key, only the licence files copied into its `LICENCE_DIR`.

## 0.4 Accounts and sign-in codes

| Account | How it exists | Second step |
|---|---|---|
| **Platform superuser** | `superuser@rummaan53.com` (fixed in code) with `OWNER_PASSWORD`, created at boot **on the console only** | A code emailed to that address. |
| **Organisation owner** | The first sign-up with the address named in the organisation's licence file becomes its Administrator | A code emailed to the owner, always. |
| **Anyone else** | **Staff → Add** by an administrator, who shares the sign-in with them | A code emailed to their sign-in address, or texted when they are set to text and the organisation's SMS gateway is on. |
| **Demo logins** (`SEED_DEMO=true`, console only) | `DEMO_EMAIL` / `DEMO_PASSWORD` (owner), plus `demo-engineer@…` and `demo-viewer@…` | The texted-code step, with the code (`DEMO_SMS_CODE`) shown on screen instead of sent. At most `DEMO_SEATS` walkthroughs at once. |

To act as a person in a manual test, read their code from the **code outbox**: when `OTP_OUTBOX_FILE` is set, every code is appended to that file as one JSON line (`challenge_id`, `channel`, `to`, `code`, `at`) instead of being emailed or texted. Without it, codes go to real mailboxes through the Gmail accounts in `SMTP_FROM_n` / `SMTP_PASSWORD_n`.

> [!IMPORTANT]
> 🔒 **INVARIANT** — production refuses `OTP_OUTBOX_FILE`, because a file of live sign-in codes is a second factor held by everyone who can read the file. It also refuses to boot with no SMTP account, since email is everyone's fallback, and refuses a `LICENCE_PUBLIC_KEY` that would replace the vendor's key. `packetpulsego/pkg/common/config/Config.go` returns an error at boot; `REL-004` checks the live host.

## 0.5 The automated gate

`./packetpulsetest.sh` runs every suite; `./packetpulsetest.sh unit client` is the fast pair. Manual testing **adds to** this gate — it does not replace it. A release needs both (Part V).

| Suite | What it proves | Needs |
|---|---|---|
| `tenancy` | One organisation can never read or change another's data | server |
| `assignment` | Licence files: owner sign-up, sign-in only under a genuine licence, people counted, switched off and deleted | server |
| `acl` | Every role gets exactly its capabilities, both ways | server |
| `audit` | The registry covers every mutation; the chain verifies; sign-ins (and failed ones) are recorded | server |
| `translation` | The catalogue is complete and served | server |
| `contract` | The Results API and CSV shapes, the export's destination guard | server |
| `monitor` | SLA grading, schedules, alert damping, maintenance windows | server |
| `unit` | Every Go package with `-race -shuffle`, coverage floors, govulncheck, the string generator | database |
| `client` | Every Flutter screen, `flutter analyze`, the client coverage floor | — |
| `backup` | The backup, drill and restore scripts, for real | database |
| `docs` | This guide is current and every number, path, route and capability in it matches source | — |
| `load` *(opt-in)* | 20 people, a 200-site sweep, API and CSV pulls against p95 budgets | server |

Integration suites read `PACKETPULSE_TEST_URL` (default `http://localhost:8080`). `PACKETPULSE_TEST_REQUIRE_SUPERUSER=1` turns a missing superuser into a failure rather than a skip — a skipped suite looks exactly like a passing one.

## 0.6 Deploying

`scripts/deploy/PacketPulseDeploy.sh --host mshop.rummaan53.com` builds from the working tree, runs `unit client backup` and the integration suites against `PACKETPULSE_TEST_URL`, cross-compiles for linux/arm64, builds the web app with base href `/app/`, backs up the live database, then switches the release and restarts. `--api-only`, `--web-only` and `--site-only` ship one part; `--skip-guards` and `--skip-backup` are for a logged emergency only. Part V §5.4 is what to check afterwards.

---
