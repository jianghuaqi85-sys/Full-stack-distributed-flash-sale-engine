import { useEffect, useState } from 'react'
import { Tag, Empty } from 'antd'
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
    getTransferHistory().then((res) => setTransfers(res.data.data || [])).finally(() => setLoading(false))
  }, [])

  return (
    <div className="page-enter">
      <div className="page-header">
        <h1>转让</h1>
        <div className="subtitle">转让记录</div>
      </div>

      {loading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>加载中...</div>
      ) : transfers.length === 0 ? (
        <Empty description="暂无转让记录" />
      ) : (
        <div style={{ border: '1px solid var(--color-border)' }}>
          <div style={{
            display: 'grid',
            gridTemplateColumns: '80px 100px 80px 80px 80px 1fr 140px',
            gap: 16,
            padding: '12px 20px',
            borderBottom: '1px solid var(--color-border)',
            fontSize: 12,
            color: 'var(--color-text-tertiary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontFamily: 'var(--font-mono)',
          }}>
            <div>票务</div><div>类型</div><div>转让方</div><div>接收方</div><div>状态</div><div>原因</div><div>时间</div>
          </div>
          {transfers.map(t => (
            <div key={t.id} style={{
              display: 'grid',
              gridTemplateColumns: '80px 100px 80px 80px 80px 1fr 140px',
              gap: 16,
              padding: '12px 20px',
              borderBottom: '1px solid var(--color-border)',
              alignItems: 'center',
              fontSize: 14,
            }}>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.ticket_id}</div>
              <div><Tag color={typeMap[t.transfer_type]?.color}>{typeMap[t.transfer_type]?.text}</Tag></div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.from_user_id}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.to_user_id}</div>
              <div><Tag color={statusMap[t.status]?.color}>{statusMap[t.status]?.text}</Tag></div>
              <div style={{ color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.reason || '-'}</div>
              <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)' }}>{new Date(t.created_at).toLocaleDateString('zh-CN')}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
