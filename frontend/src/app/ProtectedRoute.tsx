import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '@/features/auth/AuthProvider'

export function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  // Wait for the boot-time silent refresh (see auth.tsx) to resolve before
  // deciding — otherwise every hard refresh would bounce a logged-in user to
  // /login for a frame while the httpOnly cookie is still being exchanged.
  if (isLoading) {
    return <div className="route-loading" aria-hidden="true" />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <Outlet />
}
