import { http } from '@/lib/http'
import type { Tenant, TenantRegistrationPayload } from './tenant.api'

export interface TenantRegistration extends Omit<TenantRegistrationPayload, 'name' | 'slug'> {
  organization: Tenant
  review_decision: string
  review_note: string
  reviewer_id: string
  reviewed_at: string
}

export const tenantReviewApi = {
  async list(status: string) {
    const { data } = await http.get<{ items?: TenantRegistration[] }>('/api/v1/admin/tenants', { params: { status } })
    return data.items ?? []
  },
  async review(id: string, decision: 'approve' | 'reject', note: string) {
    const { data } = await http.post<{ registration: TenantRegistration }>(`/api/v1/admin/tenants/${id}/${decision}`, { note })
    return data.registration
  }
}
