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

## Frontend and customer links

- Deliver owner onboarding through the frontend together with the Identity API.
- A tenant shares `https://<user-domain>/t/{slug}` with customers;
  an organization may contain multiple venues with their own detail links.
- Guests can browse an active tenant's published venues. Login or registration
  returns them to the selected tenant/venue; customer signup grants no staff access.
- Availability, reservation, and checkout remain Booking-owned capabilities.

See [Owner Onboarding and Tenant Share Links](../design/07-owner-onboarding-and-tenant-links.md)
for planned folders/files and delivery boundaries.
