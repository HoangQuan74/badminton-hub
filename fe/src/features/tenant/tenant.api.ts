import { BuiltinRole, MembershipStatus, type TenantStatus } from '@/constants/domain'
import { http } from '@/lib/http'
import type { User } from '@/features/auth/auth.types'

export interface Tenant {
  id: string
  name: string
  slug: string
  status: TenantStatus
  created_at: string
  updated_at: string
}

export interface TenantPayload { name: string; slug: string }
export interface TenantRegistrationPayload extends TenantPayload {
 contact_name: string
 contact_email: string
 contact_phone: string
 note: string
}
export interface CreateTenantResponse {
  user: User
  organization: Tenant
  membership: { organization_id: string; user_id: string; role: typeof BuiltinRole.TenantOwner; status: typeof MembershipStatus.Active }
}

export const tenantApi = {
  async list() {
    const { data } = await http.get<{ items?: Tenant[] }>('/api/v1/tenant')
    return data.items ?? []
  },
  async create(payload: TenantRegistrationPayload, key: string) {
    const { data } = await http.post<CreateTenantResponse>('/api/v1/tenant', payload, {
      headers: { 'Idempotency-Key': key }
    })
    return data
  },
  async update(id: string, payload: TenantPayload) {
    const { data } = await http.patch<{ organization: Tenant }>(`/api/v1/tenant/${id}`, payload)
    return data.organization
  },
  async archive(id: string) {
    await http.delete(`/api/v1/tenant/${id}`)
  }
}
