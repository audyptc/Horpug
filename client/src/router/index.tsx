import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Users } from '@/pages/Users'
import { Analytics } from '@/pages/Analytics'
import { Settings } from '@/pages/Settings'
import { Placeholder } from '@/pages/Placeholder'
import { NotFound } from '@/pages/NotFound'
import { Login } from '@/pages/Login'
import { useAuth } from '@/context/AuthContext'

function ProtectedRoute() {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <Outlet />
}

function GuestRoute() {
  const { isAuthenticated } = useAuth()
  if (isAuthenticated) return <Navigate to="/" replace />
  return <Outlet />
}

export const router = createBrowserRouter([
  {
    element: <GuestRoute />,
    children: [{ path: '/login', element: <Login /> }],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: '/',
        element: <Layout />,
        children: [
          { index: true, element: <Dashboard /> },
          { path: 'analytics', element: <Analytics /> },
          { path: 'notifications', element: <Placeholder title="Notifications" /> },
          {
            path: 'settings',
            children: [
              { index: true, element: <Navigate to="general" replace /> },
              { path: 'general', element: <Settings /> },
              { path: 'roles', element: <Placeholder title="Roles & Permissions" /> },
              { path: 'users', element: <Users /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFound /> },
])
