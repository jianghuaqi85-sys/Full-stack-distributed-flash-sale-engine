import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Button, Tag, Spin, Descriptions, Space, message, Empty } from 'antd'
import { ArrowLeftOutlined, ShoppingOutlined } from '@ant-design/icons'
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
    getListing(Number(id))
      .then((res) => setListing(res.data))
      .finally(() => setLoading(false))
  }, [id])

  const handleBuy = async () => {
    if (!listing) return
    setBuying(true)
    try {
      await buyListing(listing.id)
      message.success('购买成功')
      const res = await getListing(listing.id)
      setListing(res.data)
    } catch {
      // handled by interceptor
    } finally {
      setBuying(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!listing) return <Empty description="票务不存在" />

  return (
    <div className="animate-slide-up">
      <Button icon={<ArrowLeftOutlined />} style={{ marginBottom: 16, borderRadius: 8 }} onClick={() => navigate('/marketplace')}>
        返回市场
      </Button>

      <Card style={{ borderRadius: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
          <div style={{ flex: 1, minWidth: 300 }}>
            <h2 style={{ margin: '0 0 16px', fontSize: 22, fontWeight: 700 }}>
              {listing.ticket_name}
              <Tag color={statusMap[listing.status]?.color} style={{ marginLeft: 12, fontSize: 14 }}>
                {statusMap[listing.status]?.text}
              </Tag>
            </h2>

            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="活动">{listing.event_title}</Descriptions.Item>
              <Descriptions.Item label="票种">{listing.ticket_name}</Descriptions.Item>
              <Descriptions.Item label="售价">
                <span style={{ color: 'var(--color-gold)', fontWeight: 600, fontSize: 18 }}>
                  ¥{listing.price}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label="卖家">{listing.seller_name}</Descriptions.Item>
              {listing.description && (
                <Descriptions.Item label="描述">{listing.description}</Descriptions.Item>
              )}
              <Descriptions.Item label="发布时间">
                {new Date(listing.created_at).toLocaleString('zh-CN')}
              </Descriptions.Item>
            </Descriptions>

            <Space style={{ marginTop: 24 }}>
              {listing.status === 'active' && listing.seller_id !== user?.id && (
                <Button
                  type="primary"
                  icon={<ShoppingOutlined />}
                  loading={buying}
                  onClick={handleBuy}
                  style={{ height: 44, borderRadius: 10, fontWeight: 600 }}
                >
                  购买
                </Button>
              )}
              <Button onClick={() => navigate('/marketplace')}>
                返回市场
              </Button>
            </Space>
          </div>
        </div>
      </Card>
    </div>
  )
}
