import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, InputNumber, DatePicker, message, Drawer, List, Empty, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, CalendarOutlined, DeleteOutlined, StopOutlined } from '@ant-design/icons'
import { getEvents, createEvent, updateEvent, publishEvent, unpublishEvent, endEvent, createTicketType, updateTicketType, deleteTicketType, Event } from '../../../api/events'
import { getEventShows, createShow, updateShow, deleteShow, publishShow, unpublishShow, Show } from '../../../api/shows'
import dayjs from 'dayjs'

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: 'default', text: '草稿' },
  on_sale: { color: 'success', text: '售票中' },
  off_sale: { color: 'warning', text: '已下架' },
  ended: { color: 'error', text: '已结束' },
}

export default function AdminEvents() {
  const [events, setEvents] = useState<Event[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingEvent, setEditingEvent] = useState<Event | null>(null)
  const [form] = Form.useForm()

  // 场次管理
  const [showDrawerVisible, setShowDrawerVisible] = useState(false)
  const [currentEvent, setCurrentEvent] = useState<Event | null>(null)
  const [shows, setShows] = useState<Show[]>([])
  const [showLoading, setShowLoading] = useState(false)
  const [showModalVisible, setShowModalVisible] = useState(false)
  const [editingShow, setEditingShow] = useState<Show | null>(null)
  const [showForm] = Form.useForm()

  // 票种管理
  const [ticketTypeDrawerVisible, setTicketTypeDrawerVisible] = useState(false)
  const [ticketTypes, setTicketTypes] = useState<Event['ticket_types']>([])
  const [ticketTypeModalVisible, setTicketTypeModalVisible] = useState(false)
  const [editingTicketType, setEditingTicketType] = useState<{ id: number; name: string; price: number; stock: number; max_per_user: number; sort_order: number } | null>(null)
  const [ticketTypeForm] = Form.useForm()

  const fetchEvents = useCallback(() => {
    setLoading(true)
    getEvents(page, 10)
      .then((res) => {
        setEvents(res.data.data)
        setTotal(res.data.total)
      })
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => {
    fetchEvents()
  }, [fetchEvents])

  const handleCreate = () => {
    setEditingEvent(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (event: Event) => {
    setEditingEvent(event)
    form.setFieldsValue({
      ...event,
      start_time: dayjs(event.start_time),
      end_time: dayjs(event.end_time),
    })
    setModalVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data = {
        ...values,
        start_time: values.start_time.toISOString(),
        end_time: values.end_time.toISOString(),
      }

      if (editingEvent) {
        await updateEvent(editingEvent.id, data)
        message.success('更新成功')
      } else {
        await createEvent(data)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchEvents()
    } catch {
      // validation error
    }
  }

  const handlePublish = async (id: number) => {
    try {
      await publishEvent(id)
      message.success('发布成功')
      fetchEvents()
    } catch {
      // handled by interceptor
    }
  }

  const handleUnpublish = async (id: number) => {
    try {
      await unpublishEvent(id)
      message.success('下架成功')
      fetchEvents()
    } catch {
      // handled by interceptor
    }
  }

  const handleEndEvent = async (id: number) => {
    try {
      await endEvent(id)
      message.success('已结束')
      fetchEvents()
    } catch {
      // handled by interceptor
    }
  }

  // 场次管理
  const handleManageShows = (event: Event) => {
    setCurrentEvent(event)
    setShowDrawerVisible(true)
    fetchShows(event.id)
  }

  const fetchShows = (eventId: number) => {
    setShowLoading(true)
    getEventShows(eventId)
      .then((res) => setShows(res.data.data || []))
      .finally(() => setShowLoading(false))
  }

  const handleCreateShow = () => {
    setEditingShow(null)
    showForm.resetFields()
    setShowModalVisible(true)
  }

  const handleEditShow = (show: Show) => {
    setEditingShow(show)
    showForm.setFieldsValue({
      ...show,
      show_time: dayjs(show.show_time),
      end_time: dayjs(show.end_time),
    })
    setShowModalVisible(true)
  }

  const handleShowSubmit = async () => {
    if (!currentEvent) return
    try {
      const values = await showForm.validateFields()
      const data = {
        ...values,
        show_time: values.show_time.toISOString(),
        end_time: values.end_time.toISOString(),
      }

      if (editingShow) {
        await updateShow(editingShow.id, data)
        message.success('更新场次成功')
      } else {
        await createShow(currentEvent.id, data)
        message.success('创建场次成功')
      }
      setShowModalVisible(false)
      fetchShows(currentEvent.id)
    } catch {
      // validation error
    }
  }

  const handlePublishShow = async (showId: number) => {
    try {
      await publishShow(showId)
      message.success('上架成功')
      if (currentEvent) fetchShows(currentEvent.id)
    } catch {
      // handled
    }
  }

  const handleUnpublishShow = async (showId: number) => {
    try {
      await unpublishShow(showId)
      message.success('下架成功')
      if (currentEvent) fetchShows(currentEvent.id)
    } catch {
      // handled
    }
  }

  const handleDeleteShow = async (showId: number) => {
    try {
      await deleteShow(showId)
      message.success('删除成功')
      if (currentEvent) fetchShows(currentEvent.id)
    } catch {
      // handled
    }
  }

  // 票种管理
  const handleManageTicketTypes = (event: Event) => {
    setCurrentEvent(event)
    setTicketTypeDrawerVisible(true)
    setTicketTypes(event.ticket_types || [])
  }

  const handleCreateTicketType = () => {
    setEditingTicketType(null)
    ticketTypeForm.resetFields()
    ticketTypeForm.setFieldsValue({ max_per_user: 4, sort_order: 0 })
    setTicketTypeModalVisible(true)
  }

  const handleEditTicketType = (tt: { id: number; name: string; price: number; stock: number; max_per_user: number; sort_order: number }) => {
    setEditingTicketType(tt)
    ticketTypeForm.setFieldsValue(tt)
    setTicketTypeModalVisible(true)
  }

  const handleTicketTypeSubmit = async () => {
    if (!currentEvent) return
    try {
      const values = await ticketTypeForm.validateFields()

      if (editingTicketType) {
        await updateTicketType(editingTicketType.id, values)
        message.success('更新票种成功')
      } else {
        await createTicketType(currentEvent.id, values)
        message.success('创建票种成功')
      }
      setTicketTypeModalVisible(false)
      // 刷新活动列表以获取最新票种
      fetchEvents()
      // 更新当前活动的票种
      if (editingTicketType) {
        setTicketTypes((prev) =>
          prev?.map((tt) => (tt?.id === editingTicketType.id ? { ...tt, ...values } : tt)) || []
        )
      } else {
        // 简单刷新：重新获取活动详情
        getEvents(page, 10).then((res) => {
          const updated = res.data.data.find((e) => e.id === currentEvent.id)
          if (updated) {
            setTicketTypes(updated.ticket_types || [])
            setCurrentEvent(updated)
          }
        })
      }
    } catch {
      // validation error
    }
  }

  const handleDeleteTicketType = async (ttId: number) => {
    try {
      await deleteTicketType(ttId)
      message.success('删除票种成功')
      setTicketTypes((prev) => prev?.filter((tt) => tt?.id !== ttId) || [])
      fetchEvents()
    } catch {
      // handled
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
      title: '活动名称',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: '地点',
      dataIndex: 'location',
      key: 'location',
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
    },
    {
      title: '结束时间',
      dataIndex: 'end_time',
      key: 'end_time',
      render: (time: string) => new Date(time).toLocaleString('zh-CN'),
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
      title: '库存',
      dataIndex: 'total_stock',
      key: 'total_stock',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: Event) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button size="small" onClick={() => handleManageTicketTypes(record)}>
            票种
          </Button>
          <Button size="small" icon={<CalendarOutlined />} onClick={() => handleManageShows(record)}>
            场次
          </Button>
          {record.status === 'draft' && (
            <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => handlePublish(record.id)}>
              发布
            </Button>
          )}
          {record.status === 'on_sale' && (
            <>
              <Button size="small" danger onClick={() => handleUnpublish(record.id)}>
                下架
              </Button>
              <Popconfirm title="确定结束此活动？结束后不可恢复。" onConfirm={() => handleEndEvent(record.id)}>
                <Button size="small" danger icon={<StopOutlined />}>
                  结束
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card
        title="活动管理"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            创建活动
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={events}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            total,
            pageSize: 10,
            onChange: setPage,
            showTotal: (t) => `共 ${t} 个活动`,
          }}
        />
      </Card>

      {/* 活动编辑弹窗 */}
      <Modal
        title={editingEvent ? '编辑活动' : '创建活动'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={600}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="活动名称" rules={[{ required: true, message: '请输入活动名称' }]}>
            <Input placeholder="请输入活动名称" />
          </Form.Item>
          <Form.Item name="description" label="活动描述">
            <Input.TextArea rows={3} placeholder="请输入活动描述" />
          </Form.Item>
          <Form.Item name="location" label="活动地点" rules={[{ required: true, message: '请输入活动地点' }]}>
            <Input placeholder="请输入活动地点" />
          </Form.Item>
          <Form.Item name="cover_image" label="封面图片URL">
            <Input placeholder="请输入封面图片URL" />
          </Form.Item>
          <Form.Item name="start_time" label="开始时间" rules={[{ required: true, message: '请选择开始时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="end_time" label="结束时间" rules={[{ required: true, message: '请选择结束时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 场次管理抽屉 */}
      <Drawer
        title={currentEvent ? `${currentEvent.title} - 场次管理` : '场次管理'}
        open={showDrawerVisible}
        onClose={() => setShowDrawerVisible(false)}
        width={600}
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateShow}>
            添加场次
          </Button>
        }
      >
        <List
          loading={showLoading}
          dataSource={shows}
          locale={{ emptyText: <Empty description="暂无场次" /> }}
          renderItem={(show) => (
            <List.Item
              actions={[
                show.status === 'draft' && (
                  <Button size="small" type="primary" onClick={() => handlePublishShow(show.id)}>上架</Button>
                ),
                show.status === 'on_sale' && (
                  <Button size="small" danger onClick={() => handleUnpublishShow(show.id)}>下架</Button>
                ),
                <Button size="small" icon={<EditOutlined />} onClick={() => handleEditShow(show)}>编辑</Button>,
                show.status === 'draft' && (
                  <Popconfirm title="确定删除？" onConfirm={() => handleDeleteShow(show.id)}>
                    <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                  </Popconfirm>
                ),
              ].filter(Boolean)}
            >
              <List.Item.Meta
                title={
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span>{show.name}</span>
                    <Tag color={statusMap[show.status]?.color}>{statusMap[show.status]?.text}</Tag>
                  </div>
                }
                description={
                  <div>
                    <div>开始: {new Date(show.show_time).toLocaleString('zh-CN')}</div>
                    <div>结束: {new Date(show.end_time).toLocaleString('zh-CN')}</div>
                    <div>库存: {show.stock} 张 | 已售: {show.sold_count} 张</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Drawer>

      {/* 场次编辑弹窗 */}
      <Modal
        title={editingShow ? '编辑场次' : '添加场次'}
        open={showModalVisible}
        onOk={handleShowSubmit}
        onCancel={() => setShowModalVisible(false)}
        width={500}
      >
        <Form form={showForm} layout="vertical">
          <Form.Item name="name" label="场次名称" rules={[{ required: true, message: '请输入场次名称' }]}>
            <Input placeholder="如：第一场、下午场" />
          </Form.Item>
          <Form.Item name="show_time" label="开始时间" rules={[{ required: true, message: '请选择开始时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="end_time" label="结束时间" rules={[{ required: true, message: '请选择结束时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="stock" label="库存" rules={[{ required: true, message: '请输入库存' }]}>
            <InputNumber min={0} style={{ width: '100%' }} placeholder="请输入库存数量" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber placeholder="数字越小越靠前" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* 票种管理抽屉 */}
      <Drawer
        title={currentEvent ? `${currentEvent.title} - 票种管理` : '票种管理'}
        open={ticketTypeDrawerVisible}
        onClose={() => setTicketTypeDrawerVisible(false)}
        width={600}
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreateTicketType}>
            添加票种
          </Button>
        }
      >
        <List
          dataSource={ticketTypes || []}
          locale={{ emptyText: <Empty description="暂无票种" /> }}
          renderItem={(tt) => tt && (
            <List.Item
              actions={[
                <Button size="small" icon={<EditOutlined />} onClick={() => handleEditTicketType(tt)}>编辑</Button>,
                <Popconfirm title="确定删除此票种？" onConfirm={() => handleDeleteTicketType(tt.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
                </Popconfirm>,
              ]}
            >
              <List.Item.Meta
                title={
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <span style={{ fontWeight: 600 }}>{tt.name}</span>
                    <span style={{ color: 'var(--color-gold)', fontWeight: 600 }}>¥{tt.price}</span>
                  </div>
                }
                description={
                  <div>
                    <div>库存: {tt.stock} 张 | 每人限购: {tt.max_per_user} 张 | 排序: {tt.sort_order}</div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Drawer>

      {/* 票种编辑弹窗 */}
      <Modal
        title={editingTicketType ? '编辑票种' : '添加票种'}
        open={ticketTypeModalVisible}
        onOk={handleTicketTypeSubmit}
        onCancel={() => setTicketTypeModalVisible(false)}
        width={500}
      >
        <Form form={ticketTypeForm} layout="vertical">
          <Form.Item name="name" label="票种名称" rules={[{ required: true, message: '请输入票种名称' }]}>
            <Input placeholder="如：VIP票、普通票" />
          </Form.Item>
          <Form.Item name="price" label="价格" rules={[{ required: true, message: '请输入价格' }]}>
            <InputNumber min={0} step={0.01} prefix="¥" style={{ width: '100%' }} placeholder="请输入价格" />
          </Form.Item>
          <Form.Item name="stock" label="库存" rules={[{ required: true, message: '请输入库存' }]}>
            <InputNumber min={0} style={{ width: '100%' }} placeholder="请输入库存数量" />
          </Form.Item>
          <Form.Item name="max_per_user" label="每人限购">
            <InputNumber min={1} max={100} style={{ width: '100%' }} placeholder="默认4张" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber placeholder="数字越小越靠前" style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
