import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@/app/App'
import { portal } from '@/app/portal'
import '@/styles/global.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App portal={portal} />
  </StrictMode>
)
