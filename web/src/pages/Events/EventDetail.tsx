import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Tag, Spin, message, InputNumber, Space } from 'antd'
import { getEvent, getEventStock, Event, TicketType, Show } from '../../api/events'
import { getEventShows } from '../../api/shows'
import { purchaseTicket } from '../../api/tickets'
import { joinQueue, leaveQueue } from '../../api/queue'
import { getActiveListings, MarketplaceListing } from '../../api/marketplace'
import QueueWaiting from '../../components/QueueWaiting'
import WaitlistButton from '../../components/WaitlistButton'
import PromoCodeInput from '../../components/PromoCodeInput'
import Countdown from '../../components/Countdown'

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
  const [queueState, setQueueState] = useState<'none' | 'queuing' | 'ready' | 'error'>('none')
  const [pendingPurchase, setPendingPurchase] = useState<{ ticketType: TicketType; quantity: number } | null>(null)
  const [promoDiscount, setPromoDiscount] = useState<number | null>(null)
  const [promoCode, setPromoCode] = useState<string | null>(null)
  const [marketplaceListings, setMarketplaceListings] = useState<MarketplaceListing[]>([])

  const fetchEvent = useCallback(() => {
    if (!id) return
    setLoading(true)
    getEvent(Number(id)).then((res) => setEvent(res.data)).finally(() => setLoading(false))
  }, [id])

  const fetchStock = useCallback(() => {
    if (!id) return
    getEventStock(Number(id)).then((res) => setRealTimeStock(res.data)).catch(() => {})
  }, [id])

  const fetchShows = useCallback(() => {
    if (!id) return
    getEventShows(Number(id)).then((res) => setShows(res.data.data || [])).catch(() => {})
  }, [id])

  const fetchMarketplace = useCallback(() => {
    if (!id) return
    getActiveListings(1, 5).then((res) => {
      setMarketplaceListings((res.data.data || []).filter((l: MarketplaceListing) => l.event_id === Number(id)))
    }).catch(() => {})
  }, [id])

  useEffect(() => {
    fetchEvent()
    fetchStock()
    fetchShows()
    fetchMarketplace()
    const interval = setInterval(() => { fetchStock(); fetchShows() }, 5000)
    return () => clearInterval(interval)
  }, [fetchEvent, fetchStock, fetchShows, fetchMarketplace])

  useEffect(() => {
    if (shows.length > 0 && selectedShow === null) {
      const activeShow = shows.find((s) => s.status === 'on_sale')
      if (activeShow) setSelectedShow(activeShow.id)
    }
  }, [shows, selectedShow])

  const executePurchase = async (ticketType: TicketType, quantity: number) => {
    setPurchasing(ticketType.id)
    try {
      await purchaseTicket({ event_id: Number(id), show_id: selectedShow || undefined, ticket_type_id: ticketType.id, quantity })
      message.success('排队中，请稍候...')
      setTimeout(() => navigate('/tickets'), 2000)
    } catch { /* handled */ } finally { setPurchasing(null) }
  }

  const handlePurchase = async (ticketType: TicketType) => {
    const qty = quantities[ticketType.id] || 1
    setPendingPurchase({ ticketType, quantity: qty })
    setQueueState('queuing')
    try {
      await joinQueue(Number(id))
    } catch {
      setQueueState('none')
      setPendingPurchase(null)
      executePurchase(ticketType, qty)
    }
  }

  const handleQueueReady = () => {
    setQueueState('ready')
    if (pendingPurchase) {
      executePurchase(pendingPurchase.ticketType, pendingPurchase.quantity)
      setQueueState('none')
      setPendingPurchase(null)
    }
  }

  const handleQueueLeave = () => {
    leaveQueue(Number(id)).catch(() => {})
    setQueueState('none')
    setPendingPurchase(null)
  }

  const getStock = (ticketTypeId: number) => realTimeStock[String(ticketTypeId)] ?? 0
  const totalStock = event?.ticket_types?.reduce((sum, tt) => sum + (getStock(tt.id) || 0), 0) ?? 0
  const isSoldOut = totalStock === 0 && event?.status === 'on_sale'
  const currentShow = shows.find((s) => s.id === selectedShow)

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!event) return <div>活动不存在</div>

  return (
    <div className="page-enter">
      {/* Back */}
      <button
        onClick={() => navigate('/events')}
        style={{
          border: 'none',
          background: 'transparent',
          color: 'var(--color-text-tertiary)',
          cursor: 'pointer',
          fontSize: 13,
          fontFamily: 'var(--font-mono)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          padding: 0,
          marginBottom: 24,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        ← 返回列表
      </button>

      {/* Header */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <Tag color={statusMap[event.status]?.color}>{statusMap[event.status]?.text}</Tag>
        </div>
        <h1 style={{
          fontSize: 48,
          fontWeight: 700,
          letterSpacing: '-0.04em',
          lineHeight: 1.1,
          color: 'var(--color-text-primary)',
          margin: 0,
        }}>
          {event.title}
        </h1>
        {event.description && (
          <p style={{ fontSize: 15, color: 'var(--color-text-secondary)', marginTop: 12, maxWidth: 600, lineHeight: 1.6 }}>
            {event.description}
          </p>
        )}
      </div>

      {/* Info grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
        gap: 1,
        background: 'var(--color-border)',
        border: '1px solid var(--color-border)',
        marginBottom: 32,
      }}>
        <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
          <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>开始时间</div>
          <div style={{ fontSize: 15, fontWeight: 500, fontFamily: 'var(--font-mono)' }}>{new Date(event.start_time).toLocaleString('zh-CN')}</div>
        </div>
        <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
          <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>地点</div>
          <div style={{ fontSize: 15, fontWeight: 500 }}>{event.location}</div>
        </div>
        <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
          <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>总库存</div>
          <div style={{ fontSize: 15, fontWeight: 500, fontFamily: 'var(--font-mono)' }}>{totalStock}</div>
        </div>
      </div>

      {/* Countdown for draft events */}
      {event.status === 'draft' && new Date(event.start_time) > new Date() && (
        <div style={{ padding: '24px', border: '1px solid var(--color-border)', marginBottom: 32, textAlign: 'center' }}>
          <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 8 }}>距开售还有</div>
          <Countdown targetDate={event.start_time} size="large" format="DD天 HH:mm:ss" onFinish={fetchEvent} />
        </div>
      )}

      {/* Queue */}
      {queueState === 'queuing' && (
        <div style={{ marginBottom: 32 }}>
          <QueueWaiting eventId={Number(id)} onReady={handleQueueReady} onLeave={handleQueueLeave} />
        </div>
      )}

      {/* Show selection */}
      {shows.length > 0 && (
        <div style={{ marginBottom: 32 }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 12 }}>选择场次</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 1, background: 'var(--color-border)' }}>
            {shows.map((show) => (
              <button
                key={show.id}
                onClick={() => show.status === 'on_sale' && setSelectedShow(show.id)}
                style={{
                  padding: '16px 20px',
                  border: 'none',
                  background: selectedShow === show.id ? 'var(--color-accent-soft)' : 'var(--color-bg-container)',
                  cursor: show.status === 'on_sale' ? 'pointer' : 'not-allowed',
                  opacity: show.status !== 'on_sale' ? 0.5 : 1,
                  textAlign: 'left',
                }}
              >
                <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 4 }}>{show.name}</div>
                <div style={{ fontSize: 12, color: 'var(--color-text-secondary)', fontFamily: 'var(--font-mono)' }}>
                  {new Date(show.show_time).toLocaleString('zh-CN')}
                </div>
                <div style={{ fontSize: 12, marginTop: 4, color: show.status === 'on_sale' ? 'var(--color-success)' : 'var(--color-error)' }}>
                  {show.status === 'on_sale' ? `剩余 ${show.stock} 张` : (statusMap[show.status]?.text || show.status)}
                </div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Ticket types */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 12 }}>
          {currentShow ? `${currentShow.name} — 选择票种` : '选择票种'}
        </div>
        <div style={{ border: '1px solid var(--color-border)' }}>
          {(event.ticket_types || []).map((tt) => {
            const stock = getStock(tt.id)
            return (
              <div
                key={tt.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '16px 20px',
                  borderBottom: '1px solid var(--color-border)',
                }}
              >
                <div>
                  <div style={{ fontWeight: 600, fontSize: 15 }}>{tt.name}</div>
                  <div style={{ display: 'flex', gap: 16, marginTop: 4 }}>
                    <span style={{ fontSize: 13, fontFamily: 'var(--font-mono)', color: stock > 0 ? 'var(--color-success)' : 'var(--color-error)' }}>
                      剩余 {stock} 张
                    </span>
                    <span style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}>
                      每人限购 {tt.max_per_user} 张
                    </span>
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                  <span style={{ fontSize: 24, fontWeight: 700, fontFamily: 'var(--font-mono)', letterSpacing: '-0.02em' }}>
                    ¥{tt.price}
                  </span>
                  {event.status === 'on_sale' && stock > 0 && queueState !== 'queuing' && (
                    <Space>
                      <InputNumber
                        min={1}
                        max={tt.max_per_user}
                        value={quantities[tt.id] || 1}
                        onChange={(v) => setQuantities({ ...quantities, [tt.id]: v || 1 })}
                        style={{ width: 64 }}
                      />
                      <Button
                        type="primary"
                        loading={purchasing === tt.id}
                        onClick={() => handlePurchase(tt)}
                      >
                        抢购
                      </Button>
                    </Space>
                  )}
                </div>
              </div>
            )
          })}
        </div>

        {/* Promo code */}
        {event.status === 'on_sale' && !isSoldOut && (
          <div style={{ marginTop: 16, maxWidth: 400 }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 8 }}>优惠码</div>
            <PromoCodeInput amount={0} onApply={(discount, _final, code) => { setPromoDiscount(discount); setPromoCode(code) }} />
            {promoDiscount && promoDiscount > 0 && (
              <div style={{ marginTop: 8, fontSize: 13, color: 'var(--color-success)' }}>
                已优惠 ¥{promoDiscount.toFixed(2)}（码：{promoCode}）
              </div>
            )}
          </div>
        )}

        {/* Sold out */}
        {isSoldOut && (
          <div style={{ marginTop: 16, textAlign: 'center', padding: '24px 0' }}>
            <div style={{ fontSize: 14, color: 'var(--color-text-tertiary)', marginBottom: 12 }}>所有票种已售罄</div>
            <WaitlistButton eventId={Number(id)} isSoldOut={isSoldOut} />
          </div>
        )}
      </div>

      {/* Marketplace listings */}
      {marketplaceListings.length > 0 && (
        <div>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 12 }}>二手票</div>
          <div style={{ border: '1px solid var(--color-border)' }}>
            {marketplaceListings.map((listing) => (
              <div
                key={listing.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '12px 20px',
                  borderBottom: '1px solid var(--color-border)',
                }}
              >
                <div>
                  <div style={{ fontWeight: 500, fontSize: 14 }}>{listing.ticket_name}</div>
                  <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)' }}>
                    卖家：{listing.seller_name}
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                  <span style={{ fontSize: 18, fontWeight: 700, fontFamily: 'var(--font-mono)' }}>¥{listing.price}</span>
                  <Button size="small" onClick={() => navigate('/marketplace')}>查看</Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
