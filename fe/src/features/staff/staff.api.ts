import { MembershipStatus, type BuiltinRole } from '@/constants/domain'
import { http } from '@/lib/http'
import type { Staff, UserStatus } from '@/features/auth/auth.types'

export interface StaffPayload { email: string; full_name: string; status: UserStatus }
export interface CreateStaffPayload {
  email: string
  full_name: string
  password: string
  is_admin: boolean
  role_id: string
}

const path = '/api/v1/admin/staff'
export const staffApi = {
  async list() {
    const { data } = await http.get<{ staff?: Staff[] }>(`${path}/`, { params: { limit: 100 } })
    return data.staff ?? []
  },
  async create(payload: CreateStaffPayload) {
    const { data } = await http.post<Staff>(`${path}/`, payload)
    return data
  },
  async update(id: string, payload: StaffPayload) {
    const { data } = await http.put<Staff>(`${path}/${id}`, payload)
    return data
  },
  async remove(id: string) { await http.delete(`${path}/${id}`) }
}

export interface TenantMember {
  organization_id: string
  user_id: string
  email: string
  full_name: string
  role: Exclude<BuiltinRole, 'customer' | 'admin'>
  status: MembershipStatus
  venue_ids?: string[]
}

export const tenantStaffApi = {
  async list(organizationId: string) {
    const { data } = await http.get<{ items?: TenantMember[] }>(`/api/v1/tenant/${organizationId}/staff`)
    return data.items ?? []
  },
  async add(organizationId: string, email: string, role: 'tenant_staff' | 'venue_staff') {
    await http.post(`/api/v1/tenant/${organizationId}/staff`, { email, role })
  },
  async update(member: TenantMember, status: TenantMember['status']) {
    await http.put(`/api/v1/tenant/${member.organization_id}/staff/${member.user_id}`, { role: member.role, status })
  },
  async remove(member: TenantMember) {
    await http.delete(`/api/v1/tenant/${member.organization_id}/staff/${member.user_id}`)
  },
  async assignVenue(venueId: string, userId: string, active: boolean) {
    await http.put(`/api/v1/tenant/venues/${venueId}/staff/${userId}`, { status: active ? 'ACTIVE' : 'INACTIVE' })
  }
}
