import { Card, Row, Col, Button, List } from 'antd'
import { CalendarOutlined, TagOutlined, UserOutlined, RocketOutlined, ShopOutlined, SwapOutlined, ClockCircleOutlined, ThunderboltOutlined, SafetyOutlined, BellOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { getCardIconGradient } from '../../theme/gradients'

export default function Dashboard() {
  const navigate = useNavigate()
  const { user } = useAuthStore()

  const hour = new Date().getHours()
  const greeting = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'

  const features = [
    { icon: <CalendarOutlined />, label: '活动中心', desc: '浏览活动并抢票', gradient: getCardIconGradient(0), path: '/events' },
    { icon: <TagOutlined />, label: '我的票务', desc: '查看已购票', gradient: getCardIconGradient(1), path: '/tickets' },
    { icon: <ShopOutlined />, label: '二手市场', desc: '买卖二手票', gradient: getCardIconGradient(2), path: '/marketplace' },
    { icon: <SwapOutlined />, label: '转让记录', desc: '查看转让历史', gradient: getCardIconGradient(3), path: '/transfer-records' },
    { icon: <UserOutlined />, label: '个人中心', desc: '查看信息', gradient: getCardIconGradient(4), path: '/profile' },
  ]

  const systemFeatures = [
    { icon: <CalendarOutlined style={{ color: '#5B2FE8' }} />, text: '活动管理 - 创建和管理票务活动' },
    { icon: <ThunderboltOutlined style={{ color: '#D4A843' }} />, text: '在线抢票 - Redis 原子扣库存，公平抢票' },
    { icon: <BellOutlined style={{ color: '#6366F1' }} />, text: '实时通知 - WebSocket 推送抢票结果' },
    { icon: <TagOutlined style={{ color: '#22C55E' }} />, text: '票务管理 - 支付、退票、使用' },
    { icon: <ShopOutlined style={{ color: '#F59E0B' }} />, text: '二手市场 - 官方认证二手票交易' },
    { icon: <SafetyOutlined style={{ color: '#EC4899' }} />, text: '票务转让 - 转赠好友或出售' },
  ]

  return (
    <div className="animate-slide-up">
      {/* Welcome Banner */}
      <div style={{
        background: 'linear-gradient(135deg, #5B2FE8 0%, #6366F1 50%, #D4A843 100%)',
        borderRadius: 16,
        padding: '32px 40px',
        marginBottom: 24,
        position: 'relative',
        overflow: 'hidden',
        color: '#fff',
      }}>
        <div style={{
          position: 'absolute',
          inset: 0,
          background: 'repeating-linear-gradient(45deg, transparent, transparent 20px, rgba(255,255,255,0.03) 20px, rgba(255,255,255,0.03) 40px)',
        }} />
        <div style={{ position: 'relative', zIndex: 1 }}>
          <h1 style={{ margin: 0, fontSize: 28, fontWeight: 700 }}>
            {greeting}，<span style={{ color: '#D4A843' }}>{user?.username || '用户'}</span>
          </h1>
          <p style={{ margin: '8px 0 0', opacity: 0.8, fontSize: 15 }}>
            <ClockCircleOutlined style={{ marginRight: 6 }} />
            {new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' })}
          </p>
        </div>
      </div>

      {/* Feature Cards */}
      <Row gutter={[16, 16]}>
        {features.map((f) => (
          <Col key={f.path} xs={12} sm={8} lg={4} xl={4}>
            <Card
              hoverable
              onClick={() => navigate(f.path)}
              style={{
                borderRadius: 16,
                borderLeft: `4px solid ${f.gradient.includes('#5B2FE8') ? '#5B2FE8' : f.gradient.includes('#D4A843') ? '#D4A843' : f.gradient.includes('#6366F1') ? '#6366F1' : f.gradient.includes('#14B8A6') ? '#14B8A6' : '#EC4899'}`,
                cursor: 'pointer',
                transition: 'all 0.2s ease',
              }}
              bodyStyle={{ padding: 20 }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <div className="icon-circle" style={{ background: f.gradient, width: 44, height: 44, fontSize: 20 }}>
                  {f.icon}
                </div>
                <div>
                  <div style={{ fontWeight: 600, fontSize: 15, marginBottom: 2 }}>{f.label}</div>
                  <div style={{ color: 'var(--color-text-tertiary)', fontSize: 12 }}>{f.desc}</div>
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        <Col xs={24} lg={12}>
          <Card title="快捷操作" style={{ borderRadius: 16 }} bodyStyle={{ padding: '12px 24px' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              <Button
                type="primary"
                block
                icon={<RocketOutlined />}
                onClick={() => navigate('/events')}
                style={{ height: 44, borderRadius: 10, textAlign: 'left', fontWeight: 500 }}
              >
                立即抢票
              </Button>
              <Button
                block
                icon={<ShopOutlined />}
                onClick={() => navigate('/marketplace')}
                style={{ height: 44, borderRadius: 10, textAlign: 'left' }}
              >
                二手市场
              </Button>
              <Button
                block
                icon={<SwapOutlined />}
                onClick={() => navigate('/transfer-records')}
                style={{ height: 44, borderRadius: 10, textAlign: 'left' }}
              >
                转让记录
              </Button>
            </div>
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="系统功能" style={{ borderRadius: 16 }} bodyStyle={{ padding: '12px 24px' }}>
            <List
              dataSource={systemFeatures}
              renderItem={(item) => (
                <List.Item style={{ padding: '8px 0', border: 'none' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 14 }}>
                    <div style={{
                      width: 32,
                      height: 32,
                      borderRadius: 8,
                      background: 'var(--color-primary-bg)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: 16,
                      flexShrink: 0,
                    }}>
                      {item.icon}
                    </div>
                    {item.text}
                  </div>
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
