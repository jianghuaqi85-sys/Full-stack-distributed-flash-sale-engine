import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  DashboardOutlined,
  CalendarOutlined,
  ShopOutlined,
  TagOutlined,
  SwapOutlined,
  UserOutlined,
  BarChartOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../stores/authStore'

interface Props {
  onClose: () => void
}

interface CommandItem {
  key: string
  label: string
  description: string
  icon: React.ReactNode
  action: () => void
  shortcut?: string
}

export default function CommandPalette({ onClose }: Props) {
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const navigate = useNavigate()
  const { user } = useAuthStore()

  const isAdmin = user?.role === 'admin'

  const allItems: CommandItem[] = [
    { key: '/', label: '仪表盘', description: '查看概览', icon: <DashboardOutlined />, action: () => navigate('/') },
    { key: '/events', label: '活动中心', description: '浏览活动', icon: <CalendarOutlined />, action: () => navigate('/events') },
    { key: '/marketplace', label: '二手市场', description: '票务交易', icon: <ShopOutlined />, action: () => navigate('/marketplace') },
    { key: '/tickets', label: '我的票务', description: '查看票务', icon: <TagOutlined />, action: () => navigate('/tickets') },
    { key: '/transfer-records', label: '转让记录', description: '查看转让', icon: <SwapOutlined />, action: () => navigate('/transfer-records') },
    { key: '/profile', label: '个人中心', description: '账户设置', icon: <UserOutlined />, action: () => navigate('/profile') },
    ...(isAdmin ? [
      { key: '/admin/dashboard', label: '数据仪表盘', description: '统计数据', icon: <BarChartOutlined />, action: () => navigate('/admin/dashboard') },
      { key: '/admin/events', label: '活动管理', description: '管理活动', icon: <SettingOutlined />, action: () => navigate('/admin/events') },
    ] : []),
  ]

  const filtered = allItems.filter(item =>
    item.label.includes(query) || item.description.includes(query)
  )

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex(prev => Math.min(prev + 1, filtered.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex(prev => Math.max(prev - 1, 0))
    } else if (e.key === 'Enter' && filtered[selectedIndex]) {
      filtered[selectedIndex].action()
      onClose()
    }
  }

  return (
    <div className="command-palette-backdrop" onClick={onClose}>
      <div className="command-palette" onClick={e => e.stopPropagation()}>
        <input
          ref={inputRef}
          value={query}
          onChange={e => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="搜索页面、活动..."
          style={{
            width: '100%',
            padding: '16px 20px',
            fontSize: 16,
            border: 'none',
            borderBottom: '1px solid var(--color-border)',
            background: 'transparent',
            color: 'var(--color-text-primary)',
            outline: 'none',
            fontFamily: 'var(--font-body)',
          }}
        />
        <div style={{ maxHeight: 320, overflowY: 'auto' }}>
          {filtered.length === 0 ? (
            <div style={{ padding: '20px', textAlign: 'center', color: 'var(--color-text-tertiary)', fontSize: 14 }}>
              无结果
            </div>
          ) : (
            filtered.map((item, index) => (
              <div
                key={item.key}
                className={`command-palette-item ${index === selectedIndex ? 'active' : ''}`}
                onClick={() => {
                  item.action()
                  onClose()
                }}
                onMouseEnter={() => setSelectedIndex(index)}
              >
                <span style={{ fontSize: 16, color: 'var(--color-text-tertiary)', width: 20, textAlign: 'center' }}>
                  {item.icon}
                </span>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 500 }}>{item.label}</div>
                  <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}>{item.description}</div>
                </div>
              </div>
            ))
          )}
        </div>
        <div style={{
          padding: '8px 20px',
          borderTop: '1px solid var(--color-border)',
          display: 'flex',
          gap: 16,
          fontSize: 11,
          color: 'var(--color-text-tertiary)',
          fontFamily: 'var(--font-mono)',
        }}>
          <span>↑↓ 导航</span>
          <span>↵ 选择</span>
          <span>esc 关闭</span>
        </div>
      </div>
    </div>
  )
}
