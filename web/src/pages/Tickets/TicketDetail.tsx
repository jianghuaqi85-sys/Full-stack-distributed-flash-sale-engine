import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, Button, Tag, Spin, Descriptions, QRCode, Space, Modal, message } from 'antd'
import { ArrowLeftOutlined, QrcodeOutlined } from '@ant-design/icons'
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
    getTicketDetail(Number(id))
      .then((res) => setTicket(res.data))
      .finally(() => setLoading(false))
  }, [id])

  const handlePay = async () => {
    if (!ticket) return
    try {
      await payTicket(ticket.id)
      message.success('支付成功')
      const res = await getTicketDetail(ticket.id)
      setTicket(res.data)
    } catch {
      // handled by interceptor
    }
  }

  const handleCancel = () => {
    Modal.confirm({
      title: '确认取消',
      content: '取消后库存将自动恢复，确定要取消吗？',
      onOk: async () => {
        if (!ticket) return
        try {
          await cancelTicket(ticket.id)
          message.success('取消成功')
          const res = await getTicketDetail(ticket.id)
          setTicket(res.data)
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const handleUse = () => {
    Modal.confirm({
      title: '确认使用',
      content: '使用后此票将标记为已使用，确定吗？',
      onOk: async () => {
        if (!ticket) return
        try {
          await useTicket(ticket.id)
          message.success('核销成功')
          const res = await getTicketDetail(ticket.id)
          setTicket(res.data)
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!ticket) return <div>票务不存在</div>

  return (
    <div className="animate-slide-up">
      <Button icon={<ArrowLeftOutlined />} style={{ marginBottom: 16, borderRadius: 8 }} onClick={() => navigate('/tickets')}>
        返回列表
      </Button>

      <Card style={{ borderRadius: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
          <div style={{ flex: 1, minWidth: 300 }}>
            <h2 style={{ margin: '0 0 16px', fontSize: 22, fontWeight: 700 }}>
              票务详情
              <Tag color={statusMap[ticket.status]?.color} style={{ marginLeft: 12, fontSize: 14 }}>
                {statusMap[ticket.status]?.text}
              </Tag>
            </h2>

            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="订单号">{ticket.order_no || '-'}</Descriptions.Item>
              <Descriptions.Item label="票种">{ticket.ticket_name}</Descriptions.Item>
              <Descriptions.Item label="数量">{ticket.quantity}</Descriptions.Item>
              <Descriptions.Item label="总价">
                <span style={{ color: 'var(--color-gold)', fontWeight: 600, fontSize: 16 }}>
                  ¥{ticket.total_price}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(ticket.created_at).toLocaleString('zh-CN')}
              </Descriptions.Item>
            </Descriptions>

            <Space style={{ marginTop: 24 }}>
              {ticket.status === 'reserved' && (
                <>
                  <Button type="primary" onClick={handlePay}>支付</Button>
                  <Button danger onClick={handleCancel}>取消</Button>
                </>
              )}
              {ticket.status === 'paid' && (
                <>
                  <Button icon={<QrcodeOutlined />} onClick={() => setQrModalVisible(true)}>
                    查看入场凭证
                  </Button>
                  <Button onClick={handleUse}>使用</Button>
                  <TransferButton ticketId={ticket.id} onSuccess={() => {
                    getTicketDetail(ticket.id).then((res) => setTicket(res.data))
                  }} />
                  <Button danger onClick={handleCancel}>退票</Button>
                </>
              )}
            </Space>
          </div>

          {ticket.status === 'paid' && ticket.qr_code && (
            <div style={{
              textAlign: 'center',
              padding: 24,
              border: '2px dashed var(--color-border)',
              borderRadius: 16,
              background: 'var(--color-bg-layout)',
              minWidth: 240,
            }}>
              <div className="gradient-text" style={{ fontSize: 16, fontWeight: 700, marginBottom: 16 }}>
                入场凭证
              </div>
              <QRCode value={ticket.qr_code} size={160} style={{ marginBottom: 12 }} />
              <p style={{ color: 'var(--color-text-secondary)', fontSize: 12, margin: 0 }}>
                请在入场时出示此二维码
              </p>
            </div>
          )}
        </div>
      </Card>

      <Modal
        title="入场凭证"
        open={qrModalVisible}
        onCancel={() => setQrModalVisible(false)}
        footer={null}
        width={360}
      >
        <div style={{
          textAlign: 'center',
          padding: 24,
          border: '2px dashed var(--color-border)',
          borderRadius: 16,
          background: 'var(--color-bg-layout)',
        }}>
          <div className="gradient-text" style={{ fontSize: 18, fontWeight: 700, marginBottom: 16 }}>
            入场凭证
          </div>
          <QRCode value={ticket.qr_code || ''} size={180} style={{ marginBottom: 16 }} />
          <p style={{ color: 'var(--color-text-secondary)', fontSize: 13, margin: 0 }}>
            请在入场时出示此二维码
          </p>
        </div>
      </Modal>
    </div>
  )
}
