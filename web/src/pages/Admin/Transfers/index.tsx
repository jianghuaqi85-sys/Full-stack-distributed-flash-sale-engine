import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Input, message, Empty } from 'antd'
import { CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { getPendingTransfers, approveTransfer, rejectTransfer, TicketTransfer } from '../../../api/transfer'

const typeMap: Record<string, { color: string; text: string }> = {
  gift: { color: 'blue', text: '转赠' },
  marketplace: { color: 'orange', text: '二手交易' },
}

export default function AdminTransfers() {
  const [transfers, setTransfers] = useState<TicketTransfer[]>([])
  const [loading, setLoading] = useState(false)

  const fetchTransfers = useCallback(() => {
    setLoading(true)
    getPendingTransfers()
      .then((res) => setTransfers(res.data.data || []))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchTransfers()
  }, [fetchTransfers])

  const handleApprove = async (id: number) => {
    try {
      await approveTransfer(id)
      message.success('审批通过')
      fetchTransfers()
    } catch {
      // handled by interceptor
    }
  }

  const handleReject = (id: number) => {
    Modal.confirm({
      title: '拒绝转让',
      content: (
        <div>
          <p>确定要拒绝此转让申请吗？</p>
          <Input.TextArea id="reject-reason" placeholder="拒绝原因（选填）" rows={3} />
        </div>
      ),
      onOk: async () => {
        const reason = (document.getElementById('reject-reason') as HTMLTextAreaElement)?.value
        try {
          await rejectTransfer(id, reason)
          message.success('已拒绝')
          fetchTransfers()
        } catch {
          // handled by interceptor
        }
      },
    })
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '票务ID',
      dataIndex: 'ticket_id',
      key: 'ticket_id',
    },
    {
      title: '转让类型',
      dataIndex: 'transfer_type',
      key: 'transfer_type',
      render: (type: string) => {
        const t = typeMap[type] || { color: 'default', text: type }
        return <Tag color={t.color}>{t.text}</Tag>
      },
    },
    {
      title: '转让方',
      dataIndex: 'from_user_id',
      key: 'from_user_id',
    },
    {
      title: '接收方',
      dataIndex: 'to_user_id',
      key: 'to_user_id',
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: TicketTransfer) => (
        <Space>
          <Button
            type="primary"
            size="small"
            icon={<CheckOutlined />}
            onClick={() => handleApprove(record.id)}
          >
            通过
          </Button>
          <Button
            danger
            size="small"
            icon={<CloseOutlined />}
            onClick={() => handleReject(record.id)}
          >
            拒绝
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card title="转让审核" style={{ borderRadius: 16 }}>
      <Table
        columns={columns}
        dataSource={transfers}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: <Empty description="暂无待审核转让" /> }}
      />
    </Card>
  )
}
