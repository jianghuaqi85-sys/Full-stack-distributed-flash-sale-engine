import { useEffect, useState, useCallback } from 'react'
import { Button, Tag, Modal, Form, Input, InputNumber, DatePicker, Select, message, Empty, Popconfirm } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { getPromoCodes, createPromoCode, deletePromoCode, PromoCode } from '../../../api/promo'
import { getEvents, Event } from '../../../api/events'

export default function AdminPromoCodes() {
  const [promoCodes, setPromoCodes] = useState<PromoCode[]>([])
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  const fetchPromoCodes = useCallback(() => {
    setLoading(true)
    getEvents(1, 100).then(async (res) => {
      const allEvents = res.data.data || []
      setEvents(allEvents)
      const allPromos: PromoCode[] = []
      for (const evt of allEvents) {
        try {
          const promoRes = await getPromoCodes(evt.id)
          if (promoRes.data.data) allPromos.push(...promoRes.data.data)
        } catch { /* ignore */ }
      }
      setPromoCodes(allPromos)
    }).finally(() => setLoading(false))
  }, [])

  useEffect(() => { fetchPromoCodes() }, [fetchPromoCodes])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createPromoCode({
        code: values.code, event_id: values.event_id, discount_type: values.discount_type,
        discount_value: values.discount_value, min_amount: values.min_amount || 0,
        max_uses: values.max_uses || 100, start_time: values.start_time?.toISOString(), end_time: values.end_time?.toISOString(),
      })
      message.success('创建成功'); setModalVisible(false); form.resetFields(); fetchPromoCodes()
    } catch { /* */ }
  }

  const handleDelete = async (id: number) => { try { await deletePromoCode(id); message.success('删除成功'); fetchPromoCodes() } catch { /* */ } }

  return (
    <div className="page-enter">
      <div className="page-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h1>促销码</h1>
            <div className="subtitle">管理促销码</div>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setModalVisible(true) }}>创建促销码</Button>
        </div>
      </div>

      {loading ? (
        <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>加载中...</div>
      ) : promoCodes.length === 0 ? (
        <Empty description="暂无促销码" />
      ) : (
        <div style={{ border: '1px solid var(--color-border)' }}>
          <div style={{
            display: 'grid',
            gridTemplateColumns: '60px 120px 1fr 100px 80px 80px 80px 80px 80px',
            gap: 16,
            padding: '12px 20px',
            borderBottom: '1px solid var(--color-border)',
            fontSize: 12,
            color: 'var(--color-text-tertiary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontFamily: 'var(--font-mono)',
          }}>
            <div>ID</div><div>促销码</div><div>活动</div><div>类型</div><div>优惠值</div><div>最低消费</div><div>使用</div><div>状态</div><div>操作</div>
          </div>
          {promoCodes.map(p => (
            <div key={p.id} style={{
              display: 'grid',
              gridTemplateColumns: '60px 120px 1fr 100px 80px 80px 80px 80px 80px',
              gap: 16,
              padding: '12px 20px',
              borderBottom: '1px solid var(--color-border)',
              alignItems: 'center',
              fontSize: 14,
            }}>
              <div style={{ fontFamily: 'var(--font-mono)', color: 'var(--color-text-tertiary)' }}>{p.id}</div>
              <div><Tag>{p.code}</Tag></div>
              <div style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{events.find(e => e.id === p.event_id)?.title || `活动#${p.event_id}`}</div>
              <div><Tag color={p.discount_type === 'percent' ? 'blue' : 'green'}>{p.discount_type === 'percent' ? '百分比' : '固定'}</Tag></div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{p.discount_type === 'percent' ? `${p.discount_value}%` : `¥${p.discount_value}`}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>¥{p.min_amount}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{p.used_count}/{p.max_uses}</div>
              <div><Tag color={p.is_active ? 'success' : 'default'}>{p.is_active ? '启用' : '禁用'}</Tag></div>
              <div><Popconfirm title="确定删除？" onConfirm={() => handleDelete(p.id)}><Button size="small" danger icon={<DeleteOutlined />}>删除</Button></Popconfirm></div>
            </div>
          ))}
        </div>
      )}

      <Modal title="创建促销码" open={modalVisible} onOk={handleCreate} onCancel={() => setModalVisible(false)} width={600}>
        <Form form={form} layout="vertical" initialValues={{ discount_type: 'percent', min_amount: 0, max_uses: 100 }}>
          <Form.Item name="code" label="促销码" rules={[{ required: true }]}><Input placeholder="如：SUMMER2024" /></Form.Item>
          <Form.Item name="event_id" label="关联活动" rules={[{ required: true }]}><Select placeholder="选择活动" allowClear options={events.map(e => ({ value: e.id, label: e.title }))} /></Form.Item>
          <Form.Item name="discount_type" label="类型" rules={[{ required: true }]}><Select options={[{ value: 'percent', label: '百分比' }, { value: 'fixed', label: '固定金额' }]} /></Form.Item>
          <Form.Item name="discount_value" label="优惠值" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="min_amount" label="最低消费"><InputNumber min={0} prefix="¥" style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="max_uses" label="最大使用次数"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="start_time" label="生效时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="end_time" label="失效时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
