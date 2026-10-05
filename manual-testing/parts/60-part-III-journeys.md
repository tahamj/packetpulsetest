# Part III — End-to-End Journeys

Each journey crosses several chapters the way a real customer does. Run them on a disposable stack after the chapter cases; every step's result is the precondition of the next, so stop at the first failure and file it against the step's case ID.

---

### JRN-001 · A new customer, from contract to first evidence

**Personas:** platform superuser, ACME owner, NOC engineer · **Covers:** `PLAT-001`, `AUTH-008`, `AUTH-001`, `STF-001`, `SITE-002`, `DIAG-001`, `DIAG-004`, `CSV-001`

1. On the console, the superuser creates **ACME**, issues a licence for 10 people and downloads its licence file, naming the owner's address.
2. The file goes into `LICENCE_DIR` on ACME's server. The owner signs up there with that address and enters the emailed code.
3. The owner adds an engineer (*NOC Engineer*) and pastes twenty sites with **Bulk import**.
4. The engineer signs in with their emailed code and runs a diagnostic **from the server** with TT `TT-JRN-001`.
5. The engineer exports the **PDF** and the **CSV**.

**Expected:** the PDF and CSV carry `TT-JRN-001`; Activity shows the owner's sign-in, the staff and site changes, both exports. 🛑 The engineer's sign-in is **not** in Activity.

---

### JRN-002 · A NOC engineer works a ticket

**Covers:** `DIAG-001`–`DIAG-005`, `HIST-001`, `V6-001`

1. Ticket `TT-JRN-002` arrives: Customer `CUST-88412`, slow calls.
2. Run against the customer's sites. Read loss, jitter and MOS on each row.
3. Re-run after the carrier's fix.
4. In **History**, compare both attempts.

**Expected:** both attempts kept; the second's average loss and jitter lower; the dual-stack site shows its IPv6 row *not counted*.

---

### JRN-003 · A field engineer's day

**Covers:** `STF-004`, `LOC-002`, `CHK-001`–`CHK-002`, `DEV-001`–`DEV-006`

1. The administrator sets the engineer to **Record location** and turns on **Require location to sign in**.
2. On a phone, the engineer signs in, allowing location.
3. At the customer's site: **Run diagnostic → From this device**, then attach the run to `TT-JRN-003`.
4. Signs out, allowing location.

**Expected:** **Check-ins** shows the sign-in and sign-out with places and a map link; the ticket shows the device run, *measured on a device*; the administrator's Activity shows no entry for the engineer's sign-in.

---

### JRN-004 · The IT system takes the results

**Covers:** `API-001`–`API-003`, `EXP-002`, `EXP-006`–`EXP-008`

1. The administrator issues a Results API key; the IT system pulls yesterday's CSV.
2. The administrator sets up **Result export** to the IT team's SFTP server: test, trust the key, save, every 15 minutes.
3. Engineers run diagnostics for an hour.
4. The SFTP password is changed on the server for one interval, then restored.

**Expected:** a file per interval with results; the interval with the wrong password shows *Not delivered*, and the next file holds that period's results too; every `result_id` appears exactly once across all the files — a failed delivery sent nothing, so nothing is sent twice.

---

### JRN-005 · Monitoring catches an outage and its recovery

**Covers:** `MON-001`–`MON-005`, `MON-007`

1. A default target; a schedule every 5 minutes over every enabled site; a webhook channel; damping 2.
2. Black-hole one site for 15 minutes, then restore it.
3. Declare a maintenance window over another site and break it inside the window.

**Expected:** one breach alert, one recovery; the maintenance site raises nothing and its availability is untouched.

---

### JRN-006 · An attack on an administrator's account

**Covers:** `AUD-003`–`AUD-006`, `AUTH-006`, `AUTH-007`

1. From an unfamiliar browser, try the owner's address with ten wrong passwords.
2. Then, with the right password (a leaked one), ten wrong codes.

**Expected:** Activity shows the wrong passwords and wrong codes as *Sign-in failed* with the address and browser, ending in *Locked after too many wrong codes*; the owner, from their own browser with an emailed code, is locked for 15 minutes and then signs in. 🛑 The wrong passwords alone did not lock the account.

---

### JRN-007 · A lost phone

**Covers:** `AUTH-012`, `WHO-003`, `STF-004`

1. An engineer set to text loses their phone, still signed in to the app on it.
2. Their administrator signs that device out (**Who is signed in → Sign device out**) and switches them to **Email**.
3. The engineer signs in on a laptop with the code emailed to them.

**Expected:** the lost phone's next request returns it to sign-in; the sign-out and the change of method are in Activity, naming who made them; the engineer's code arrives by email, not by text.

---

### JRN-008 · A licence lapses

**Covers:** `DIAG-006`, `PLAT-002`, `API-002`

1. The superuser suspends ACME's licence.
2. Engineers try to run a diagnostic; read and export old ones; the IT system pulls through the API.
3. The superuser resumes it.

**Expected:** runs refused with who can renew; reading, exports and the Results API keep working throughout.

---
