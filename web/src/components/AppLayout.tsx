import { useState } from 'react'
import { Layout, Menu, Button, theme, Space, Tag, Avatar } from 'antd'
import { DashboardOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined, CalendarOutlined, TagOutlined, UserOutlined, SettingOutlined, BarChartOutlined, ShopOutlined, SwapOutlined, AuditOutlined, PercentageOutlined } from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import NotificationBell from './NotificationBell'
import ThemeToggle from './ThemeToggle'
import BrandLogo from './BrandLogo'
import PageTransition from './PageTransition'

const { Header, Sider, Content } = Layout

interface Props {
  isDark: boolean
  onThemeToggle: () => void
}

export default function AppLayout({ isDark, onThemeToggle }: Props) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken()

  const isAdmin = user?.role === 'admin'

  const menuItems = [
    { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
    { key: '/events', icon: <CalendarOutlined />, label: '活动中心' },
    { key: '/marketplace', icon: <ShopOutlined />, label: '二手市场' },
    { key: '/tickets', icon: <TagOutlined />, label: '我的票务' },
    { key: '/transfer-records', icon: <SwapOutlined />, label: '转让记录' },
    { key: '/profile', icon: <UserOutlined />, label: '个人中心' },
    ...(isAdmin ? [
      { key: '/admin/dashboard', icon: <BarChartOutlined />, label: '数据仪表盘' },
      { key: '/admin/events', icon: <SettingOutlined />, label: '活动管理' },
      { key: '/admin/transfers', icon: <AuditOutlined />, label: '转让审核' },
      { key: '/admin/promo', icon: <PercentageOutlined />, label: '促销码管理' },
    ] : []),
  ]

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          background: 'linear-gradient(180deg, #151025 0%, #0D0A1A 100%)',
          borderRight: '1px solid var(--color-border)',
        }}
      >
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 20px',
          gap: 10,
          borderBottom: '1px solid rgba(139, 111, 255, 0.15)',
        }}>
          <BrandLogo size={32} collapsed={collapsed} />
          {!collapsed && (
            <span style={{ color: '#F0EDFC', fontWeight: 700, fontSize: 16, whiteSpace: 'nowrap' }}>
              票务系统
            </span>
          )}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ background: 'transparent', borderRight: 0 }}
        />
      </Sider>
      <Layout>
        <Header style={{
          padding: '0 24px',
          background: colorBgContainer,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: '1px solid var(--color-border-light)',
          position: 'relative',
        }}>
          <div style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: 2,
            background: 'linear-gradient(90deg, var(--color-primary), transparent)',
            opacity: 0.3,
          }} />
          <Button type="text" icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => setCollapsed(!collapsed)} />
          <Space>
            <NotificationBell />
            <ThemeToggle isDark={isDark} onToggle={onThemeToggle} />
            <Avatar size={28} style={{ background: 'linear-gradient(135deg, #5B2FE8, #D4A843)' }}>
              {user?.username?.[0]?.toUpperCase()}
            </Avatar>
            <span style={{ fontWeight: 500 }}>{user?.username}</span>
            {isAdmin && <Tag color="gold" style={{ borderRadius: 8 }}>管理员</Tag>}
            <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>退出</Button>
          </Space>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: colorBgContainer, borderRadius: borderRadiusLG, minHeight: 280 }}>
          <PageTransition>
            <Outlet />
          </PageTransition>
        </Content>
      </Layout>
    </Layout>
  )
}
