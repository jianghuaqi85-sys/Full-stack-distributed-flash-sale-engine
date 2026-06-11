import { useEffect, useState, useCallback } from 'react'
import { Button, Tag, Space, Modal, Input, message, Empty } from 'antd'
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
    getPendingTransfers().then((res) => setTransfers(res.data.data || [])).finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchTransfers() }, [fetchTransfers])

  const handleApprove = async (id: number) => {
    try { await approveTransfer(id); message.success('审批通过'); fetchTransfers() } catch { /* */ }
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
        try { await rejectTransfer(id, reason); message.success('已拒绝'); fetchTransfers() } catch { /* */ }
      },
    })
  }

  return (
    <div className="page-enter">
      <div className="page-header">
        <h1>审核</h1>
        <div className="subtitle">转让申请审核</div>
      </div>

      {loading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>加载中...</div>
      ) : transfers.length === 0 ? (
        <Empty description="暂无待审核转让" />
      ) : (
        <div style={{ border: '1px solid var(--color-border)' }}>
          <div style={{
            display: 'grid',
            gridTemplateColumns: '60px 80px 100px 80px 80px 1fr 140px 140px',
            gap: 16,
            padding: '12px 20px',
            borderBottom: '1px solid var(--color-border)',
            fontSize: 12,
            color: 'var(--color-text-tertiary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontFamily: 'var(--font-mono)',
          }}>
            <div>ID</div><div>票务</div><div>类型</div><div>转让方</div><div>接收方</div><div>原因</div><div>时间</div><div>操作</div>
          </div>
          {transfers.map(t => (
            <div key={t.id} style={{
              display: 'grid',
              gridTemplateColumns: '60px 80px 100px 80px 80px 1fr 140px 140px',
              gap: 16,
              padding: '12px 20px',
              borderBottom: '1px solid var(--color-border)',
              alignItems: 'center',
              fontSize: 14,
            }}>
              <div style={{ fontFamily: 'var(--font-mono)', color: 'var(--color-text-tertiary)' }}>{t.id}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.ticket_id}</div>
              <div><Tag color={typeMap[t.transfer_type]?.color}>{typeMap[t.transfer_type]?.text}</Tag></div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.from_user_id}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{t.to_user_id}</div>
              <div style={{ color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.reason || '-'}</div>
              <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)' }}>{new Date(t.created_at).toLocaleDateString('zh-CN')}</div>
              <div>
                <Space size={4}>
                  <Button type="primary" size="small" onClick={() => handleApprove(t.id)}>通过</Button>
                  <Button danger size="small" onClick={() => handleReject(t.id)}>拒绝</Button>
                </Space>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
