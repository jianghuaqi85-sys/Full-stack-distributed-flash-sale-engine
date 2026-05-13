import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Button, Tag, Modal, Form, Input, InputNumber, DatePicker, Select, message, Empty, Popconfirm } from 'antd'
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
    // 获取所有活动的促销码（逐个活动获取）
    getEvents(1, 100)
      .then(async (res) => {
        const allEvents = res.data.data || []
        setEvents(allEvents)
        const allPromos: PromoCode[] = []
        for (const evt of allEvents) {
          try {
            const promoRes = await getPromoCodes(evt.id)
            if (promoRes.data.data) {
              allPromos.push(...promoRes.data.data)
            }
          } catch {
            // ignore
          }
        }
        setPromoCodes(allPromos)
      })
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    fetchPromoCodes()
  }, [fetchPromoCodes])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()
      await createPromoCode({
        code: values.code,
        event_id: values.event_id,
        discount_type: values.discount_type,
        discount_value: values.discount_value,
        min_amount: values.min_amount || 0,
        max_uses: values.max_uses || 100,
        start_time: values.start_time?.toISOString(),
        end_time: values.end_time?.toISOString(),
      })
      message.success('创建成功')
      setModalVisible(false)
      form.resetFields()
      fetchPromoCodes()
    } catch {
      // validation error
    }
  }

  const handleDelete = async (id: number) => {
    try {
      await deletePromoCode(id)
      message.success('删除成功')
      fetchPromoCodes()
    } catch {
      // handled by interceptor
    }
  }

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '促销码',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => <Tag color="purple">{code}</Tag>,
    },
    {
      title: '活动',
      dataIndex: 'event_id',
      key: 'event_id',
      render: (eventId: number) => {
        const evt = events.find((e) => e.id === eventId)
        return evt?.title || `活动#${eventId}`
      },
    },
    {
      title: '优惠类型',
      dataIndex: 'discount_type',
      key: 'discount_type',
      render: (type: string) => (
        <Tag color={type === 'percent' ? 'blue' : 'green'}>
          {type === 'percent' ? '百分比' : '固定金额'}
        </Tag>
      ),
    },
    {
      title: '优惠值',
      dataIndex: 'discount_value',
      key: 'discount_value',
      render: (value: number, record: PromoCode) => (
        record.discount_type === 'percent' ? `${value}%` : `¥${value}`
      ),
    },
    {
      title: '最低消费',
      dataIndex: 'min_amount',
      key: 'min_amount',
      render: (amount: number) => `¥${amount}`,
    },
    {
      title: '使用情况',
      key: 'usage',
      render: (_: unknown, record: PromoCode) => `${record.used_count} / ${record.max_uses}`,
    },
    {
      title: '状态',
      dataIndex: 'is_active',
      key: 'is_active',
      render: (active: boolean) => (
        <Tag color={active ? 'success' : 'default'}>{active ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: PromoCode) => (
        <Popconfirm title="确定删除此促销码？" onConfirm={() => handleDelete(record.id)}>
          <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
        </Popconfirm>
      ),
    },
  ]

  return (
    <Card
      title="促销码管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setModalVisible(true) }}>
          创建促销码
        </Button>
      }
      style={{ borderRadius: 16 }}
    >
      <Table
        columns={columns}
        dataSource={promoCodes}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
        locale={{ emptyText: <Empty description="暂无促销码" /> }}
      />

      <Modal
        title="创建促销码"
        open={modalVisible}
        onOk={handleCreate}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical" initialValues={{ discount_type: 'percent', min_amount: 0, max_uses: 100 }}>
          <Form.Item name="code" label="促销码" rules={[{ required: true, message: '请输入促销码' }]}>
            <Input placeholder="如：SUMMER2024" style={{ textTransform: 'uppercase' }} />
          </Form.Item>
          <Form.Item name="event_id" label="关联活动" rules={[{ required: true, message: '请选择活动' }]}>
            <Select
              placeholder="选择活动（留空则全局有效）"
              allowClear
              options={events.map((e) => ({ value: e.id, label: e.title }))}
            />
          </Form.Item>
          <Form.Item name="discount_type" label="优惠类型" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'percent', label: '百分比折扣' },
                { value: 'fixed', label: '固定金额' },
              ]}
            />
          </Form.Item>
          <Form.Item name="discount_value" label="优惠值" rules={[{ required: true, message: '请输入优惠值' }]}>
            <InputNumber min={0} style={{ width: '100%' }} placeholder="百分比输入如 20，固定金额输入如 50" />
          </Form.Item>
          <Form.Item name="min_amount" label="最低消费">
            <InputNumber min={0} prefix="¥" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="max_uses" label="最大使用次数">
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="start_time" label="生效时间">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="end_time" label="失效时间">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
