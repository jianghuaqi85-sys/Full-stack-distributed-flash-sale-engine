import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button, Tag, Spin, QRCode, Modal, message } from 'antd'
import { getTicketDetail, payTicket, cancelTicket, useTicket, Ticket } from '../../api/tickets'
import TransferButton from '../../components/TransferButton'

const statusMap: Record<string, { color: string; text: string }> = {
  reserved: { color: 'processing', text: '待支付' },
  paid: { color: 'success', text: '已支付' },
  used: { color: 'default', text: '已使用' },
  expired: { color: 'warning', text: '已过期' },
  cancelled: { color: 'error', text: '已取消' },
}

export default function TicketDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [loading, setLoading] = useState(true)
  const [qrModalVisible, setQrModalVisible] = useState(false)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    getTicketDetail(Number(id)).then((res) => setTicket(res.data)).finally(() => setLoading(false))
  }, [id])

  const refresh = () => { if (ticket) getTicketDetail(ticket.id).then((res) => setTicket(res.data)) }

  const handlePay = async () => { if (!ticket) return; try { await payTicket(ticket.id); message.success('支付成功'); refresh() } catch { /* */ } }
  const handleCancel = () => {
    Modal.confirm({ title: '确认取消', content: '取消后库存将自动恢复，确定要取消吗？', onOk: async () => { if (!ticket) return; try { await cancelTicket(ticket.id); message.success('取消成功'); refresh() } catch { /* */ } } })
  }
  const handleUse = () => {
    Modal.confirm({ title: '确认使用', content: '使用后此票将标记为已使用，确定吗？', onOk: async () => { if (!ticket) return; try { await useTicket(ticket.id); message.success('核销成功'); refresh() } catch { /* */ } } })
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!ticket) return <div>票务不存在</div>

  return (
    <div className="page-enter">
      <button onClick={() => navigate('/tickets')} style={{ border: 'none', background: 'transparent', color: 'var(--color-text-tertiary)', cursor: 'pointer', fontSize: 13, fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.05em', padding: 0, marginBottom: 24 }}>
        ← 返回列表
      </button>

      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <Tag color={statusMap[ticket.status]?.color}>{statusMap[ticket.status]?.text}</Tag>
        </div>
        <h1 style={{ fontSize: 32, fontWeight: 700, letterSpacing: '-0.03em', margin: 0 }}>{ticket.ticket_name}</h1>
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
          { label: '订单号', value: ticket.order_no || '-', mono: true },
          { label: '票种', value: ticket.ticket_name },
          { label: '数量', value: String(ticket.quantity), mono: true },
          { label: '总价', value: `¥${ticket.total_price}`, mono: true },
          { label: '创建时间', value: new Date(ticket.created_at).toLocaleString('zh-CN'), mono: true },
        ].map(item => (
          <div key={item.label} style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>{item.label}</div>
            <div style={{ fontSize: 15, fontWeight: 500, fontFamily: item.mono ? 'var(--font-mono)' : 'inherit' }}>{item.value}</div>
          </div>
        ))}
      </div>

      {/* QR code for paid tickets */}
      {ticket.status === 'paid' && ticket.qr_code && (
        <div style={{ marginBottom: 32, padding: '24px', border: '1px solid var(--color-border)', textAlign: 'center' }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 16 }}>入场凭证</div>
          <QRCode value={ticket.qr_code} size={160} style={{ marginBottom: 12 }} />
          <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, margin: 0 }}>请在入场时出示此二维码</p>
        </div>
      )}

      {/* Actions */}
      <div style={{ display: 'flex', gap: 8 }}>
        {ticket.status === 'reserved' && (
          <>
            <Button type="primary" onClick={handlePay}>支付</Button>
            <Button danger onClick={handleCancel}>取消</Button>
          </>
        )}
        {ticket.status === 'paid' && (
          <>
            <Button onClick={() => setQrModalVisible(true)}>查看入场凭证</Button>
            <Button onClick={handleUse}>使用</Button>
            <TransferButton ticketId={ticket.id} onSuccess={refresh} />
            <Button danger onClick={handleCancel}>退票</Button>
          </>
        )}
      </div>

      <Modal title="入场凭证" open={qrModalVisible} onCancel={() => setQrModalVisible(false)} footer={null} width={360}>
        <div style={{ textAlign: 'center', padding: 24 }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 16 }}>入场凭证</div>
          <QRCode value={ticket.qr_code || ''} size={180} style={{ marginBottom: 16 }} />
          <p style={{ color: 'var(--color-text-tertiary)', fontSize: 13, margin: 0 }}>请在入场时出示此二维码</p>
        </div>
      </Modal>
    </div>
  )
}
