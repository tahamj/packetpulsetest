## Group 1 — Access

Signing in, the second step, and the organisation's own rules for both.

---

### 1.1 🔑 Sign-in and the second step

**Screen:** the sign-in page at `/app/` · **Routes:** `POST /user/signin`, `/user/signin/verify`, `/user/signin/resend`, `/user/signin/enrol`, `/user/signout`

- 🌟 **Commercial Presentation & Sales Pitch**: Two-step sign-in for *everyone*, not a premium add-on. A stolen password alone opens nothing. Field staff can get codes by text through **the operator's own Twilio account** — their sender, their DLT registration, their bill — while owners and the platform operator always use an authenticator app, so the accounts that matter most never depend on a gateway someone else configures. Recovery codes mean a lost phone is an inconvenience, not a helpdesk ticket.
- 📖 **User Guide & Operational Flow**:
  - **First sign-in:** email and password → scan the QR code with any authenticator app (or type the key shown beneath it) → enter one code → **save the ten recovery codes** shown once → you are in.
  - **Every sign-in after:** email and password → the 6-digit code from the app (or texted to •••• 1234 if your administrator set you up for text; *Resend* after 30 seconds).
  - **Lost phone:** *Lost your phone? Use a recovery code* on the code screen. Each works once. No codes left? Your administrator uses **Staff → Reset authenticator**; an owner asks the platform operator.
  - **Location:** if you are asked to share your location, the browser or phone asks your permission. See §1.2 for when sharing is required.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `AUTH-001` | New engineer | First sign-in sets up the authenticator | Administrator adds the engineer (**Staff → Add**). Engineer signs in with email and password | QR code and key shown; one correct code opens the app **after** the ten recovery codes are shown. 🛑 **Must NOT** reach the app before the codes are shown, or show them again later |
  | `AUTH-002` | Engineer | Normal sign-in | Sign out; sign in; enter the current code | Signed in. The **Sessions and seats** row shows this device |
  | `AUTH-003` | Engineer | A wrong code says how many tries are left | Enter `000000` | Refused (`invalid_code`) with attempts left; after **5** wrong codes on one challenge it is spent (`challenge_expired`, *Start again*). 🛑 **Must NOT** accept a 6th guess on the same challenge |
  | `AUTH-004` | Engineer | A code cannot be used twice | Sign in with a code; sign out; sign in again within the same 30 seconds with the **same** code | Second use refused; the next code (30 s later) works. 🛑 **Must NOT** accept a replayed code |
  | `AUTH-005` | Engineer | A recovery code works once | On the code screen choose *Lost your phone? Use a recovery code*, enter one; sign out; try the same one again | First signs in; second refused. 🛑 **Must NOT** accept a spent recovery code |
  | `AUTH-006` | Tester | Ten wrong codes lock the account, a wrong password never does | Give 10 wrong codes across fresh challenges; then the right password and right code. Separately, 15 wrong **passwords** on another account | First account: *Locked — try again later* (`locked_out`, 15 min) even with the right code. Second account: still signs in with the right password. 🛑 **Must NOT** lock an account by wrong passwords — a stranger could lock anyone out |
  | `AUTH-007` | Tester | Sign-in does not reveal which addresses exist | Sign in with an unknown email; then a known email with a wrong password | The same message and a similar response time. 🛑 **Must NOT** say "no such user" or answer the unknown address noticeably faster |
  | `AUTH-008` | Owner | Owners always use an authenticator | Set the owner to *Text message* in **Staff → How they sign in** with a gateway on; sign in as the owner | Authenticator code asked, not a text. 🛑 **Must NOT** text an owner's code |
  | `AUTH-009` | Engineer set to text | A texted code, with resend limits | Gateway on (§1.2), engineer has a mobile number and *Text message*. Sign in; wait; *Resend* | Code arrives; hint shows the last four digits; *Resend* unavailable for 30 s; at most 3 sends per sign-in and 5 per hour (`rate_limited`) |
  | `AUTH-010` | Engineer set to text | No gateway means the authenticator, not a lock-out | Switch the gateway **off**; sign in as the engineer | Authenticator set-up (or code) asked instead. 🛑 **Must NOT** refuse the sign-in for want of a gateway |
  | `AUTH-011` | Engineer set to text | A failing gateway is said, not bypassed | Gateway on with a wrong auth token; sign in as the engineer | 503 *Your sign-in code could not be sent just now. Try again.* 🛑 **Must NOT** fall back to an authenticator the engineer never set up |
  | `AUTH-012` | Administrator | Reset authenticator for a lost phone | **Staff → (engineer) → Reset authenticator** | The engineer's sessions end at once; their next sign-in sets up a new authenticator. Activity shows the reset, by whom |
  | `AUTH-013` | Superuser | Only the platform operator resets an owner | **Platform → (organisation) → Owners → Reset**; then try `POST /user/{ownerId}/secondfactor/reset` as an organisation administrator | Superuser: reset done, audited. Administrator: 403. 🛑 **Must NOT** let anyone inside the organisation reset its owner |
  | `AUTH-014` | Engineer | An ended session returns to sign-in once | Revoke the engineer's session from another device, then use the app | One clean return to the sign-in page. 🛑 **Must NOT** loop between sign-in and an error |
  | `AUTH-015` | Tester | A forged client address is ignored | `curl -H 'True-Client-IP: 6.6.6.6' -H 'X-Forwarded-For: 6.6.6.6'` a sign-in from a machine that is not a trusted proxy | Sessions and Activity show the real peer address. 🛑 **Must NOT** record `6.6.6.6` |
