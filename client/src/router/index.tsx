import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { Dashboard } from '@/features/dashboard/Dashboard'
import { Users } from '@/features/users/Users'
import { Analytics } from '@/features/analytics/Analytics'
import { Settings } from '@/features/settings/Settings'
import { RoomTypes } from '@/features/roomTypes/RoomTypes'
import { Roles } from '@/features/roles/Roles'
import { Rooms } from '@/features/rooms/Rooms'
import { Tenants } from '@/features/tenants/Tenants'
import { Contracts } from '@/features/contracts/Contracts'
import { MeterReadings } from '@/features/meters/MeterReadings'
import { Bills } from '@/features/billing/Bills'
import { Expenses } from '@/features/expenses/Expenses'
import { MaintenanceRequests } from '@/features/maintenance/MaintenanceRequests'
import { Payments } from '@/features/payments/Payments'
import { Announcements } from '@/features/announcements/Announcements'
import { Reports } from '@/features/reports/Reports'
import { ActivityLogs } from '@/features/activity-logs/ActivityLogs'
import { Parking } from '@/features/parking/Parking'
import { Parcels } from '@/features/parcels/Parcels'
import { Documents } from '@/features/documents/Documents'
import { NotFound } from '@/components/shared/NotFound'
import { Login } from '@/features/auth/Login'
import { useAuth } from '@/features/auth/AuthContext'

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
          { path: 'announcements', element: <Announcements /> },
          { path: 'parking', element: <Parking /> },
          { path: 'parcels', element: <Parcels /> },
          { path: 'documents', element: <Documents /> },
          { path: 'analytics', element: <Analytics /> },
          { path: 'reports', element: <Reports /> },
          { path: 'activity-logs', element: <ActivityLogs /> },
          {
            path: 'settings',
            children: [
              { index: true, element: <Navigate to="users" replace /> },
              { path: 'users', element: <Users /> },
              { path: 'roles', element: <Roles /> },
              { path: 'general', element: <Settings /> },
              { path: 'room-types', element: <RoomTypes /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFound /> },
])
