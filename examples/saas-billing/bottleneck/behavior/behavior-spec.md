# Behavior Specification

## Expected Behavior

### BEHAVIOR-001: Customer updates payment method
Critical: true
Refs:
- INTENT-001
- ASSURANCE-001

When a customer replaces an expired card, the subscription billing service stores only the payment provider token, updates the active subscription payment method, and preserves the existing invoice state.

### BEHAVIOR-002: Failed invoice is retried after payment method update
Critical: true
Refs:
- INTENT-001
- ASSURANCE-002

When an invoice is past due and the customer updates their payment method, the billing service retries the failed invoice exactly once and records the retry attempt on the invoice timeline.

### BEHAVIOR-003: Payment retry does not create duplicate charges
Critical: true
Refs:
- INTENT-001

When the retry request is submitted more than once for the same invoice, the billing service uses the existing idempotency key and returns the original retry result instead of creating a second charge.

## Unacceptable Behavior

- The system must not store raw card numbers, CVV values, or payment details outside the payment provider token.
- The system must not create duplicate charges for the same failed invoice retry.
- The system must not mark an invoice paid unless the payment provider confirms the charge succeeded.
