import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Button, Tag, Spin, List, message, InputNumber, Space, Empty } from 'antd'
import { CalendarOutlined, EnvironmentOutlined, ArrowLeftOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { getEvent, getEventStock, Event, TicketType, Show } from '../../api/events'
import { getEventShows } from '../../api/shows'
import { purchaseTicket } from '../../api/tickets'
import { getEventGradient } from '../../theme/gradients'

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: 'default', text: '草稿' },
  on_sale: { color: 'success', text: '售票中' },
  off_sale: { color: 'warning', text: '已下架' },
  ended: { color: 'error', text: '已结束' },
}

export default function EventDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [event, setEvent] = useState<Event | null>(null)
  const [loading, setLoading] = useState(false)
  const [purchasing, setPurchasing] = useState<number | null>(null)
  const [quantities, setQuantities] = useState<Record<number, number>>({})
  const [realTimeStock, setRealTimeStock] = useState<Record<string, number>>({})
  const [shows, setShows] = useState<Show[]>([])
  const [selectedShow, setSelectedShow] = useState<number | null>(null)

  const fetchEvent = useCallback(() => {
    if (!id) return
    setLoading(true)
    getEvent(Number(id))
      .then((res) => setEvent(res.data))
      .finally(() => setLoading(false))
  }, [id])

  const fetchStock = useCallback(() => {
    if (!id) return
    getEventStock(Number(id))
      .then((res) => setRealTimeStock(res.data))
      .catch(() => {})
  }, [id])

  const fetchShows = useCallback(() => {
    if (!id) return
    getEventShows(Number(id))
      .then((res) => setShows(res.data.data || []))
      .catch(() => {})
  }, [id])

  useEffect(() => {
    fetchEvent()
    fetchStock()
    fetchShows()

    const interval = setInterval(() => {
      fetchStock()
      fetchShows()
    }, 5000)
    return () => clearInterval(interval)
  }, [fetchEvent, fetchStock, fetchShows])

  useEffect(() => {
    if (shows.length > 0 && selectedShow === null) {
      const activeShow = shows.find((s) => s.status === 'on_sale')
      if (activeShow) setSelectedShow(activeShow.id)
    }
  }, [shows, selectedShow])

  const handlePurchase = async (ticketType: TicketType) => {
    const qty = quantities[ticketType.id] || 1
    setPurchasing(ticketType.id)
    try {
      await purchaseTicket({
        event_id: Number(id),
        show_id: selectedShow || undefined,
        ticket_type_id: ticketType.id,
        quantity: qty,
      })
      message.success('排队中，请稍候...')
      setTimeout(() => navigate('/tickets'), 2000)
    } catch {
      // error handled by interceptor
    } finally {
      setPurchasing(null)
    }
  }

  const getStock = (ticketTypeId: number) => realTimeStock[String(ticketTypeId)] ?? 0

  const activeShows = shows.filter((s) => s.status === 'on_sale')
  const hasShows = shows.length > 0
  const currentShow = shows.find((s) => s.id === selectedShow)

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!event) return <div>活动不存在</div>

  return (
    <div className="animate-slide-up">
      <Button icon={<ArrowLeftOutlined />} style={{ marginBottom: 16, borderRadius: 8 }} onClick={() => navigate('/events')}>
        返回列表
      </Button>

      {/* Hero Banner */}
      <div className="hero-banner" style={{
        background: getEventGradient(event.id),
        marginBottom: 20,
        minHeight: 200,
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'flex-end',
      }}>
        <div style={{ position: 'relative', zIndex: 1 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 12 }}>
            <div>
              <h1 style={{ margin: 0, fontSize: 28, fontWeight: 700, textShadow: '0 2px 8px rgba(0,0,0,0.3)' }}>
                {event.title}
              </h1>
              <div style={{ display: 'flex', gap: 16, marginTop: 12, opacity: 0.9, fontSize: 14 }}>
                <span><CalendarOutlined style={{ marginRight: 6 }} />{new Date(event.start_time).toLocaleString('zh-CN')}</span>
                <span><EnvironmentOutlined style={{ marginRight: 6 }} />{event.location}</span>
              </div>
            </div>
            <Tag color={statusMap[event.status]?.color} style={{ fontSize: 14, padding: '4px 14px', borderRadius: 8 }}>
              {statusMap[event.status]?.text}
            </Tag>
          </div>
          {event.description && (
            <p style={{ margin: '12px 0 0', opacity: 0.8, fontSize: 14 }}>{event.description}</p>
          )}
        </div>
      </div>

      {/* Show Selection */}
      {hasShows && (
        <Card title={<><ClockCircleOutlined /> 选择场次</>} style={{ borderRadius: 16, marginBottom: 16 }}>
          {activeShows.length > 0 ? (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 12 }}>
              {shows.map((show) => (
                <div
                  key={show.id}
                  onClick={() => show.status === 'on_sale' && setSelectedShow(show.id)}
                  style={{
                    padding: '14px 16px',
                    borderRadius: 12,
                    border: `2px solid ${selectedShow === show.id ? 'var(--color-primary)' : 'var(--color-border)'}`,
                    background: selectedShow === show.id ? 'var(--color-primary-bg)' : 'var(--color-bg-container)',
                    cursor: show.status === 'on_sale' ? 'pointer' : 'not-allowed',
                    opacity: show.status !== 'on_sale' ? 0.5 : 1,
                    transition: 'all 0.2s ease',
                  }}
                >
                  <div style={{ fontWeight: 600, marginBottom: 4 }}>{show.name}</div>
                  <div style={{ fontSize: 12, color: 'var(--color-text-secondary)' }}>
                    {new Date(show.show_time).toLocaleString('zh-CN')}
                  </div>
                  <div style={{ fontSize: 12, marginTop: 4 }}>
                    {show.status === 'on_sale'
                      ? <span style={{ color: 'var(--color-success)' }}>剩余 {show.stock} 张</span>
                      : <span style={{ color: 'var(--color-error)' }}>{statusMap[show.status]?.text || show.status}</span>
                    }
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <Empty description="暂无可选场次" />
          )}
        </Card>
      )}

      {/* Ticket Types */}
      <Card
        title={currentShow ? `${currentShow.name} - 选择票种` : '选择票种'}
        style={{ borderRadius: 16 }}
      >
        <List
          dataSource={event.ticket_types || []}
          renderItem={(tt) => {
            const stock = getStock(tt.id)
            return (
              <List.Item
                style={{ padding: '16px 0' }}
                actions={
                  event.status === 'on_sale' && stock > 0
                    ? [
                        <Space key="buy">
                          <InputNumber
                            min={1}
                            max={tt.max_per_user}
                            value={quantities[tt.id] || 1}
                            onChange={(v) => setQuantities({ ...quantities, [tt.id]: v || 1 })}
                            style={{ width: 70, borderRadius: 8 }}
                          />
                          <Button
                            type="primary"
                            loading={purchasing === tt.id}
                            onClick={() => handlePurchase(tt)}
                            style={{ borderRadius: 8, fontWeight: 600 }}
                          >
                            抢购
                          </Button>
                        </Space>,
                      ]
                    : []
                }
              >
                <List.Item.Meta
                  title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <span style={{
                        width: 4,
                        height: 24,
                        borderRadius: 2,
                        background: 'linear-gradient(180deg, var(--color-primary), var(--color-gold))',
                      }} />
                      <span style={{ fontSize: 16, fontWeight: 600 }}>{tt.name}</span>
                      <span style={{ color: 'var(--color-gold)', fontSize: 22, fontWeight: 700 }}>
                        ¥{tt.price}
                      </span>
                    </div>
                  }
                  description={
                    <div style={{ display: 'flex', gap: 20, marginLeft: 16 }}>
                      <span style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 13 }}>
                        <span className={`status-dot ${stock > 0 ? 'active' : 'error'}`} />
                        <span style={{ color: stock > 0 ? 'var(--color-success)' : 'var(--color-error)', fontWeight: 500 }}>
                          剩余 {stock} 张
                        </span>
                      </span>
                      <span style={{ color: 'var(--color-text-tertiary)', fontSize: 13 }}>每人限购 {tt.max_per_user} 张</span>
                    </div>
                  }
                />
              </List.Item>
            )
          }}
        />
      </Card>
    </div>
  )
}