- ⚙️ **Developer Guide & Release Confidence**:
  - Flow: `pinglego/pkg/usermicroservice/userservice/UserSignInSteps.go` — the password step returns a challenge (`totp`, `totp_enrol` or `sms`), never a session; `newSession` runs only after verify. Challenge ids are 32 random bytes, stored as SHA-256; texted codes are keyed HMACs.
  - 🔒 Attempts are counted **before** the code is checked, in one conditional `UPDATE … RETURNING`, so parallel guesses cannot slip under the limit.
  - Limits: `maxAttempts = 5`, `lockAfter = 10`, `lockFor = 15m`, `resendAfter = 30s`, 3 sends per challenge, 5 texts per hour; a separate rate limiter for the second step (`SECOND_STEP_ATTEMPTS_PER_MINUTE`).
  - Client IPs: `pinglego/pkg/common/apiratelimit/ApiRateLimit.go` walks `X-Forwarded-For` from the right past `TRUSTED_PROXIES` only, and returns canonical addresses (`::ffff:a.b.c.d` → `a.b.c.d`).
  - Coverage: `pinglego/pkg/usermicroservice/userservice/UserSignInSteps_test.go`, the `assignment` and `audit` suites (every integration test signs in with a real TOTP code).

---

### 1.2 🛡️ Sign-in security — the SMS gateway and the location rule

**Screen:** Configure → **Sign-in security** · **Routes:** `GET/PUT /sms/gateway`, `POST /sms/gateway/test`, `GET/PUT /organisation/settings` · **Capability:** `staff_manage`

- 🌟 **Commercial Presentation & Sales Pitch**: The operator brings its own SMS provider account, so codes go out under its registered sender at its negotiated rates — and in India, under its own DLT template. One switch makes **sharing a location a condition of signing in**, for operators who must prove where field staff were. Owners are always let in, so a browser that will not share a position can never lock the organisation out.
- 📖 **User Guide & Operational Flow**:
  - **SMS gateway:** Twilio account SID (`AC…`), auth token (stored encrypted, never shown again — leave empty to keep it), a sender number or a messaging service SID, and the message with `%s` where the code goes. **Send a test to my mobile** texts *your own* number from your staff record.
  - **Location at sign-in:** *Require location to sign in* refuses a sign-in without a position for people whose location is recorded. Who is recorded is set per person on **Staff → Record location at sign-in and sign-out**.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `SMS-001` | Administrator | The form names each mistake | Save with SID `AB123`, sender `98765`, message without `%s`, and a 301-character message | Each field shows its own error (`validation_failed`). 🛑 **Must NOT** save any part of it |
  | `SMS-002` | Administrator | The token is kept, never shown | Save a full gateway; reload; save again with the token empty | *A token is stored. Leave this empty to keep it.* hint; the second save keeps it. `GET /sms/gateway` carries `has_auth_token: true` and no token. 🛑 **Must NOT** return the token or clear it on an empty save |
  | `SMS-003` | Administrator | Switching on needs everything a send needs | Turn the gateway on with no sender and no token | Refused against those fields. Saved switched **off**, the half-filled form is kept |
  | `SMS-004` | Administrator | The test goes only to my own mobile | Clear your own mobile number on the Staff screen; *Send a test to my mobile*; then add it and send again | First: *Add your own mobile number on the Staff screen*. Second: sent, naming your number. 🛑 **Must NOT** accept a number from the request — the endpoint takes none |
  | `SMS-005` | Administrator | A refused account is named | Save a wrong token, switch on, send a test | *The SMS provider refused these credentials* against the token. 🛑 **Must NOT** show the provider's raw reply (it can carry the SID) |
  | `LOC-001` | Administrator | The location rule saves alone | Turn *Require location to sign in* on | Saved; the device-test target (§4.1) is unchanged. 🛑 **Must NOT** reset any other organisation setting |
  | `LOC-002` | Engineer | Required means refused without a position | Rule on; engineer with *Record location* on; deny the browser's location prompt at sign-in | Refused (`location_required`, action *share location*) with guidance to allow it. Allowing it signs in. The owner, denying it, still signs in |
  | `LOC-003` | Engineer | Not required means recorded as refused | Rule off; deny the prompt | Signed in; **Check-ins** shows *Location refused* for that sign-in |
  | `LOC-004` | Engineer | A garbled position is not trusted | Send `POST /user/signin/verify` with `latitude: 123` | Refused as an invalid location (422) at sign-in; at sign-out recorded as *unavailable* rather than refusing the sign-out |
- ⚙️ **Developer Guide & Release Confidence**:
  - Gateway: `pinglego/pkg/smsmicroservice/smsservice/SmsService.go` (token sealed with the server's secret box, keyed from `JWT_SECRET`); provider: `pinglego/pkg/common/smsprovider/` (Twilio, fixed host).
  - Settings: `pinglego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`; `PUT /organisation/settings` requires **both** `require_checkin_location` and `device_test_target` — a whole-row save.
  - Policy: superusers and owners are always asked for a position and never required to give one; no staff row means not asked; otherwise the person's `capture_location` and the organisation's rule decide. An unreadable policy fails closed.
  - Coverage: `pinglego/pkg/smsmicroservice/**`, `pinglego/pkg/usermicroservice/userservice/UserCheckin_test.go`, `pingleflutter/test/sign_in_security_screen_test.dart`.

---
