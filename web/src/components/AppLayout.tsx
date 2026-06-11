import { useState, useEffect, useCallback } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Avatar, Tooltip } from 'antd'
import {
  DashboardOutlined,
  CalendarOutlined,
  ShopOutlined,
  TagOutlined,
  SwapOutlined,
  UserOutlined,
  BarChartOutlined,
  SettingOutlined,
  AuditOutlined,
  PercentageOutlined,
  LogoutOutlined,
  SearchOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'
import NotificationBell from './NotificationBell'
import CommandPalette from './CommandPalette'

interface Props {
  isDark: boolean
  onThemeToggle: () => void
}

export default function AppLayout({ isDark, onThemeToggle }: Props) {
  const [expanded, setExpanded] = useState(false)
  const [showPalette, setShowPalette] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const isAdmin = user?.role === 'admin'

  const navItems = [
    { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/events', icon: <CalendarOutlined />, label: '活动' },
    { key: '/marketplace', icon: <ShopOutlined />, label: '市场' },
    { key: '/tickets', icon: <TagOutlined />, label: '票务' },
    { key: '/transfer-records', icon: <SwapOutlined />, label: '转让' },
    { key: '/profile', icon: <UserOutlined />, label: '个人' },
    ...(isAdmin ? [
      { key: 'divider', icon: null, label: '' },
      { key: '/admin/dashboard', icon: <BarChartOutlined />, label: '数据' },
      { key: '/admin/events', icon: <SettingOutlined />, label: '管理' },
      { key: '/admin/transfers', icon: <AuditOutlined />, label: '审核' },
      { key: '/admin/promo', icon: <PercentageOutlined />, label: '促销' },
    ] : []),
  ]

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      setShowPalette(prev => !prev)
    }
    if (e.key === 'Escape') {
      setShowPalette(false)
    }
  }, [])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  return (
    <div style={{ display: 'flex', minHeight: '100vh', background: 'var(--color-bg-base)' }}>
      {/* Sidebar */}
      <nav
        onMouseEnter={() => setExpanded(true)}
        onMouseLeave={() => setExpanded(false)}
        style={{
          width: expanded ? 240 : 64,
          transition: 'width 0.2s ease',
          background: '#0A0A0A',
          display: 'flex',
          flexDirection: 'column',
          position: 'fixed',
          top: 0,
          left: 0,
          bottom: 0,
          zIndex: 100,
          overflow: 'hidden',
          borderRight: '1px solid #1A1A1A',
        }}
      >
        {/* Brand */}
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          padding: expanded ? '0 20px' : '0 0',
          justifyContent: expanded ? 'flex-start' : 'center',
          borderBottom: '1px solid #1A1A1A',
          flexShrink: 0,
        }}>
          <span style={{
            fontFamily: 'var(--font-mono)',
            fontSize: expanded ? 18 : 14,
            fontWeight: 700,
            color: '#FFFFFF',
            letterSpacing: '-0.04em',
            whiteSpace: 'nowrap',
          }}>
            {expanded ? 'TICKET' : 'T'}
          </span>
        </div>

        {/* Navigation */}
        <div style={{ flex: 1, padding: '8px 0', overflowY: 'auto' }}>
          {navItems.map(item => {
            if (item.key === 'divider') {
              return <div key="divider" style={{ height: 1, background: '#1A1A1A', margin: '8px 16px' }} />
            }
            const isActive = location.pathname === item.key
            return (
              <Tooltip key={item.key} title={expanded ? '' : item.label} placement="right">
                <button
                  onClick={() => navigate(item.key)}
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    padding: expanded ? '0 20px' : '0 0',
                    justifyContent: expanded ? 'flex-start' : 'center',
                    height: 40,
                    border: 'none',
                    background: isActive ? 'rgba(255,255,255,0.1)' : 'transparent',
                    color: isActive ? '#FFFFFF' : '#666666',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    fontSize: 14,
                    fontFamily: 'var(--font-body)',
                    borderRadius: expanded ? 4 : 0,
                    margin: expanded ? '2px 8px' : 0,
                  }}
                  onMouseEnter={e => {
                    if (!isActive) e.currentTarget.style.color = '#999999'
                  }}
                  onMouseLeave={e => {
                    if (!isActive) e.currentTarget.style.color = '#666666'
                  }}
                >
                  <span style={{ fontSize: 18, lineHeight: 1 }}>{item.icon}</span>
                  {expanded && <span style={{ whiteSpace: 'nowrap' }}>{item.label}</span>}
                </button>
              </Tooltip>
            )
          })}
        </div>

        {/* Bottom actions */}
        <div style={{
          padding: expanded ? '12px 16px' : '12px 0',
          borderTop: '1px solid #1A1A1A',
          display: 'flex',
          flexDirection: 'column',
          gap: 4,
          alignItems: expanded ? 'stretch' : 'center',
          flexShrink: 0,
        }}>
          {/* Search trigger */}
          <Tooltip title={expanded ? '' : '搜索'} placement="right">
            <button
              onClick={() => setShowPalette(true)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: expanded ? '8px 12px' : '8px',
                justifyContent: expanded ? 'flex-start' : 'center',
                border: '1px solid #262626',
                background: 'transparent',
                color: '#666666',
                cursor: 'pointer',
                fontSize: 13,
                fontFamily: 'var(--font-body)',
                borderRadius: 4,
                marginBottom: 8,
              }}
            >
              <SearchOutlined style={{ fontSize: 14 }} />
              {expanded && (
                <>
                  <span style={{ flex: 1, textAlign: 'left' }}>搜索</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: '#444' }}>⌘K</span>
                </>
              )}
            </button>
          </Tooltip>

          {/* Theme toggle */}
          <Tooltip title={expanded ? '' : (isDark ? '浅色模式' : '深色模式')} placement="right">
            <button
              onClick={onThemeToggle}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: expanded ? '8px 12px' : '8px',
                justifyContent: expanded ? 'flex-start' : 'center',
                border: 'none',
                background: 'transparent',
                color: '#666666',
                cursor: 'pointer',
                fontSize: 14,
                borderRadius: 4,
              }}
            >
              {isDark ? <SunOutlined /> : <MoonOutlined />}
              {expanded && <span>{isDark ? '浅色' : '深色'}</span>}
            </button>
          </Tooltip>

          {/* Notifications */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: expanded ? 'flex-start' : 'center',
            padding: expanded ? '0 12px' : 0,
          }}>
            <NotificationBell />
          </div>

          {/* User */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '8px 0',
            justifyContent: expanded ? 'flex-start' : 'center',
            marginTop: 4,
          }}>
            <Avatar size={28} style={{ background: '#0066FF', fontSize: 12, flexShrink: 0 }}>
              {user?.username?.[0]?.toUpperCase()}
            </Avatar>
            {expanded && (
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ color: '#F5F5F5', fontSize: 13, fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {user?.username}
                </div>
                <div style={{ color: '#666', fontSize: 11 }}>{isAdmin ? '管理员' : '用户'}</div>
              </div>
            )}
            {expanded && (
              <button
                onClick={handleLogout}
                style={{
                  border: 'none',
                  background: 'transparent',
                  color: '#666',
                  cursor: 'pointer',
                  padding: 4,
                  fontSize: 14,
                }}
              >
                <LogoutOutlined />
              </button>
            )}
          </div>
        </div>
      </nav>

      {/* Main content */}
      <main style={{
        flex: 1,
        marginLeft: 64,
        minHeight: '100vh',
        transition: 'margin-left 0.2s ease',
      }}>
        <div style={{
          maxWidth: 1200,
          margin: '0 auto',
          padding: '32px 32px',
        }}>
          <Outlet />
        </div>
      </main>

      {/* Command palette */}
      {showPalette && <CommandPalette onClose={() => setShowPalette(false)} />}
    </div>
  )
}
