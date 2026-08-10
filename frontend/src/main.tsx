import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './i18n'
import App from './App'
import { ThemeProvider } from '@/lib/theme'
import { AuthProvider } from '@/features/auth/AuthContext'
import { DormitoryProvider } from '@/features/dormitory/DormitoryContext'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider>
      <AuthProvider>
        <DormitoryProvider>
          <App />
        </DormitoryProvider>
      </AuthProvider>
    </ThemeProvider>
  </StrictMode>,
)
