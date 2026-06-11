import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

export default function Dashboard() {
  const navigate = useNavigate()
  const { user } = useAuthStore()

  const now = new Date()
  const hour = now.getHours()
  const greeting = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'
  const dateStr = now.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  })

  const navItems = [
    { path: '/events', label: '活动中心', desc: '浏览活动并抢票' },
    { path: '/tickets', label: '我的票务', desc: '查看已购票务' },
    { path: '/marketplace', label: '二手市场', desc: '买卖二手票' },
    { path: '/transfer-records', label: '转让记录', desc: '查看转让历史' },
    { path: '/profile', label: '个人中心', desc: '账户设置' },
  ]

  return (
    <div className="page-enter">
      {/* Header */}
      <div className="page-header">
        <h1>{greeting}，{user?.username}</h1>
        <div className="subtitle">{dateStr}</div>
      </div>

      {/* Navigation list */}
      <div style={{ marginTop: 48 }}>
        <div style={{
          fontSize: 13,
          color: 'var(--color-text-tertiary)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontFamily: 'var(--font-mono)',
          marginBottom: 16,
        }}>
          快速导航
        </div>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          {navItems.map(item => (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '20px 0',
                border: 'none',
                borderBottom: '1px solid var(--color-border)',
                background: 'transparent',
                cursor: 'pointer',
                textAlign: 'left',
                transition: 'padding-left 0.15s',
              }}
              onMouseEnter={e => {
                e.currentTarget.style.paddingLeft = '12px'
              }}
              onMouseLeave={e => {
                e.currentTarget.style.paddingLeft = '0'
              }}
            >
              <div>
                <div style={{
                  fontSize: 20,
                  fontWeight: 600,
                  color: 'var(--color-text-primary)',
                  letterSpacing: '-0.02em',
                }}>
                  {item.label}
                </div>
                <div style={{
                  fontSize: 13,
                  color: 'var(--color-text-tertiary)',
                  marginTop: 4,
                }}>
                  {item.desc}
                </div>
              </div>
              <span style={{
                fontSize: 20,
                color: 'var(--color-text-tertiary)',
                transition: 'color 0.15s',
              }}>
                →
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* System info */}
      <div style={{ marginTop: 48 }}>
        <div style={{
          fontSize: 13,
          color: 'var(--color-text-tertiary)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontFamily: 'var(--font-mono)',
          marginBottom: 16,
        }}>
          系统能力
        </div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          gap: 1,
          background: 'var(--color-border)',
          border: '1px solid var(--color-border)',
        }}>
          {[
            { title: '秒杀抢票', desc: 'Redis 原子扣库存，公平排队' },
            { title: '实时通知', desc: 'WebSocket 推送抢票结果' },
            { title: '票务管理', desc: '支付、退票、使用全生命周期' },
            { title: '二手市场', desc: '官方认证二手票交易' },
            { title: '票务转让', desc: '转赠好友或审核转让' },
            { title: '促销码', desc: '折扣码、满减码支持' },
          ].map(item => (
            <div
              key={item.title}
              style={{
                padding: '24px',
                background: 'var(--color-bg-container)',
              }}
            >
              <div style={{
                fontSize: 14,
                fontWeight: 600,
                color: 'var(--color-text-primary)',
                marginBottom: 4,
              }}>
                {item.title}
              </div>
              <div style={{
                fontSize: 13,
                color: 'var(--color-text-tertiary)',
              }}>
                {item.desc}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
