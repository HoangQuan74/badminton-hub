import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react-swc'
import { defineConfig } from 'vite'

export default defineConfig(({ mode }) => {
  const portal = mode === 'tenant' || mode === 'admin' ? mode : 'user'
  const port = portal === 'tenant' ? 5174 : portal === 'admin' ? 5175 : 5173
  const apiURL = portal === 'admin' ? 'http://localhost:18082' : 'http://localhost:18080'

  return {
    plugins: [react()],
    resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
    server: { port },
    define: {
      'import.meta.env.VITE_PORTAL': JSON.stringify(portal),
      'import.meta.env.VITE_API_URL': JSON.stringify(apiURL),
      'import.meta.env.VITE_USER_URL': JSON.stringify('http://localhost:5173'),
      'import.meta.env.VITE_TENANT_URL': JSON.stringify('http://localhost:5174'),
      'import.meta.env.VITE_ADMIN_URL': JSON.stringify('http://localhost:5175')
    }
  }
})
