import { CrudManager } from '@/features/admin/components/CrudManager'
import { roleApi } from './role.api'

export function RoleManager() {
  return <CrudManager
    queryKey="roles" title="role.title"
    fields={[{ name: 'name', label: 'crud.name', required: true }, { name: 'description', label: 'crud.description' }]}
    list={roleApi.list} remove={roleApi.remove}
    save={(id, values) => {
      const payload = { name: values.name ?? '', description: values.description ?? '' }
      return id ? roleApi.update(id, payload) : roleApi.create(payload)
    }}
    values={(item) => ({ name: item.name, description: item.description })}
    label={(item) => item.name} summary={(item) => item.description}
  />
}
