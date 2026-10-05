# PacketPulse Guard Test Suite (packetpulsetest)

Integration and conformance test suites verifying multi-tenant isolation, permission enforcement (ACL), immutable audit logging, API contracts, and localization coverage.

## Test Suites

- **tenancyisolation**: Strict tenant isolation across all tables.
- **tenancyassignment**: Organisation creation, holding state, and seat limits.
- **aclconformance**: Role-based access control and privilege boundaries.
- **auditconformance**: Audit log integrity and cryptographic hash chain verification.
- **translationcoverage**: Verification that all 23 language translations are complete and covered.
- **apicontract**: Conformance against REST API specifications.
- **loadtest** (opt-in, `./packetpulsetest.sh load`): one licensed organisation under load - twenty people signing in at once, a traced sweep of 200 sites, the Results API and the CSV export - against p95 budgets, with no 5xx and every result stored.

## Running

Tests are executed using the master test runner script:

```bash
./packetpulsetest.sh
```
