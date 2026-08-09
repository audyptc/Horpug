import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { Suspense, lazy } from 'react'
import type { ReactNode } from 'react'
import { NotFound } from '@/components/shared/NotFound'
import { useAuth } from '@/features/auth/AuthContext'

const Layout = lazy(() => import('@/components/layout/Layout').then((m) => ({ default: m.Layout })))
const Login = lazy(() => import('@/features/auth/Login').then((m) => ({ default: m.Login })))
const Dashboard = lazy(() => import('@/features/dashboard/Dashboard').then((m) => ({ default: m.Dashboard })))
const Users = lazy(() => import('@/features/users/Users').then((m) => ({ default: m.Users })))
const Analytics = lazy(() => import('@/features/analytics/Analytics').then((m) => ({ default: m.Analytics })))
const Settings = lazy(() => import('@/features/settings/Settings').then((m) => ({ default: m.Settings })))
const RoomTypes = lazy(() => import('@/features/roomTypes/RoomTypes').then((m) => ({ default: m.RoomTypes })))
const Roles = lazy(() => import('@/features/roles/Roles').then((m) => ({ default: m.Roles })))
const Rooms = lazy(() => import('@/features/rooms/Rooms').then((m) => ({ default: m.Rooms })))
const Tenants = lazy(() => import('@/features/tenants/Tenants').then((m) => ({ default: m.Tenants })))
const Contracts = lazy(() => import('@/features/contracts/Contracts').then((m) => ({ default: m.Contracts })))
const ElectricMeters = lazy(() => import('@/features/electric-meters/ElectricMeters').then((m) => ({ default: m.ElectricMeters })))
const WaterMeters = lazy(() => import('@/features/water-meters/WaterMeters').then((m) => ({ default: m.WaterMeters })))
const Bills = lazy(() => import('@/features/billing/Bills').then((m) => ({ default: m.Bills })))
const Expenses = lazy(() => import('@/features/expenses/Expenses').then((m) => ({ default: m.Expenses })))
const MaintenanceRequests = lazy(() => import('@/features/maintenance/MaintenanceRequests').then((m) => ({ default: m.MaintenanceRequests })))
const Payments = lazy(() => import('@/features/payments/Payments').then((m) => ({ default: m.Payments })))
const Announcements = lazy(() => import('@/features/announcements/Announcements').then((m) => ({ default: m.Announcements })))
const Reports = lazy(() => import('@/features/reports/Reports').then((m) => ({ default: m.Reports })))
const ActivityLogs = lazy(() => import('@/features/activity-logs/ActivityLogs').then((m) => ({ default: m.ActivityLogs })))
const Parking = lazy(() => import('@/features/parking/Parking').then((m) => ({ default: m.Parking })))
const Parcels = lazy(() => import('@/features/parcels/Parcels').then((m) => ({ default: m.Parcels })))
const Documents = lazy(() => import('@/features/documents/Documents').then((m) => ({ default: m.Documents })))

function withSuspense(element: ReactNode) {
  return (
    <Suspense fallback={<div className="p-6 text-sm text-muted-foreground">Loading...</div>}>
      {element}
    </Suspense>
  )
}

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
  if (canAccessPath(allowedPaths, '/')) return withSuspense(<Dashboard />)

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
    children: [{ path: '/login', element: withSuspense(<Login />) }],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: '/',
        element: withSuspense(<Layout />),
        children: [
          { index: true, element: <RootIndexRoute /> },
          { path: 'rooms', element: <PermissionRoute requiredPath="/rooms" element={withSuspense(<Rooms />)} /> },
          { path: 'tenants', element: <PermissionRoute requiredPath="/tenants" element={withSuspense(<Tenants />)} /> },
          { path: 'contracts', element: <PermissionRoute requiredPath="/contracts" element={withSuspense(<Contracts />)} /> },
          { path: 'electric-meters', element: <PermissionRoute requiredPath="/electric-meters" element={withSuspense(<ElectricMeters />)} /> },
          { path: 'water-meters', element: <PermissionRoute requiredPath="/water-meters" element={withSuspense(<WaterMeters />)} /> },
          { path: 'bills', element: <PermissionRoute requiredPath="/bills" element={withSuspense(<Bills />)} /> },
          { path: 'expenses', element: <PermissionRoute requiredPath="/expenses" element={withSuspense(<Expenses />)} /> },
          { path: 'maintenance', element: <PermissionRoute requiredPath="/maintenance" element={withSuspense(<MaintenanceRequests />)} /> },
          { path: 'payments', element: <PermissionRoute requiredPath="/payments" element={withSuspense(<Payments />)} /> },
          { path: 'announcements', element: <PermissionRoute requiredPath="/announcements" element={withSuspense(<Announcements />)} /> },
          { path: 'parking', element: <PermissionRoute requiredPath="/parking" element={withSuspense(<Parking />)} /> },
          { path: 'parcels', element: <PermissionRoute requiredPath="/parcels" element={withSuspense(<Parcels />)} /> },
          { path: 'documents', element: <PermissionRoute requiredPath="/documents" element={withSuspense(<Documents />)} /> },
          { path: 'analytics', element: <PermissionRoute requiredPath="/analytics" element={withSuspense(<Analytics />)} /> },
          { path: 'reports', element: <PermissionRoute requiredPath="/reports" element={withSuspense(<Reports />)} /> },
          { path: 'activity-logs', element: <PermissionRoute requiredPath="/activity-logs" element={withSuspense(<ActivityLogs />)} /> },
          {
            path: 'settings',
            children: [
              { index: true, element: <SettingsIndexRedirect /> },
              { path: 'users', element: <PermissionRoute requiredPath="/settings/users" element={withSuspense(<Users />)} /> },
              { path: 'roles', element: <PermissionRoute requiredPath="/settings/roles" element={withSuspense(<Roles />)} /> },
              { path: 'general', element: <PermissionRoute requiredPath="/settings/general" element={withSuspense(<Settings />)} /> },
              { path: 'room-types', element: <PermissionRoute requiredPath="/settings/room-types" element={withSuspense(<RoomTypes />)} /> },
            ],
          },
        ],
      },
    ],
  },
  { path: '*', element: <NotFound /> },
])
