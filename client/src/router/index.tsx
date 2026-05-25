import { createBrowserRouter, Navigate } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Users } from '@/pages/Users'
import { Analytics } from '@/pages/Analytics'
import { Settings } from '@/pages/Settings'
import { Placeholder } from '@/pages/Placeholder'
import { NotFound } from '@/pages/NotFound'

export const router = createBrowserRouter([
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
  { path: '*', element: <NotFound /> },
])
