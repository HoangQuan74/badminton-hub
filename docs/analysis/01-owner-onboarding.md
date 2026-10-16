# Owner Onboarding Flow

## Actor

Venue Owner, Platform Admin.

## Flow

1. The owner registers an account.
2. The owner creates an organization.
3. The organization enters `PENDING_VERIFICATION`.
4. The admin reviews the organization.
5. The admin approves or rejects it.
6. An approved organization becomes `ACTIVE`.
7. The owner can create venues and invite staff.

## Rules

- Only active organizations can publish venues.
- Organization names and public slugs must be valid and unique.
- Every admin decision creates an audit log.
- Staff access requires an active organization membership.
