import { useEffect, useState, useCallback } from 'react'
import { Button, Tag, Modal, Form, InputNumber, Input, Select, message, Empty } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getActiveListings, getMyListings, getMyPurchases, createListing, buyListing, cancelListing, MarketplaceListing } from '../../api/marketplace'
import { getMyTickets } from '../../api/tickets'
import { useAuthStore } from '../../stores/authStore'

interface MyTicket { id: number; order_no: string; ticket_name: string; status: string }

const statusMap: Record<string, { color: string; text: string }> = {
  active: { color: 'success', text: '在售' },
  sold: { color: 'default', text: '已售' },
  cancelled: { color: 'warning', text: '已下架' },
}

export default function Marketplace() {
  const { user } = useAuthStore()
  const navigate = useNavigate()
  const [listings, setListings] = useState<MarketplaceListing[]>([])
  const [myListings, setMyListings] = useState<MarketplaceListing[]>([])
  const [myPurchases, setMyPurchases] = useState<MarketplaceListing[]>([])
  const [total, setTotal] = useState(0)
  const [page] = useState(1)
  const [loading, setLoading] = useState(false)
  const [tab, setTab] = useState<'market' | 'my-sell' | 'my-buy'>('market')
  const [sellModalVisible, setSellModalVisible] = useState(false)
  const [form] = Form.useForm()
  const [myTickets, setMyTickets] = useState<MyTicket[]>([])

  const fetchListings = useCallback(() => {
    setLoading(true)
    getActiveListings(page, 10).then((res) => { setListings(res.data.data || []); setTotal(res.data.total) }).finally(() => setLoading(false))
  }, [page])

  const fetchMyData = useCallback(() => {
    getMyListings().then((res) => setMyListings(res.data.data || [])).catch(() => {})
    getMyPurchases().then((res) => setMyPurchases(res.data.data || [])).catch(() => {})
  }, [])

  const fetchMyTickets = useCallback(() => {
    getMyTickets(1, 100).then((res) => {
      setMyTickets((res.data.data || []).filter((t: MyTicket) => t.status === 'paid'))
    }).catch(() => {})
  }, [])

  useEffect(() => { fetchListings() }, [fetchListings])
  useEffect(() => { fetchMyData(); fetchMyTickets() }, [fetchMyData, fetchMyTickets])

  const handleBuy = async (id: number) => {
    try { await buyListing(id); message.success('购买成功'); fetchListings(); fetchMyData() } catch { /* */ }
  }
  const handleCancel = async (id: number) => {
    try { await cancelListing(id); message.success('下架成功'); fetchMyData() } catch { /* */ }
  }
  const handleSell = async () => {
    try {
      const values = await form.validateFields()
      await createListing(values)
      message.success('上架成功'); setSellModalVisible(false); form.resetFields(); fetchMyData(); fetchMyTickets()
    } catch { /* */ }
  }

  const tabs = [
    { key: 'market' as const, label: '在售', count: total },
    { key: 'my-sell' as const, label: '我的上架', count: myListings.length },
    { key: 'my-buy' as const, label: '我的购买', count: myPurchases.length },
  ]

  return (
    <div className="page-enter">
      <div className="page-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h1>市场</h1>
            <div className="subtitle">二手票务交易</div>
          </div>
          {tab === 'my-sell' && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => { fetchMyTickets(); setSellModalVisible(true) }}>上架票务</Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div style={{ display: 'flex', gap: 1, background: 'var(--color-border)', border: '1px solid var(--color-border)', marginBottom: 24 }}>
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            style={{
              flex: 1,
              padding: '12px 16px',
              border: 'none',
              background: tab === t.key ? 'var(--color-bg-container)' : 'var(--color-bg-base)',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: tab === t.key ? 600 : 400,
              color: tab === t.key ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)',
              transition: 'all 0.1s',
            }}
          >
            {t.label} <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, marginLeft: 4 }}>({t.count})</span>
          </button>
        ))}
      </div>

      {/* Market listings */}
      {tab === 'market' && (
        loading ? <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>加载中...</div>
        : listings.length === 0 ? <Empty description="暂无在售票务" />
        : (
          <div style={{ border: '1px solid var(--color-border)' }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 100px 1fr 140px 140px', gap: 16, padding: '12px 20px', borderBottom: '1px solid var(--color-border)', fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)' }}>
              <div>票种</div><div>售价</div><div>描述</div><div>时间</div><div>操作</div>
            </div>
            {listings.map(l => (
              <div key={l.id} style={{ display: 'grid', gridTemplateColumns: '1fr 100px 1fr 140px 140px', gap: 16, padding: '12px 20px', borderBottom: '1px solid var(--color-border)', alignItems: 'center', fontSize: 14 }}>
                <div style={{ fontWeight: 500 }}>{l.ticket_name}</div>
                <div style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>¥{l.price}</div>
                <div style={{ color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{l.description || '-'}</div>
                <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)' }}>{new Date(l.created_at).toLocaleDateString('zh-CN')}</div>
                <div style={{ display: 'flex', gap: 4 }}>
                  <Button size="small" onClick={() => navigate(`/marketplace/${l.id}`)}>详情</Button>
                  {l.seller_id !== user?.id && <Button type="primary" size="small" onClick={() => handleBuy(l.id)}>购买</Button>}
                </div>
              </div>
            ))}
          </div>
        )
      )}

      {/* My listings */}
      {tab === 'my-sell' && (
        myListings.length === 0 ? <Empty description="暂无上架记录" />
        : (
          <div style={{ border: '1px solid var(--color-border)' }}>
            {myListings.map(l => (
              <div key={l.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '14px 20px', borderBottom: '1px solid var(--color-border)' }}>
                <div>
                  <div style={{ fontWeight: 500 }}>{l.ticket_name}</div>
                  <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}>¥{l.price}</div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Tag color={statusMap[l.status]?.color}>{statusMap[l.status]?.text}</Tag>
                  {l.status === 'active' && <Button danger size="small" onClick={() => handleCancel(l.id)}>下架</Button>}
                </div>
              </div>
            ))}
          </div>
        )
      )}

      {/* My purchases */}
      {tab === 'my-buy' && (
        myPurchases.length === 0 ? <Empty description="暂无购买记录" />
        : (
          <div style={{ border: '1px solid var(--color-border)' }}>
            {myPurchases.map(l => (
              <div key={l.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '14px 20px', borderBottom: '1px solid var(--color-border)' }}>
                <div style={{ fontWeight: 500 }}>{l.ticket_name}</div>
                <div style={{ display: 'flex', gap: 16, alignItems: 'center' }}>
                  <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>¥{l.price}</span>
                  <span style={{ fontSize: 13, color: 'var(--color-text-tertiary)', fontFamily: 'var(--font-mono)' }}>{new Date(l.created_at).toLocaleDateString('zh-CN')}</span>
                </div>
              </div>
            ))}
          </div>
        )
      )}

      <Modal title="上架票务" open={sellModalVisible} onOk={handleSell} onCancel={() => setSellModalVisible(false)} width={500}>
        <Form form={form} layout="vertical">
          <Form.Item name="ticket_id" label="选择票务" rules={[{ required: true }]}>
            <Select placeholder="选择票务" options={myTickets.map(t => ({ value: t.id, label: `${t.ticket_name} (${t.order_no})` }))} />
          </Form.Item>
          <Form.Item name="price" label="出售价格" rules={[{ required: true }]}>
            <InputNumber min={0.01} step={0.01} prefix="¥" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="票务描述（可选）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
