import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import type { ReactNode } from 'react'
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
import { ElectricMeters } from '@/features/electric-meters/ElectricMeters'
import { WaterMeters } from '@/features/water-meters/WaterMeters'
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

const ROUTE_CANDIDATES = [
  '/',
  '/rooms',
  '/tenants',
  '/contracts',
  '/electric-meters',
  '/water-meters',
  '/bills',
  '/payments',
  '/expenses',
  '/maintenance',
  '/announcements',
  '/parking',
  '/parcels',
  '/documents',
  '/analytics',
  '/reports',
  '/activity-logs',
  '/settings/users',
  '/settings/roles',
  '/settings/room-types',
  '/settings/general',
] as const

function normalizePath(path: string): string {
  const base = path.split(/[?#]/, 1)[0].trim()
  if (!base || base === '/') return '/'
  const collapsed = base.replace(/\/+$/g, '')
  if (!collapsed) return '/'
  return collapsed.startsWith('/') ? collapsed : `/${collapsed}`
}

function canAccessPath(allowedPaths: Set<string> | null, path: string): boolean {
  if (allowedPaths === null) return true
  return allowedPaths.has(normalizePath(path))
}

function getFirstAccessiblePath(allowedPaths: Set<string> | null, excludeRoot = false): string | null {
  const candidates = excludeRoot ? ROUTE_CANDIDATES.filter((path) => path !== '/') : ROUTE_CANDIDATES

  for (const path of candidates) {
    if (canAccessPath(allowedPaths, path)) {
      return path
    }
  }

  if (allowedPaths && allowedPaths.size > 0) {
    return Array.from(allowedPaths)[0] ?? null
  }

  return null
}

function ProtectedRoute() {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <Outlet />
}

function GuestRoute() {
  const { isAuthenticated, allowedPaths } = useAuth()
  if (isAuthenticated) {
    const redirectTo = getFirstAccessiblePath(allowedPaths) ?? '/'
    return <Navigate to={redirectTo} replace />
  }
  return <Outlet />
}

function PermissionRoute({ requiredPath, element }: { requiredPath: string; element: ReactNode }) {
  const { allowedPaths } = useAuth()
  if (!canAccessPath(allowedPaths, requiredPath)) return <NotFound />
  return element
}

function RootIndexRoute() {
  const { allowedPaths } = useAuth()
  if (canAccessPath(allowedPaths, '/')) return <Dashboard />

  const redirectTo = getFirstAccessiblePath(allowedPaths, true)
  if (redirectTo) return <Navigate to={redirectTo} replace />

  return <NotFound />
}

function SettingsIndexRedirect() {
  const { allowedPaths } = useAuth()
  const redirectTo = getFirstAccessiblePath(allowedPaths, true)
  if (redirectTo?.startsWith('/settings/')) return <Navigate to={redirectTo.replace('/settings/', '')} replace />

  return <NotFound />
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
          { index: true, element: <RootIndexRoute /> },
          { path: 'rooms', element: <PermissionRoute requiredPath="/rooms" element={<Rooms />} /> },
          { path: 'tenants', element: <PermissionRoute requiredPath="/tenants" element={<Tenants />} /> },
          { path: 'contracts', element: <PermissionRoute requiredPath="/contracts" element={<Contracts />} /> },
          { path: 'electric-meters', element: <PermissionRoute requiredPath="/electric-meters" element={<ElectricMeters />} /> },
          { path: 'water-meters', element: <PermissionRoute requiredPath="/water-meters" element={<WaterMeters />} /> },
          { path: 'bills', element: <PermissionRoute requiredPath="/bills" element={<Bills />} /> },
          { path: 'expenses', element: <PermissionRoute requiredPath="/expenses" element={<Expenses />} /> },
          { path: 'maintenance', element: <PermissionRoute requiredPath="/maintenance" element={<MaintenanceRequests />} /> },
          { path: 'payments', element: <PermissionRoute requiredPath="/payments" element={<Payments />} /> },
          { path: 'announcements', element: <PermissionRoute requiredPath="/announcements" element={<Announcements />} /> },
          { path: 'parking', element: <PermissionRoute requiredPath="/parking" element={<Parking />} /> },
          { path: 'parcels', element: <PermissionRoute requiredPath="/parcels" element={<Parcels />} /> },
          { path: 'documents', element: <PermissionRoute requiredPath="/documents" element={<Documents />} /> },
          { path: 'analytics', element: <PermissionRoute requiredPath="/analytics" element={<Analytics />} /> },
          { path: 'reports', element: <PermissionRoute requiredPath="/reports" element={<Reports />} /> },
          { path: 'activity-logs', element: <PermissionRoute requiredPath="/activity-logs" element={<ActivityLogs />} /> },
          {
            path: 'settings',
            children: [
              { index: true, element: <SettingsIndexRedirect /> },
              { path: 'users', element: <PermissionRoute requiredPath="/settings/users" element={<Users />} /> },
              { path: 'roles', element: <PermissionRoute requiredPath="/settings/roles" element={<Roles />} /> },
              { path: 'general', element: <PermissionRoute requiredPath="/settings/general" element={<Settings />} /> },
              { path: 'room-types', element: <PermissionRoute requiredPath="/settings/room-types" element={<RoomTypes />} /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFound /> },
])
