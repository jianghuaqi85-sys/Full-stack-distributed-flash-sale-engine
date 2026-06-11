import { useEffect, useState, useCallback } from 'react'
import { Tag, Button, Space, Modal, message, Empty, Select, QRCode } from 'antd'
import { useNavigate } from 'react-router-dom'
import { Ticket, getMyTickets, payTicket, cancelTicket, useTicket } from '../../api/tickets'
import TransferButton from '../../components/TransferButton'

const statusMap: Record<string, { color: string; text: string }> = {
  reserved: { color: 'processing', text: '待支付' },
  paid: { color: 'success', text: '已支付' },
  used: { color: 'default', text: '已使用' },
  expired: { color: 'warning', text: '已过期' },
  cancelled: { color: 'error', text: '已取消' },
}

export default function Tickets() {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [, setTotal] = useState(0)
  const [page] = useState(1)
  const [loading, setLoading] = useState(false)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [qrModalVisible, setQrModalVisible] = useState(false)
  const [currentQRCode, setCurrentQRCode] = useState('')
  const navigate = useNavigate()

  const fetchTickets = useCallback(() => {
    setLoading(true)
    getMyTickets(page, 20)
      .then((res) => { setTickets(res.data.data); setTotal(res.data.total) })
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => { fetchTickets() }, [fetchTickets])

  const handlePay = async (id: number) => {
    try { await payTicket(id); message.success('支付成功'); fetchTickets() } catch { /* handled */ }
  }

  const handleCancel = (id: number) => {
    Modal.confirm({
      title: '确认取消',
      content: '取消后库存将自动恢复，确定要取消吗？',
      onOk: async () => {
        try { await cancelTicket(id); message.success('取消成功'); fetchTickets() } catch { /* handled */ }
      },
    })
  }

  const handleUse = (id: number) => {
    Modal.confirm({
      title: '确认使用',
      content: '使用后此票将标记为已使用，确定吗？',
      onOk: async () => {
        try { await useTicket(id); message.success('核销成功'); fetchTickets() } catch { /* handled */ }
      },
    })
  }

  const filteredTickets = statusFilter === 'all' ? tickets : tickets.filter(t => t.status === statusFilter)

  return (
    <div className="page-enter">
      <div className="page-header">
        <h1>票务</h1>
        <div className="subtitle">管理您的票务</div>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          style={{ width: 120 }}
          options={[
            { value: 'all', label: '全部' },
            { value: 'reserved', label: '待支付' },
            { value: 'paid', label: '已支付' },
            { value: 'used', label: '已使用' },
            { value: 'expired', label: '已过期' },
            { value: 'cancelled', label: '已取消' },
          ]}
        />
        <Button type="primary" onClick={() => navigate('/events')}>去购票</Button>
      </div>

      {/* Content */}
      {filteredTickets.length === 0 && !loading ? (
        <Empty description="暂无票务" />
      ) : (
        <>
          {/* Table header */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: '1fr 120px 60px 100px 100px 140px',
            gap: 16,
            padding: '12px 20px',
            background: 'var(--color-bg-container)',
            borderBottom: '1px solid var(--color-border)',
            fontSize: 12,
            color: 'var(--color-text-tertiary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontFamily: 'var(--font-mono)',
            border: '1px solid var(--color-border)',
          }}>
            <div>订单号</div>
            <div>票种</div>
            <div>数量</div>
            <div>总价</div>
            <div>状态</div>
            <div>操作</div>
          </div>

          {/* Rows */}
          {filteredTickets.map((ticket) => (
            <div
              key={ticket.id}
              style={{
                display: 'grid',
                gridTemplateColumns: '1fr 120px 60px 100px 100px 140px',
                gap: 16,
                padding: '14px 20px',
                borderBottom: '1px solid var(--color-border)',
                borderLeft: '1px solid var(--color-border)',
                borderRight: '1px solid var(--color-border)',
                alignItems: 'center',
                cursor: 'pointer',
                transition: 'background 0.1s',
              }}
              onClick={() => navigate(`/tickets/${ticket.id}`)}
              onMouseEnter={e => e.currentTarget.style.background = 'var(--color-accent-soft)'}
              onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
            >
              <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)', color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {ticket.order_no || '-'}
              </div>
              <div style={{ fontSize: 14, fontWeight: 500 }}>{ticket.ticket_name}</div>
              <div style={{ fontSize: 14, fontFamily: 'var(--font-mono)' }}>{ticket.quantity}</div>
              <div style={{ fontSize: 14, fontFamily: 'var(--font-mono)', fontWeight: 600 }}>¥{ticket.total_price}</div>
              <div><Tag color={statusMap[ticket.status]?.color}>{statusMap[ticket.status]?.text}</Tag></div>
              <div onClick={e => e.stopPropagation()}>
                <Space size={4}>
                  {ticket.status === 'reserved' && (
                    <>
                      <Button type="primary" size="small" onClick={() => handlePay(ticket.id)}>支付</Button>
                      <Button danger size="small" onClick={() => handleCancel(ticket.id)}>取消</Button>
                    </>
                  )}
                  {ticket.status === 'paid' && (
                    <>
                      <Button size="small" onClick={() => { setCurrentQRCode(ticket.qr_code || ''); setQrModalVisible(true) }}>二维码</Button>
                      <Button size="small" onClick={() => handleUse(ticket.id)}>使用</Button>
                      <TransferButton ticketId={ticket.id} onSuccess={fetchTickets} />
                    </>
                  )}
                </Space>
              </div>
            </div>
          ))}
        </>
      )}

      {/* QR Modal */}
      <Modal
        title="入场凭证"
        open={qrModalVisible}
        onCancel={() => setQrModalVisible(false)}
        footer={null}
        width={360}
      >
        <div style={{ textAlign: 'center', padding: 24 }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 16 }}>
            入场凭证
          </div>
          <QRCode value={currentQRCode} size={180} style={{ marginBottom: 16 }} />
          <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, margin: 0 }}>请在入场时出示此二维码</p>
        </div>
      </Modal>
    </div>
  )
}
