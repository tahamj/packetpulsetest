## Group 1 — Access

Signing in, the second step, and the organisation's own rules for both.

---

### 1.1 🔑 Sign-in and the second step

**Screen:** the sign-in page at `/app/` · **Routes:** `POST /user/signin`, `/user/signin/verify`, `/user/signin/resend`, `/user/signout`

- 🌟 **Commercial Presentation & Sales Pitch**: Two-step sign-in for *everyone*, not a premium add-on. A stolen password alone opens nothing: after it comes a one-time code, emailed to the person's sign-in address or — for field staff the administrator chooses — texted through **the operator's own Twilio account** (their sender, their DLT registration, their bill). Owners always get theirs by email, so the account that matters most never depends on a gateway someone else configures. There is no app to install and nothing to recover: a lost phone is one action for an administrator.
- 📖 **User Guide & Operational Flow**:
  - **Every sign-in:** email and password → *We emailed a 6-digit code to a•••@acme.example* (or *We texted a 6-digit code to •••• 1234* when your administrator set you up for text) → enter it → you are in. A code is good for 10 minutes; *Resend* after 30 seconds.
  - **The owner's first sign-in:** *First time here? Set up the owner's account* — only the address named in this server's licence can do this (§7.1). The emailed code finishes it, and the owner arrives as the Administrator.
  - **Lost phone:** your administrator signs that device out under **Who is signed in** (§2.3) and changes your number or switches you to email under **Staff → How they sign in**.
  - **Location:** if you are asked to share your location, the browser or phone asks your permission. See §1.2 for when sharing is required.
- 🧪 **Manual Testing Playbook**:
  | ID | Persona | Scenario | Steps | Observable Expected Result |
  |---|---|---|---|---|
  | `AUTH-001` | New engineer | The first sign-in is the same as every other | Administrator adds the engineer (**Staff → Add**). Engineer signs in with email and password | The code step names their inbox, partly hidden; the emailed code opens the app. 🛑 **Must NOT** reach the app on the password alone, or show the whole address in the hint |
  | `AUTH-002` | Engineer | Normal sign-in | Sign out; sign in; enter the emailed code | Signed in. **Who is signed in** shows this device |
  | `AUTH-003` | Engineer | A wrong code says how many tries are left | Enter `000000` | Refused (`invalid_code`) with attempts left; after **5** wrong codes on one sign-in it is spent (`challenge_expired`, *Start again*). 🛑 **Must NOT** accept a 6th guess on the same sign-in |
  | `AUTH-004` | Engineer | A code belongs to its own sign-in | Sign in with a code; sign out; start another sign-in and enter the first code | Refused; the code emailed for this sign-in works. 🛑 **Must NOT** accept a code from an earlier sign-in |
  | `AUTH-005` | Tester | Codes cannot be pumped | *Resend* repeatedly; then sign in over and over within the hour | *Resend* unavailable for 30 s and at most 3 sends per sign-in; the 11th code to one person in an hour is refused (`rate_limited`). 🛑 **Must NOT** send email or texts without limit |
  | `AUTH-006` | Tester | Ten wrong codes lock the account, a wrong password never does | Give 10 wrong codes across fresh sign-ins; then the right password and right code. Separately, 15 wrong **passwords** on another account | First account: *Locked — try again later* (`locked_out`, 15 min) even with the right code. Second account: still signs in with the right password. 🛑 **Must NOT** lock an account by wrong passwords — a stranger could lock anyone out |
  | `AUTH-007` | Tester | Sign-in does not reveal which addresses exist | Sign in with an unknown email; then a known email with a wrong password | The same message and a similar response time, and no code sent. 🛑 **Must NOT** say "no such user" or answer the unknown address noticeably faster |
  | `AUTH-008` | Owner | An owner's code is always emailed | Set the owner to *Text message* in **Staff → How they sign in** with a gateway on; sign in as the owner | Code emailed, not texted. 🛑 **Must NOT** text an owner's code |
  | `AUTH-009` | Engineer set to text | A texted code, with resend limits | Gateway on (§1.2), engineer has a mobile number and *Text message*. Sign in; wait; *Resend* | Code arrives; the hint shows the last four digits; *Resend* unavailable for 30 s; at most 3 sends per sign-in and 10 codes an hour (`rate_limited`) |
  | `AUTH-010` | Engineer set to text | No gateway means email, not a lock-out | Switch the gateway **off**; sign in as the engineer | The code is emailed instead. 🛑 **Must NOT** refuse the sign-in for want of a gateway |
  | `AUTH-011` | Engineer set to text | A failing gateway is said, not hidden | Gateway on with a wrong auth token; sign in as the engineer | 503 *Your sign-in code could not be sent just now. Try again.* 🛑 **Must NOT** claim a code was sent |
  | `AUTH-012` | Administrator | A lost phone | **Who is signed in → Sign this device out** on the engineer's phone; **Staff → How they sign in → Email** | The phone's session ends at once; the engineer's next sign-in emails the code. Activity records who signed the device out |
  | `AUTH-013` | Tester | The superuser exists only on the console | On a customer's server (no `LICENCE_SIGNING_KEY`), sign in as `superuser@rummaan53.com` — or as any account flagged superuser in its database | Refused (*This account has been deactivated*). On the console the code is emailed to `superuser@rummaan53.com`. 🛑 **Must NOT** admit a superuser on a customer's server |
  | `AUTH-014` | Engineer | An ended session returns to sign-in once | Revoke the engineer's session from another device, then use the app | One clean return to the sign-in page, and one `POST /user/signout` in the browser's network panel. 🛑 **Must NOT** loop between sign-in and an error, or send sign-outs over and over |
  | `AUTH-015` | Tester | A forged client address is ignored | `curl -H 'True-Client-IP: 6.6.6.6' -H 'X-Forwarded-For: 6.6.6.6'` a sign-in from a machine that is not a trusted proxy | Sessions and Activity show the real peer address. 🛑 **Must NOT** record `6.6.6.6` |
  | `DEMO-001` | Visitor | The public demo signs in through the real two-step flow | On the console with `SEED_DEMO` on, sign in as `demo-engineer@<domain>` with the shared demo password | The real SMS step appears with the code **shown on screen** and filled in; one tap opens the app. 🛑 **Must NOT** text or email anything |
  | `DEMO-002` | Tester | The demo code is not a master key | Use the demo code (`DEMO_SMS_CODE`, default `111111`) as the second-step code for a **non-demo** account | Refused as a wrong code. 🛑 **Must NOT** admit any account outside the demo organisation |
  | `DEMO-003` | Tester | The demo is capped | Open `DEMO_SEATS` (default and minimum 5) concurrent demo sessions, then sign in once more | The extra sign-in is refused with *the demo is busy* (`demo_busy`). 🛑 **Must NOT** let shared demo accounts open unlimited sessions |
  | `DEMO-005` | Tester | A walkthrough left idle gives up its place | Fill the demo as in `DEMO-003`; close every tab without signing out; wait 30 minutes; sign in again | Admitted: only sessions used in the last 30 minutes count. Reopening a closed tab within its 12 hours still works. 🛑 **Must NOT** keep visitors out because earlier ones walked away |
  | `DEMO-004` | Tester | The shared demo cannot be hijacked or turned on the network | As a demo account try **Change password**; then run a diagnostic against a private/internal address added to the demo inventory | Password change refused (`demo_read_only`); the private target is refused by the probe policy. 🛑 **Must NOT** let a visitor change a shared credential or probe `10.x`/internal space |
