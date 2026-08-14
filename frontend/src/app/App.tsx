import { Navigate, Route, Routes } from 'react-router-dom'
import { ProtectedRoute } from '@/app/ProtectedRoute'
import { AdminLayout } from '@/app/AdminLayout'
import { AuthProvider } from '@/features/auth/AuthProvider'
import LoginPage from '@/features/auth/LoginPage'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import ResourcePage from '@/features/resource/ResourcePage'
import RolePermissionsPage from '@/features/role/RolePermissionsPage'
import DormitoryPage from '@/features/dormitory/DormitoryPage'
import RoomTypePage from '@/features/roomtype/RoomTypePage'
import RoomPage from '@/features/room/RoomPage'
import TenantPage from '@/features/tenant/TenantPage'
import ContractPage from '@/features/contract/ContractPage'
import MeterPage from '@/features/meter/MeterPage'
import WaterMeterPage from '@/features/watermeter/WaterMeterPage'
import InvoicePage from '@/features/invoice/InvoicePage'
import PaymentPage from '@/features/payment/PaymentPage'
import ExpensePage from '@/features/expense/ExpensePage'
import AnnouncementPage from '@/features/announcement/AnnouncementPage'
import RepairRequestPage from '@/features/repairrequest/RepairRequestPage'
import ParkingPage from '@/features/parking/ParkingPage'
import ParcelPage from '@/features/parcel/ParcelPage'
import DocumentPage from '@/features/document/DocumentPage'
import UserPage from '@/features/user/UserPage'
import ActivityLogPage from '@/features/activitylog/ActivityLogPage'
import { menuMeta } from '@/features/menu/menus'

function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route element={<ProtectedRoute />}>
          <Route
            path="/dashboard"
            element={
              <AdminLayout>
                <DashboardPage />
              </AdminLayout>
            }
          />
          <Route
            path="/roles"
            element={
              <AdminLayout>
                <RolePermissionsPage />
              </AdminLayout>
            }
          />
          <Route
            path="/dormitories"
            element={
              <AdminLayout>
                <DormitoryPage />
              </AdminLayout>
            }
          />
          <Route
            path="/room-types"
            element={
              <AdminLayout>
                <RoomTypePage />
              </AdminLayout>
            }
          />
          <Route
            path="/rooms"
            element={
              <AdminLayout>
                <RoomPage />
              </AdminLayout>
            }
          />
          <Route
            path="/tenants"
            element={
              <AdminLayout>
                <TenantPage />
              </AdminLayout>
            }
          />
          <Route
            path="/contracts"
            element={
              <AdminLayout>
                <ContractPage />
              </AdminLayout>
            }
          />
          <Route
            path="/meters"
            element={
              <AdminLayout>
                <MeterPage />
              </AdminLayout>
            }
          />
          <Route
            path="/water-meters"
            element={
              <AdminLayout>
                <WaterMeterPage />
              </AdminLayout>
            }
          />
          <Route
            path="/invoices"
            element={
              <AdminLayout>
                <InvoicePage />
              </AdminLayout>
            }
          />
          <Route
            path="/payments"
            element={
              <AdminLayout>
                <PaymentPage />
              </AdminLayout>
            }
          />
          <Route
            path="/expenses"
            element={
              <AdminLayout>
                <ExpensePage />
              </AdminLayout>
            }
          />
          <Route
            path="/announcements"
            element={
              <AdminLayout>
                <AnnouncementPage />
              </AdminLayout>
            }
          />
          <Route
            path="/repair-requests"
            element={
              <AdminLayout>
                <RepairRequestPage />
              </AdminLayout>
            }
          />
          <Route
            path="/parking"
            element={
              <AdminLayout>
                <ParkingPage />
              </AdminLayout>
            }
          />
          <Route
            path="/parcels"
            element={
              <AdminLayout>
                <ParcelPage />
              </AdminLayout>
            }
          />
          <Route
            path="/documents"
            element={
              <AdminLayout>
                <DocumentPage />
              </AdminLayout>
            }
          />
          <Route
            path="/users"
            element={
              <AdminLayout>
                <UserPage />
              </AdminLayout>
            }
          />
          <Route
            path="/activity-logs"
            element={
              <AdminLayout>
                <ActivityLogPage />
              </AdminLayout>
            }
          />
          {Object.entries(menuMeta)
            .filter(
              ([path]) =>
                path !== '/roles' &&
                path !== '/dormitories' &&
                path !== '/room-types' &&
                path !== '/rooms' &&
                path !== '/tenants' &&
                path !== '/contracts' &&
                path !== '/meters' &&
                path !== '/water-meters' &&
                path !== '/invoices' &&
                path !== '/payments' &&
                path !== '/expenses' &&
                path !== '/announcements' &&
                path !== '/repair-requests' &&
                path !== '/parking' &&
                path !== '/parcels' &&
                path !== '/documents' &&
                path !== '/users' &&
                path !== '/activity-logs'
            )
            .map(([path, meta]) => (
              <Route
                key={path}
                path={path}
                element={
                  <AdminLayout>
                    <ResourcePage titleKey={meta.labelKey} descriptionKey={meta.descriptionKey} endpoint={path} />
                  </AdminLayout>
                }
              />
            ))}
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </AuthProvider>
  )
}

export default App
