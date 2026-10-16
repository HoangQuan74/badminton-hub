# Use Case Diagram

The design with actors and their system actions is called a **Use Case Diagram**.

```mermaid
flowchart LR
    Customer["Customer"]
    Owner["Venue Owner"]
    Staff["Venue Staff"]
    Admin["Platform Admin"]

    UC1((Search venues))
    UC2((View availability))
    UC3((Book and pay))
    UC4((Cancel booking))
    UC5((Manage venues and courts))
    UC6((Manage pricing and schedules))
    UC7((Create walk-in booking))
    UC8((Check in customer))
    UC9((Approve organizations and venues))
    UC10((Manage payouts and failed events))

    Customer --> UC1
    Customer --> UC2
    Customer --> UC3
    Customer --> UC4

    Owner --> UC5
    Owner --> UC6
    Owner --> UC10

    Staff --> UC2
    Staff --> UC7
    Staff --> UC8

    Admin --> UC9
    Admin --> UC10
```

## Access Boundary

- Customers access only their bookings.
- Owners access only their organizations.
- Staff access only assigned venues.
- Admins access platform-wide operations.