- ⚙️ **Developer Guide & Release Confidence**:
  - Flow: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps.go` — the password step returns a challenge (`email` or `sms`), never a session; `newSession` runs only after verify, and checks the licence again (§7.1). Challenge ids are 32 random bytes, stored as SHA-256; codes are stored as keyed hashes bound to their challenge.
  - 🔒 Attempts are counted **before** the code is checked, in one conditional `UPDATE … RETURNING`, so parallel guesses cannot slip under the limit.
  - Limits: `maxAttempts = 5`, `lockAfter = 10`, `lockFor = 15m`, `resendAfter = 30s`, `maxSendsPerChallenge = 3`, `maxCodesPerHour = 10` by either channel; a separate rate limiter for the second step (`SECOND_STEP_ATTEMPTS_PER_MINUTE`).
  - Email: `packetpulsego/pkg/common/mailer/Mailer.go` — STARTTLS with a verified certificate (credentials are never sent otherwise), header injection refused, the load spread over up to four sending accounts (`SMTP_FROM_0…3`, `SMTP_PASSWORD_0…3`). A development or test server writes codes to `OTP_OUTBOX_FILE` instead (`packetpulsego/pkg/common/mailer/Outbox.go`); production refuses it.
  - Client IPs: `packetpulsego/pkg/common/apiratelimit/ApiRateLimit.go` walks `X-Forwarded-For` from the right past `TRUSTED_PROXIES` only, and returns canonical addresses (`::ffff:a.b.c.d` → `a.b.c.d`).
  - Coverage: `packetpulsego/pkg/usermicroservice/userservice/UserSignInSteps_test.go`, `packetpulsego/pkg/common/mailer/Mailer_test.go`, the `assignment` and `audit` suites (every integration test finishes its sign-in with the code from the outbox).

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
  - Gateway: `packetpulsego/pkg/smsmicroservice/smsservice/SmsService.go` (token sealed with the server's secret box, keyed from `JWT_SECRET`); provider: `packetpulsego/pkg/common/smsprovider/` (Twilio, fixed host).
  - Settings: `packetpulsego/pkg/staffmicroservice/staffservice/StaffCheckinService.go`; `PUT /organisation/settings` requires **both** `require_checkin_location` and `device_test_target` — a whole-row save.
  - Policy: superusers and owners are always asked for a position and never required to give one; no staff row means not asked; otherwise the person's `capture_location` and the organisation's rule decide. An unreadable policy fails closed.
  - Coverage: `packetpulsego/pkg/smsmicroservice/**`, `packetpulsego/pkg/usermicroservice/userservice/UserCheckin_test.go`, `packetpulseflutter/test/sign_in_security_screen_test.dart`.

---
