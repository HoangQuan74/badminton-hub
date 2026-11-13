import { http } from '@/lib/http'

export interface RolePayload { name: string; description: string }
export interface Role extends RolePayload { id: string; created_at: string }

const path = '/api/v1/admin/roles'
export const roleApi = {
  async list() {
    const { data } = await http.get<{ roles?: Role[] }>(`${path}/`)
    return data.roles ?? []
  },
  async create(payload: RolePayload) {
    const { data } = await http.post<{ role: Role }>(`${path}/`, payload)
    return data.role
  },
  async update(id: string, payload: RolePayload) {
    const { data } = await http.put<{ role: Role }>(`${path}/${id}`, payload)
    return data.role
  },
  async remove(id: string) { await http.delete(`${path}/${id}`) }
}
