import { useEffect, useState } from 'react'
import { Card, Table, Tag, Empty } from 'antd'
import { SwapOutlined } from '@ant-design/icons'
import { getTransferHistory, TicketTransfer } from '../../api/transfer'

const statusMap: Record<string, { color: string; text: string }> = {
  pending: { color: 'processing', text: '待审核' },
  approved: { color: 'success', text: '已通过' },
  rejected: { color: 'error', text: '已拒绝' },
}

const typeMap: Record<string, { color: string; text: string }> = {
  gift: { color: 'blue', text: '转赠' },
  marketplace: { color: 'orange', text: '二手交易' },
}

export default function TransferRecords() {
  const [transfers, setTransfers] = useState<TicketTransfer[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    getTransferHistory()
      .then((res) => setTransfers(res.data.data || []))
      .finally(() => setLoading(false))
  }, [])

  const columns = [
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
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const s = statusMap[status] || { color: 'default', text: status }
        return <Tag color={s.color}>{s.text}</Tag>
      },
    },
    {
      title: '原因',
      dataIndex: 'reason',
      key: 'reason',
      ellipsis: true,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
  ]

  return (
    <Card
      title={<><SwapOutlined /> 转让记录</>}
      style={{ borderRadius: 16 }}
    >
      <Table
        columns={columns}
        dataSource={transfers}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: <Empty description="暂无转让记录" /> }}
      />
    </Card>
  )
}
