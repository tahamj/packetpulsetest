# 📡 PacketPulse Manual Testing & Knowledge Documentation

The human-facing counterpart to the automated suites driven by `packetpulsetest.sh`: how to test, use, sell, operate and release PacketPulse.

---

## What is in here

| File | Audience | Purpose |
|---|---|---|
| **[`MANUAL_TESTING_GUIDE.md`](MANUAL_TESTING_GUIDE.md)** | Sales, operators, QA, backend, Flutter, security | **The 4-in-1 manual (`v2026.10-PROD-v2`)** — **generated from `parts/`; do not hand-edit**. Every feature chapter carries four lenses: 🌟 commercial pitch, 📖 user guide, 🧪 manual testing playbook (with the 🛑 *must not happen* assertions), ⚙️ developer guide. Plus the concepts, end-to-end journeys, the non-functional matrix, release governance, and appendices of every route, capability, migration and setting. |
| **[`manual-testing-guide.html`](manual-testing-guide.html)** | Same | The styled edition: contents sidebar with a working filter, light/dark theme, print to PDF. **Generated** — never hand-edit it. |
| **[`parts/`](parts/)** | Authors | **The source.** Each part of the guide is its own file, numbered in reading order. |
| **[`build/build_guide.py`](build/build_guide.py)** | Authors | The generator. Run it after editing anything under `parts/`. |

## Which path do I follow?

```
Selling PacketPulse, or onboarding a customer?   → Part L, then each chapter's 🌟 Commercial lens, then JRN-001
Using PacketPulse in a NOC or in the field?      → Part L, then each chapter's 📖 User Guide lens
Testing a release?                          → Part 0, Part II test tables, Part III, Part IV, Part V §5.3
Changing the code?                          → each chapter's ⚙️ Developer lens, Appendices A–D
Shipping it?                                → Part V
```

## Editing the guide

```bash
# once: the generator's one dependency, in packetpulsetest/.venv
python3 -m venv packetpulsetest/.venv
packetpulsetest/.venv/bin/pip install -r packetpulsetest/manual-testing/build/requirements.txt

# after every edit under parts/
packetpulsetest/.venv/bin/python3 packetpulsetest/manual-testing/build/build_guide.py

# prove it still matches the code
./packetpulsetest.sh docs
```

`./packetpulsetest.sh docs` creates the virtual environment itself when it is missing. It fails when the generated editions are stale, when a cited repository path does not exist, when a number in the front matter no longer matches source, when Appendix A, B or C drifts from the routes, capabilities or migrations, when a chapter is missing a lens, or when two test cases share an ID.

## Ground rules

1. **Never test against live customer data.** A disposable stack (Part 0 §0.3) for every pass; the live demo organisation for read-only walkthroughs only.
2. **Tenancy needs two organisations.** Cross-tenant cases prove nothing from one account.
3. **Assert the negative.** Each case exists for its 🛑 *must not happen*; a happy path rarely fails alone.
4. **Quote the test ID** in every bug report. IDs are never reused.
5. **Edit `parts/`, never the generated editions**, and run `./packetpulsetest.sh docs` before committing.
