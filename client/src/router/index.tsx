import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/pages/Dashboard'
import { Users } from '@/pages/Users'
import { Analytics } from '@/pages/Analytics'
import { Settings } from '@/pages/Settings'
import { Roles } from '@/pages/Roles'
import { Rooms } from '@/pages/Rooms'
import { Tenants } from '@/pages/Tenants'
import { Contracts } from '@/pages/Contracts'
import { MeterReadings } from '@/pages/MeterReadings'
import { Bills } from '@/pages/Bills'
import { Expenses } from '@/pages/Expenses'
import { MaintenanceRequests } from '@/pages/MaintenanceRequests'
import { Payments } from '@/pages/Payments'
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
          { path: 'rooms', element: <Rooms /> },
          { path: 'tenants', element: <Tenants /> },
          { path: 'contracts', element: <Contracts /> },
          { path: 'meters', element: <MeterReadings /> },
          { path: 'bills', element: <Bills /> },
          { path: 'expenses', element: <Expenses /> },
          { path: 'maintenance', element: <MaintenanceRequests /> },
          { path: 'payments', element: <Payments /> },
          { path: 'analytics', element: <Analytics /> },
          {
            path: 'settings',
            children: [
              { index: true, element: <Navigate to="users" replace /> },
              { path: 'users', element: <Users /> },
              { path: 'roles', element: <Roles /> },
              { path: 'general', element: <Settings /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFound /> },
])
