import { PermissionAction } from '@/constants/domain'
import { CrudManager } from '@/features/admin/components/CrudManager'
import { permissionActions, permissionApi } from './permission.api'

export function PermissionManager() {
  return <CrudManager
    queryKey="permissions" title="permission.title" description="permission.description"
    fields={[
      { name: 'module', label: 'crud.module', required: true, maxLength: 100 },
      { name: 'path', label: 'crud.path', required: true, maxLength: 300 },
      { name: 'action', label: 'crud.action', type: 'select', required: true, options: permissionActions.map((action) => ({ value: action, label: action })) },
      { name: 'description', label: 'crud.description', maxLength: 500 }
    ]}
    list={permissionApi.list} remove={permissionApi.remove}
    save={(id, values) => {
      const action = permissionActions.find((action) => action === values.action) ?? PermissionAction.Read
      const payload = { module: values.module ?? '', path: values.path ?? '', action, description: values.description ?? '' }
      return id ? permissionApi.update(id, payload) : permissionApi.create(payload)
    }}
    values={(item) => ({ module: item.module, path: item.path, action: item.action, description: item.description })}
    label={(item) => `${item.module} · ${item.action}`} summary={(item) => `${item.path} · ${item.description}`}
  />
}
