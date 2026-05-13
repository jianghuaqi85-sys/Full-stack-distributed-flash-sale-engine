import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Tag, Button, Space, Modal, message, Empty, Select, QRCode } from 'antd'
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
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [qrModalVisible, setQrModalVisible] = useState(false)
  const [currentQRCode, setCurrentQRCode] = useState('')
  const navigate = useNavigate()

  const fetchTickets = useCallback(() => {
    setLoading(true)
    getMyTickets(page, 10)
      .then((res) => {
        setTickets(res.data.data)
        setTotal(res.data.total)
      })
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => {
    fetchTickets()
  }, [fetchTickets])

  const handlePay = async (id: number) => {
    try {
      await payTicket(id)
      message.success('支付成功')
      fetchTickets()
    } catch {
      // handled by interceptor
    }
  }

  const handleCancel = (id: number) => {
    Modal.confirm({
      title: '确认取消',
      content: '取消后库存将自动恢复，确定要取消吗？',
      onOk: async () => {
        try {
          await cancelTicket(id)
          message.success('取消成功')
          fetchTickets()
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const handleUse = (id: number) => {
    Modal.confirm({
      title: '确认使用',
      content: '使用后此票将标记为已使用，确定吗？',
      onOk: async () => {
        try {
          await useTicket(id)
          message.success('核销成功')
          fetchTickets()
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const showQRCode = (qrCode: string) => {
    setCurrentQRCode(qrCode)
    setQrModalVisible(true)
  }

  const filteredTickets = statusFilter === 'all'
    ? tickets
    : tickets.filter(t => t.status === statusFilter)

  const columns = [
    {
      title: '订单号',
      dataIndex: 'order_no',
      key: 'order_no',
      render: (text: string) => text || '-',
    },
    {
      title: '票种',
      dataIndex: 'ticket_name',
      key: 'ticket_name',
    },
    {
      title: '数量',
      dataIndex: 'quantity',
      key: 'quantity',
    },
    {
      title: '总价',
      dataIndex: 'total_price',
      key: 'total_price',
      render: (price: number) => (
        <span style={{ color: 'var(--color-gold)', fontWeight: 600 }}>¥{price}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={statusMap[status]?.color}>{statusMap[status]?.text}</Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Ticket) => (
        <Space>
          {record.status === 'reserved' && (
            <>
              <Button type="primary" size="small" onClick={() => handlePay(record.id)}>
                支付
              </Button>
              <Button danger size="small" onClick={() => handleCancel(record.id)}>
                取消
              </Button>
            </>
          )}
          {record.status === 'paid' && (
            <>
              <Button size="small" onClick={() => record.qr_code && showQRCode(record.qr_code)}>
                二维码
              </Button>
              <Button size="small" onClick={() => handleUse(record.id)}>
                使用
              </Button>
              <TransferButton ticketId={record.id} onSuccess={fetchTickets} />
              <Button danger size="small" onClick={() => handleCancel(record.id)}>
                退票
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="我的票务"
        extra={
          <Space>
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              style={{ width: 120 }}
              options={[
                { value: 'all', label: '全部状态' },
                { value: 'reserved', label: '待支付' },
                { value: 'paid', label: '已支付' },
                { value: 'used', label: '已使用' },
                { value: 'expired', label: '已过期' },
                { value: 'cancelled', label: '已取消' },
              ]}
            />
            <Button type="primary" onClick={() => navigate('/events')}>
              去购票
            </Button>
          </Space>
        }
      >
        {filteredTickets.length === 0 && !loading ? (
          <Empty description="暂无票务" />
        ) : (
          <Table
            columns={columns}
            dataSource={filteredTickets}
            rowKey="id"
            loading={loading}
            onRow={(record) => ({
              style: { cursor: 'pointer' },
              onClick: () => navigate(`/tickets/${record.id}`),
            })}
            pagination={{
              current: page,
              total,
              pageSize: 10,
              onChange: setPage,
              showTotal: (t) => `共 ${t} 张票`,
            }}
          />
        )}
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
          <QRCode
            value={currentQRCode}
            size={180}
            style={{ marginBottom: 16 }}
          />
          <p style={{ color: 'var(--color-text-secondary)', fontSize: 13, margin: 0 }}>
            请在入场时出示此二维码
          </p>
        </div>
      </Modal>
    </div>
  )
}
