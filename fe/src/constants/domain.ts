export const AccountStatus = { Active: 'active', Suspended: 'suspended' } as const
export type AccountStatus = typeof AccountStatus[keyof typeof AccountStatus]

export const TenantStatus = {
  PendingVerification: 'PENDING_VERIFICATION', Active: 'ACTIVE', Suspended: 'SUSPENDED', Rejected: 'REJECTED', Archived: 'ARCHIVED'
} as const
export type TenantStatus = typeof TenantStatus[keyof typeof TenantStatus]

export const VenueStatus = {
  Draft: 'DRAFT', PendingApproval: 'PENDING_APPROVAL', Published: 'PUBLISHED', Suspended: 'SUSPENDED', Archived: 'ARCHIVED'
} as const
export type VenueStatus = typeof VenueStatus[keyof typeof VenueStatus]

export const CourtStatus = { Active: 'ACTIVE', Inactive: 'INACTIVE' } as const
export type CourtStatus = typeof CourtStatus[keyof typeof CourtStatus]

export const MembershipStatus = { Active: 'ACTIVE', Inactive: 'INACTIVE' } as const
export type MembershipStatus = typeof MembershipStatus[keyof typeof MembershipStatus]

export const BuiltinRole = {
  Customer: 'customer', TenantOwner: 'tenant_owner', TenantStaff: 'tenant_staff', VenueStaff: 'venue_staff', Admin: 'admin'
} as const
export type BuiltinRole = typeof BuiltinRole[keyof typeof BuiltinRole]

export const PermissionAction = { Create: 'CREATE', Read: 'READ', Update: 'UPDATE', Delete: 'DELETE' } as const
export type PermissionAction = typeof PermissionAction[keyof typeof PermissionAction]
