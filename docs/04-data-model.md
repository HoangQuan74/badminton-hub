# Data Model

## Conventions

- UUID primary keys.
- UTC timestamps in PostgreSQL.
- `created_at` and `updated_at` on mutable business entities.
- Monetary values stored as integer VND amounts.
- Tenant-owned rows include `organization_id` where practical.
- Soft deletion only where business history must remain accessible.

## Core Entities

### users

- `id`
- `email`
- `password_hash`
- `full_name`
- `phone`
- `status`
- `created_at`, `updated_at`

### organizations

- `id`
- `name`
- `slug`
- `status`
- `commission_rate_bps`
- `created_at`, `updated_at`

### organization_members

- `organization_id`
- `user_id`
- `role`: owner or staff
- `status`
- unique `(organization_id, user_id)`

### venues

- `id`
- `organization_id`
- `name`
- `slug`
- `timezone`
- `address`
- `latitude`, `longitude`
- `status`
- `cancellation_policy_id`

### courts

- `id`
- `organization_id`
- `venue_id`
- `name`
- `status`
- `slot_interval_minutes`

### operating_hours

- `id`
- `venue_id`
- `day_of_week`
- `opens_at`
- `closes_at`
- unique `(venue_id, day_of_week)`

### pricing_rules

- `id`
- `organization_id`
- `venue_id`
- optional `court_id`
- `day_type`
- `start_time`, `end_time`
- `price_per_hour_vnd`
- `priority`
- `effective_from`, optional `effective_to`

### court_blocks

- `id`
- `organization_id`
- `venue_id`
- `court_id`
- `starts_at`, `ends_at`
- `reason`
- `created_by`

### bookings

- `id`
- `organization_id`
- `venue_id`
- `court_id`
- optional `customer_id`
- `source`: online or walk_in
- `status`
- `starts_at`, `ends_at`
- optional `hold_expires_at`
- `base_amount_vnd`
- `discount_amount_vnd`
- `platform_fee_vnd`
- `venue_receivable_vnd`
- `total_amount_vnd`
- `pricing_snapshot` JSONB
- optional `contact_name`, `contact_phone`
- `created_by`
- `created_at`, `updated_at`

### booking_transitions

- `id`
- `booking_id`
- `from_status`
- `to_status`
- `reason`
- `actor_id`
- `created_at`

### payments

- `id`
- `booking_id`
- `status`
- `amount_vnd`
- `provider`
- `provider_transaction_id`
- `created_at`, `updated_at`
- unique `(provider, provider_transaction_id)`

### payment_events

- `id`
- `provider`
- `provider_event_id`
- `payload_hash`
- `processed_at`
- unique `(provider, provider_event_id)`

### refunds

- `id`
- `booking_id`
- `payment_id`
- `status`
- `amount_vnd`
- `reason`
- `created_at`, `updated_at`

### payouts

- `id`
- `organization_id`
- `period_start`, `period_end`
- `gross_amount_vnd`
- `refund_amount_vnd`
- `platform_fee_vnd`
- `net_amount_vnd`
- `status`
- `created_at`, `updated_at`

### idempotency_keys

- `id`
- `actor_id`
- `scope`
- `key`
- `request_hash`
- `response_status`
- `response_body` JSONB
- `expires_at`
- unique `(actor_id, scope, key)`

### outbox_events

- `id`
- `aggregate_type`
- `aggregate_id`
- `event_type`
- `payload` JSONB
- `status`
- `attempts`
- `available_at`
- optional `published_at`

### processed_events

- `consumer_name`
- `event_id`
- `processed_at`
- unique `(consumer_name, event_id)`

### audit_logs

- `id`
- optional `organization_id`
- `actor_id`
- `action`
- `entity_type`
- `entity_id`
- `metadata` JSONB
- `created_at`

## Important Constraints

- Unique user email.
- Unique organization slug.
- Unique public venue slug.
- `starts_at < ends_at` for bookings and blocks.
- Non-negative monetary values.
- Valid booking-state transitions enforced by application logic and tested.
- PostgreSQL exclusion constraint prevents overlapping active reservations for the same court.
- Tenant repository methods require organization context for owner/staff operations.
