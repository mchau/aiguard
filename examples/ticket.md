# Ticket: Add user reason field to refund flow

## Description

When a customer requests a refund, they must provide a reason. The reason should
be visible in the UI, passed to the backend service, persisted to the database,
and recorded in the audit log.

## Acceptance Criteria

- AC1: The refund form must include a "Reason" text field (max 500 chars)
- AC2: The reason must be passed to the refund service in the API request
- AC3: The reason must be persisted to the refunds table in the database
- AC4: The audit log must record the reason alongside the refund transaction ID

## Non-Goals

- Admin users do not need this field in this iteration
- No analytics dashboard for reason categorization

## Definition of Done

- [ ] Unit tests cover all 4 ACs
- [ ] Integration test verifies DB persistence
- [ ] Code reviewed and approved
