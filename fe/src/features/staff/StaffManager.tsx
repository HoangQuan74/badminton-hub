import { AccountStatus } from '@/constants/domain'
import { useQuery } from '@tanstack/react-query'
import { CrudManager } from '@/features/admin/components/CrudManager'
import { roleApi } from '@/features/role/role.api'
import { useAuthStore } from '@/features/auth/auth.store'
import { useI18n } from '@/lib/use-i18n'
import { staffApi } from './staff.api'

export function StaffManager() {
  const { t } = useI18n()
  const user = useAuthStore((state) => state.user)!
  const roles = useQuery({ queryKey: ['roles'], queryFn: roleApi.list })
  return <CrudManager
    queryKey="staff" title="staff.title" description="staff.description"
    fields={[
      { name: 'full_name', label: 'crud.name', required: true },
      { name: 'email', label: 'crud.email', type: 'email', required: true },
      { name: 'password', label: 'crud.password', type: 'password', required: true, minLength: 8, createOnly: true },
      { name: 'role_id', label: 'crud.role', type: 'select', required: true, createOnly: true, options: roles.data?.map((role) => ({ value: role.id, label: role.name })) },
      { name: 'status', label: 'crud.status', type: 'select', required: true, editOnly: true, options: [{ value: AccountStatus.Active, label: t('crud.active') }, { value: AccountStatus.Suspended, label: t('crud.suspended') }] }
    ]}
    list={staffApi.list} remove={staffApi.remove} canDelete={(item) => item.id !== user.id}
    save={(id, values) => id
      ? staffApi.update(id, { email: values.email ?? '', full_name: values.full_name ?? '', status: values.status === AccountStatus.Suspended ? AccountStatus.Suspended : AccountStatus.Active })
      : staffApi.create({ email: values.email ?? '', full_name: values.full_name ?? '', password: values.password ?? '', is_admin: false, role_id: values.role_id ?? '' })}
    values={(item) => ({ email: item.email, full_name: item.full_name, status: item.status })}
    label={(item) => item.full_name} summary={(item) => `${item.email} · ${item.is_admin ? t('portal.admin') : item.role_name || '—'} · ${item.status === AccountStatus.Active ? t('crud.active') : t('crud.suspended')}`}
  />
}
