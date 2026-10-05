# Part L — Learn PacketPulse

The ideas every other part assumes, in plain words. Read this once, whatever your role.

## L.1 Organisations, people and the licence

An **organisation** is one customer — a telecom operator such as Northwind Telecom. Everything it owns (sites, diagnostics, keys, settings, its trail) is invisible to every other organisation. 🔒 **INVARIANT** — every tenant query carries the caller's `organisation_id`; a request with none is refused, never answered for everyone.

Each organisation runs on a **licence file** signed by PacketPulse and copied onto its server. The server checks it at **every sign-in**: without a genuine licence in force, nobody in that organisation signs in. The licence names the owner's address and how many **people** it covers. Everyone on the books counts, switched on or switched off, because a person who leaves is switched off rather than erased. Someone who never ran a test or signed in can be **deleted**, and then no longer counts. Signing up does not create an organisation: the first person to sign up with the owner's address, proved by an emailed code, becomes its Administrator and adds everyone else.

People hold a **role** (Administrator, NOC Engineer, Viewer, or one of the organisation's own) made of **capabilities** such as `diagnostic_run` or `staff_manage` (Appendix B). A person can be given an individual *allow* or *deny* on top. Changing someone's authority signs them out at once.

## L.2 Sites, tickets and sweeps

A **site** is an address worth testing: a customer router, a point of presence, a resolver. It may be an IP address or a hostname (resolved at test time, so a name that stops resolving is itself a fault).

A **diagnostic** tests sites and files the result against a **TT number** — a trouble ticket in the operator's own system — and a **Customer ID**. A ticket can be tested many times; the attempts together show how the fault was worked. The engine sends ICMP echoes (or TCP connects where ICMP is unavailable), in parallel.

## L.3 What the numbers mean

| Figure | Meaning | Good |
|---|---|---|
| **Loss** | Packets sent minus received, as a percentage. On a device test it is *query* loss, because a browser cannot send a packet. | 0% |
| **Round trip** | Average, fastest and slowest response time, in ms. | Under the site's target |
| **Jitter** | RFC 3550 interarrival jitter — how much *consecutive* round trips differ. Voice cares about this more than raw latency. | Low single digits of ms |
| **MOS** | Mean Opinion Score (ITU-T G.107 E-model): how a phone call would sound over this path, 1–5. | Above 4.0; below 3.6 is noticed on a call |
| **SLA** | **OK**, **Degraded** (within 80% of a threshold) or **Breached**, against the site's targets. | OK |

## L.4 IPv4, IPv6 and "not counted"

A site whose name resolves to both an IPv4 and an IPv6 address is measured over **both**. The IPv6 result is shown, kept and exported, but marked **not counted**: the site's totals, alerts and SLA availability are carried by IPv4, because the customer's agreement is for the service and IPv4 still carries it. A site that is IPv6 *only* counts over IPv6 — otherwise it could never fail. When the server itself has no IPv6 route, an IPv6 address is reported as *could not be tested from here* rather than as an outage.

## L.5 Two vantage points

**Run diagnostic** has two modes. **From the server** measures from PacketPulse's data centre. **From this device** measures from wherever the person is, over their own connection — the right tool for "is it slow for me?". The two will not match, and both are correct: they measure different layers from different places. A device test can be **attached** to a ticket, where it sits beside the server's figures, labelled as measured on a device.

## L.6 Two-step sign-in

After the password, everyone enters a one-time **6-digit code**. It is **emailed** to their sign-in address, or **texted** to their mobile when an administrator chose text for them and the organisation's own SMS gateway is on. An owner's code is always emailed, so the account that runs the organisation never depends on a gateway. A code is good for ten minutes and five tries; ten wrong codes in all lock the account for fifteen minutes, and nobody is sent more than ten codes an hour. A wrong *password* never locks anything, so a stranger cannot lock someone out by typing their address. A lost phone is handled by an administrator, who signs that device out and switches the person to email.

## L.7 Location at sign-in

For people set to **Record location**, the device's position is recorded when they sign in and out — with the browser's or phone's permission, and a refusal recorded as one. A field engineer's appear on **Check-ins**; an administrator's beside their entries in **Activity**. An organisation can make sharing a position a condition of signing in. The address is looked up afterwards (Google, or OpenStreetMap when no key is set), so it may appear a moment later; the position is the record.

## L.8 The activity trail

Every successful change — who, what, from where, when — is written to a **hash chain**: each entry's hash covers the one before, so altering or removing any entry breaks every entry after it, and **Verify** names the first that does not follow. Administrators' sign-ins and sign-outs are in it, and so are **failed** sign-ins to administrators' accounts (at most twenty an hour per account). A position is shown beside an entry but never sealed into its hash: it is personal data that may have to be erased, and a chain entry never can be.

## L.9 Results leaving PacketPulse

| Way out | Who uses it | What |
|---|---|---|
| **Export PDF** | A person, from a ticket | The evidence document for the ticket or the customer |
| **Export CSV** | A person, from a ticket | The same results as a spreadsheet, one row per site and family |
| **Results API** | A machine with an API key | One ticket by TT number (JSON), or every result of up to 31 days as CSV |
| **Result export** | PacketPulse, on a schedule | A CSV file per period, delivered to the organisation's own SFTP, FTPS or FTP server |

All CSV comes from one writer with one header — the contract an IT system's importer is written against (`packetpulsetest/contracts/result_export_columns.json`).

---
