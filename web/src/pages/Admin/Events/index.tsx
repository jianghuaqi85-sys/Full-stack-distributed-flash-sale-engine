import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, Input, DatePicker, message, Drawer, List, Empty, Popconfirm } from 'antd'
import { PlusOutlined, EditOutlined, CalendarOutlined, DeleteOutlined } from '@ant-design/icons'
import { getEvents, createEvent, updateEvent, publishEvent, unpublishEvent, Event } from '../../../api/events'
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
          <Button size="small" icon={<CalendarOutlined />} onClick={() => handleManageShows(record)}>
            场次
          </Button>
          {record.status === 'draft' && (
            <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => handlePublish(record.id)}>
              发布
            </Button>
          )}
          {record.status === 'on_sale' && (
            <Button size="small" danger onClick={() => handleUnpublish(record.id)}>
              下架
            </Button>
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
            <Input type="number" min={0} placeholder="请输入库存数量" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <Input type="number" placeholder="数字越小越靠前" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
