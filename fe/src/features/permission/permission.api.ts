import { PermissionAction } from '@/constants/domain'
import { http } from '@/lib/http'

export const permissionActions = Object.values(PermissionAction)
export interface PermissionPayload {
  module: string
  path: string
  action: typeof permissionActions[number]
  description: string
}
export interface Permission extends PermissionPayload { id: string }

const path = '/api/v1/admin/permissions'
export const permissionApi = {
  async list() {
    const { data } = await http.get<{ permissions?: Permission[] }>(`${path}/`)
    return data.permissions ?? []
  },
  async create(payload: PermissionPayload) {
    const { data } = await http.post<{ permission: Permission }>(`${path}/`, payload)
    return data.permission
  },
  async update(id: string, payload: PermissionPayload) {
    const { data } = await http.put<{ permission: Permission }>(`${path}/${id}`, payload)
    return data.permission
  },
  async remove(id: string) { await http.delete(`${path}/${id}`) }
}
