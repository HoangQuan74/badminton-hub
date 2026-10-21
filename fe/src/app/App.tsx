import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router'
import { AuthBootstrap } from '@/features/auth/components/AuthBootstrap'
import { queryClient } from '@/lib/query-client'
import { I18nProvider } from '@/lib/i18n-provider'
import { router } from '@/router'
import type { Portal } from './portal'

export function App({ portal }: { portal: Portal }) {
  return (
    <I18nProvider>
      <QueryClientProvider client={queryClient}>
        <AuthBootstrap>
          <RouterProvider router={router(portal)} />
        </AuthBootstrap>
      </QueryClientProvider>
    </I18nProvider>
  )
}
