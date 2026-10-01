# Pingle Guard Test Suite (pingletest)

Integration and conformance test suites verifying multi-tenant isolation, permission enforcement (ACL), immutable audit logging, API contracts, and localization coverage.

## Test Suites

- **tenancyisolation**: Strict tenant isolation across all tables.
- **tenancyassignment**: Organisation creation, holding state, and seat limits.
- **aclconformance**: Role-based access control and privilege boundaries.
- **auditconformance**: Audit log integrity and cryptographic hash chain verification.
- **translationcoverage**: Verification that all 23 language translations are complete and covered.
- **apicontract**: Conformance against REST API specifications.

## Running

Tests are executed using the master test runner script:

```bash
./pingletest.sh
```
