import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { lightTheme, darkTheme } from './theme'
import Login from './pages/Login'
import Register from './pages/Register'
import Dashboard from './pages/Dashboard'
import Profile from './pages/Profile'
import Events from './pages/Events'
import EventDetail from './pages/Events/EventDetail'
import Marketplace from './pages/Events/Marketplace'
import Tickets from './pages/Tickets'
import TransferRecords from './pages/Tickets/TransferRecords'
import AdminEvents from './pages/Admin/Events'
import AdminDashboard from './pages/Admin/Dashboard'
import AppLayout from './components/AppLayout'
import ProtectedRoute from './components/ProtectedRoute'
import ErrorBoundary from './components/ErrorBoundary'
import RouteErrorBoundary from './components/RouteErrorBoundary'
import { useAuthStore } from './stores/authStore'

function AdminRoute({ children }: { children: React.ReactNode }) {
  const { user } = useAuthStore()
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

export default function App() {
  const [isDark, setIsDark] = useState(() => {
    return localStorage.getItem('theme') === 'dark'
  })

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light')
    localStorage.setItem('theme', isDark ? 'dark' : 'light')
  }, [isDark])

  return (
    <ErrorBoundary>
      <ConfigProvider locale={zhCN} theme={isDark ? darkTheme : lightTheme}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<RouteErrorBoundary><Login /></RouteErrorBoundary>} />
            <Route path="/register" element={<RouteErrorBoundary><Register /></RouteErrorBoundary>} />
            <Route element={<ProtectedRoute />}>
              <Route element={<AppLayout isDark={isDark} onThemeToggle={() => setIsDark(!isDark)} />}>
                <Route path="/" element={<RouteErrorBoundary><Dashboard /></RouteErrorBoundary>} />
                <Route path="/events" element={<RouteErrorBoundary><Events /></RouteErrorBoundary>} />
                <Route path="/events/:id" element={<RouteErrorBoundary><EventDetail /></RouteErrorBoundary>} />
                <Route path="/marketplace" element={<RouteErrorBoundary><Marketplace /></RouteErrorBoundary>} />
                <Route path="/tickets" element={<RouteErrorBoundary><Tickets /></RouteErrorBoundary>} />
                <Route path="/transfer-records" element={<RouteErrorBoundary><TransferRecords /></RouteErrorBoundary>} />
                <Route path="/profile" element={<RouteErrorBoundary><Profile /></RouteErrorBoundary>} />
                <Route path="/admin/dashboard" element={<RouteErrorBoundary><AdminRoute><AdminDashboard /></AdminRoute></RouteErrorBoundary>} />
                <Route path="/admin/events" element={<RouteErrorBoundary><AdminRoute><AdminEvents /></AdminRoute></RouteErrorBoundary>} />
              </Route>
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </ErrorBoundary>
  )
}
