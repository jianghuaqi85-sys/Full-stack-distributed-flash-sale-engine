import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Tag, Spin, Empty, message } from 'antd'
import { getListing, buyListing, MarketplaceListing } from '../../api/marketplace'
import { useAuthStore } from '../../stores/authStore'

const statusMap: Record<string, { color: string; text: string }> = {
  active: { color: 'success', text: '在售' },
  sold: { color: 'default', text: '已售' },
  cancelled: { color: 'warning', text: '已下架' },
}

export default function MarketplaceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [listing, setListing] = useState<MarketplaceListing | null>(null)
  const [loading, setLoading] = useState(true)
  const [buying, setBuying] = useState(false)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    getListing(Number(id)).then((res) => setListing(res.data)).finally(() => setLoading(false))
  }, [id])

  const handleBuy = async () => {
    if (!listing) return
    setBuying(true)
    try {
      await buyListing(listing.id)
      message.success('购买成功')
      const res = await getListing(listing.id)
      setListing(res.data)
    } catch { /* */ } finally { setBuying(false) }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!listing) return <Empty description="票务不存在" />

  return (
    <div className="page-enter">
      <button onClick={() => navigate('/marketplace')} style={{ border: 'none', background: 'transparent', color: 'var(--color-text-tertiary)', cursor: 'pointer', fontSize: 13, fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.05em', padding: 0, marginBottom: 24 }}>
        ← 返回市场
      </button>

      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <Tag color={statusMap[listing.status]?.color}>{statusMap[listing.status]?.text}</Tag>
        </div>
        <h1 style={{ fontSize: 32, fontWeight: 700, letterSpacing: '-0.03em', margin: 0 }}>{listing.ticket_name}</h1>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
        gap: 1,
        background: 'var(--color-border)',
        border: '1px solid var(--color-border)',
        marginBottom: 32,
      }}>
        {[
          { label: '活动', value: listing.event_title || '-' },
          { label: '售价', value: `¥${listing.price}`, mono: true },
          { label: '卖家', value: listing.seller_name || '-' },
          { label: '描述', value: listing.description || '-' },
          { label: '发布时间', value: new Date(listing.created_at).toLocaleString('zh-CN'), mono: true },
        ].map(item => (
          <div key={item.label} style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>{item.label}</div>
            <div style={{ fontSize: 15, fontWeight: 500, fontFamily: item.mono ? 'var(--font-mono)' : 'inherit' }}>{item.value}</div>
          </div>
        ))}
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        {listing.status === 'active' && listing.seller_id !== user?.id && (
          <Button type="primary" loading={buying} onClick={handleBuy}>购买</Button>
        )}
        <Button onClick={() => navigate('/marketplace')}>返回市场</Button>
      </div>
    </div>
  )
}
