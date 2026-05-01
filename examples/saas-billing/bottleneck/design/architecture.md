# Architecture

### DESIGN-001: Tokenized billing retry flow
Refs:
- INTENT-001
- BEHAVIOR-001
- BEHAVIOR-002
- BEHAVIOR-003

The subscription billing service receives a payment method update from the account settings page, exchanges card details for a payment provider token, stores the token on the active subscription, and emits a billing event. If an invoice is past due, the invoice retry worker uses an idempotency key derived from the customer ID and invoice ID before calling the payment provider.

Key components:
- Account settings payment method form
- Payment provider tokenization API
- Subscription billing service
- Invoice retry worker
- Billing event log
- Duplicate-charge guard based on idempotency keys
