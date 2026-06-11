import { useEffect, useState, useCallback } from 'react'
import { Button, Tag, Space, Modal, Form, Input, InputNumber, DatePicker, message, Drawer, List, Empty, Popconfirm } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
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
  const [showDrawerVisible, setShowDrawerVisible] = useState(false)
  const [currentEvent, setCurrentEvent] = useState<Event | null>(null)
  const [shows, setShows] = useState<Show[]>([])
  const [showLoading, setShowLoading] = useState(false)
  const [showModalVisible, setShowModalVisible] = useState(false)
  const [editingShow, setEditingShow] = useState<Show | null>(null)
  const [showForm] = Form.useForm()
  const [ticketTypeDrawerVisible, setTicketTypeDrawerVisible] = useState(false)
  const [ticketTypes, setTicketTypes] = useState<Event['ticket_types']>([])
  const [ticketTypeModalVisible, setTicketTypeModalVisible] = useState(false)
  const [editingTicketType, setEditingTicketType] = useState<{ id: number; name: string; price: number; stock: number; max_per_user: number; sort_order: number } | null>(null)
  const [ticketTypeForm] = Form.useForm()

  const fetchEvents = useCallback(() => {
    setLoading(true)
    getEvents(page, 10).then((res) => { setEvents(res.data.data || []); setTotal(res.data.total) }).finally(() => setLoading(false))
  }, [page])

  useEffect(() => { fetchEvents() }, [fetchEvents])

  const handleCreate = () => { setEditingEvent(null); form.resetFields(); setModalVisible(true) }
  const handleEdit = (event: Event) => {
    setEditingEvent(event)
    form.setFieldsValue({ ...event, start_time: dayjs(event.start_time), end_time: dayjs(event.end_time) })
    setModalVisible(true)
  }

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const data = { ...values, start_time: values.start_time.toISOString(), end_time: values.end_time.toISOString() }
      if (editingEvent) { await updateEvent(editingEvent.id, data); message.success('更新成功') }
      else { await createEvent(data); message.success('创建成功') }
      setModalVisible(false); fetchEvents()
    } catch { /* validation */ }
  }

  const handlePublish = async (id: number) => { try { await publishEvent(id); message.success('发布成功'); fetchEvents() } catch { /* */ } }
  const handleUnpublish = async (id: number) => { try { await unpublishEvent(id); message.success('下架成功'); fetchEvents() } catch { /* */ } }
  const handleEndEvent = async (id: number) => { try { await endEvent(id); message.success('已结束'); fetchEvents() } catch { /* */ } }

  const handleManageShows = (event: Event) => { setCurrentEvent(event); setShowDrawerVisible(true); fetchShows(event.id) }
  const fetchShows = (eventId: number) => { setShowLoading(true); getEventShows(eventId).then((res) => setShows(res.data.data || [])).finally(() => setShowLoading(false)) }
  const handleCreateShow = () => { setEditingShow(null); showForm.resetFields(); setShowModalVisible(true) }
  const handleEditShow = (show: Show) => {
    setEditingShow(show)
    showForm.setFieldsValue({ ...show, show_time: dayjs(show.show_time), end_time: dayjs(show.end_time) })
    setShowModalVisible(true)
  }
  const handleShowSubmit = async () => {
    if (!currentEvent) return
    try {
      const values = await showForm.validateFields()
      const data = { ...values, show_time: values.show_time.toISOString(), end_time: values.end_time.toISOString() }
      if (editingShow) { await updateShow(editingShow.id, data); message.success('更新场次成功') }
      else { await createShow(currentEvent.id, data); message.success('创建场次成功') }
      setShowModalVisible(false); fetchShows(currentEvent.id)
    } catch { /* */ }
  }
  const handlePublishShow = async (showId: number) => { try { await publishShow(showId); message.success('上架成功'); if (currentEvent) fetchShows(currentEvent.id) } catch { /* */ } }
  const handleUnpublishShow = async (showId: number) => { try { await unpublishShow(showId); message.success('下架成功'); if (currentEvent) fetchShows(currentEvent.id) } catch { /* */ } }
  const handleDeleteShow = async (showId: number) => { try { await deleteShow(showId); message.success('删除成功'); if (currentEvent) fetchShows(currentEvent.id) } catch { /* */ } }

  const handleManageTicketTypes = (event: Event) => { setCurrentEvent(event); setTicketTypeDrawerVisible(true); setTicketTypes(event.ticket_types || []) }
  const handleCreateTicketType = () => { setEditingTicketType(null); ticketTypeForm.resetFields(); ticketTypeForm.setFieldsValue({ max_per_user: 4, sort_order: 0 }); setTicketTypeModalVisible(true) }
  const handleEditTicketType = (tt: { id: number; name: string; price: number; stock: number; max_per_user: number; sort_order: number }) => { setEditingTicketType(tt); ticketTypeForm.setFieldsValue(tt); setTicketTypeModalVisible(true) }
  const handleTicketTypeSubmit = async () => {
    if (!currentEvent) return
    try {
      const values = await ticketTypeForm.validateFields()
      if (editingTicketType) { await updateTicketType(editingTicketType.id, values); message.success('更新票种成功') }
      else { await createTicketType(currentEvent.id, values); message.success('创建票种成功') }
      setTicketTypeModalVisible(false); fetchEvents()
      if (editingTicketType) { setTicketTypes((prev) => prev?.map((tt) => (tt?.id === editingTicketType.id ? { ...tt, ...values } : tt)) || []) }
      else {
        getEvents(page, 10).then((res) => {
          const updated = res.data.data.find((e) => e.id === currentEvent.id)
          if (updated) { setTicketTypes(updated.ticket_types || []); setCurrentEvent(updated) }
        })
      }
    } catch { /* */ }
  }
  const handleDeleteTicketType = async (ttId: number) => { try { await deleteTicketType(ttId); message.success('删除票种成功'); setTicketTypes((prev) => prev?.filter((tt) => tt?.id !== ttId) || []); fetchEvents() } catch { /* */ } }

  return (
    <div className="page-enter">
      <div className="page-header">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h1>活动管理</h1>
            <div className="subtitle">创建和管理票务活动</div>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>创建活动</Button>
        </div>
      </div>

      {/* Table */}
      <div style={{ border: '1px solid var(--color-border)' }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: '60px 1fr 120px 140px 80px 60px 200px',
          gap: 16,
          padding: '12px 20px',
          borderBottom: '1px solid var(--color-border)',
          fontSize: 12,
          color: 'var(--color-text-tertiary)',
          textTransform: 'uppercase',
          letterSpacing: '0.05em',
          fontFamily: 'var(--font-mono)',
        }}>
          <div>ID</div><div>活动名称</div><div>地点</div><div>开始时间</div><div>库存</div><div>状态</div><div>操作</div>
        </div>
        {loading ? (
          <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>加载中...</div>
        ) : events.length === 0 ? (
          <div style={{ padding: 40, textAlign: 'center', color: 'var(--color-text-tertiary)' }}>暂无活动</div>
        ) : (
          events.map(event => (
            <div key={event.id} style={{
              display: 'grid',
              gridTemplateColumns: '60px 1fr 120px 140px 80px 60px 200px',
              gap: 16,
              padding: '12px 20px',
              borderBottom: '1px solid var(--color-border)',
              alignItems: 'center',
              fontSize: 14,
            }}>
              <div style={{ fontFamily: 'var(--font-mono)', color: 'var(--color-text-tertiary)' }}>{event.id}</div>
              <div style={{ fontWeight: 600 }}>{event.title}</div>
              <div style={{ color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{event.location}</div>
              <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)' }}>{new Date(event.start_time).toLocaleDateString('zh-CN')}</div>
              <div style={{ fontFamily: 'var(--font-mono)' }}>{event.total_stock}</div>
              <div><Tag color={statusMap[event.status]?.color}>{statusMap[event.status]?.text}</Tag></div>
              <div>
                <Space size={4}>
                  <Button size="small" onClick={() => handleEdit(event)}>编辑</Button>
                  <Button size="small" onClick={() => handleManageTicketTypes(event)}>票种</Button>
                  <Button size="small" onClick={() => handleManageShows(event)}>场次</Button>
                  {event.status === 'draft' && <Button size="small" type="primary" onClick={() => handlePublish(event.id)}>发布</Button>}
                  {event.status === 'on_sale' && (
                    <>
                      <Button size="small" danger onClick={() => handleUnpublish(event.id)}>下架</Button>
                      <Popconfirm title="确定结束？" onConfirm={() => handleEndEvent(event.id)}>
                        <Button size="small" danger>结束</Button>
                      </Popconfirm>
                    </>
                  )}
                </Space>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Pagination */}
      <div style={{ textAlign: 'center', marginTop: 16, fontSize: 13, color: 'var(--color-text-tertiary)' }}>
        共 {total} 个活动 — 第 {page} 页
        {total > 10 && (
          <Space style={{ marginLeft: 16 }}>
            <Button size="small" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>上一页</Button>
            <Button size="small" disabled={page * 10 >= total} onClick={() => setPage(p => p + 1)}>下一页</Button>
          </Space>
        )}
      </div>

      {/* Event modal */}
      <Modal title={editingEvent ? '编辑活动' : '创建活动'} open={modalVisible} onOk={handleSubmit} onCancel={() => setModalVisible(false)} width={600}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="活动名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="location" label="地点" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="cover_image" label="封面URL"><Input /></Form.Item>
          <Form.Item name="start_time" label="开始时间" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="end_time" label="结束时间" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>

      {/* Shows drawer */}
      <Drawer title={currentEvent ? `${currentEvent.title} — 场次` : '场次'} open={showDrawerVisible} onClose={() => setShowDrawerVisible(false)} width={600} extra={<Button type="primary" icon={<PlusOutlined />} onClick={handleCreateShow}>添加</Button>}>
        <List loading={showLoading} dataSource={shows} locale={{ emptyText: <Empty description="暂无场次" /> }} renderItem={(show) => (
          <List.Item actions={[
            show.status === 'draft' && <Button size="small" type="primary" onClick={() => handlePublishShow(show.id)}>上架</Button>,
            show.status === 'on_sale' && <Button size="small" danger onClick={() => handleUnpublishShow(show.id)}>下架</Button>,
            <Button size="small" onClick={() => handleEditShow(show)}>编辑</Button>,
            show.status === 'draft' && <Popconfirm title="确定？" onConfirm={() => handleDeleteShow(show.id)}><Button size="small" danger>删除</Button></Popconfirm>,
          ].filter(Boolean)}>
            <List.Item.Meta title={<div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><span>{show.name}</span><Tag color={statusMap[show.status]?.color}>{statusMap[show.status]?.text}</Tag></div>} description={<div style={{ fontSize: 13, fontFamily: 'var(--font-mono)' }}>{new Date(show.show_time).toLocaleString('zh-CN')} — 库存 {show.stock} / 已售 {show.sold_count}</div>} />
          </List.Item>
        )} />
      </Drawer>

      <Modal title={editingShow ? '编辑场次' : '添加场次'} open={showModalVisible} onOk={handleShowSubmit} onCancel={() => setShowModalVisible(false)} width={500}>
        <Form form={showForm} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input placeholder="如：第一场" /></Form.Item>
          <Form.Item name="show_time" label="开始" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="end_time" label="结束" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="stock" label="库存" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="sort_order" label="排序"><InputNumber style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>

      {/* Ticket types drawer */}
      <Drawer title={currentEvent ? `${currentEvent.title} — 票种` : '票种'} open={ticketTypeDrawerVisible} onClose={() => setTicketTypeDrawerVisible(false)} width={600} extra={<Button type="primary" icon={<PlusOutlined />} onClick={handleCreateTicketType}>添加</Button>}>
        <List dataSource={ticketTypes || []} locale={{ emptyText: <Empty description="暂无票种" /> }} renderItem={(tt) => tt && (
          <List.Item actions={[
            <Button size="small" onClick={() => handleEditTicketType(tt)}>编辑</Button>,
            <Popconfirm title="确定？" onConfirm={() => handleDeleteTicketType(tt.id)}><Button size="small" danger>删除</Button></Popconfirm>,
          ]}>
            <List.Item.Meta title={<div style={{ display: 'flex', alignItems: 'center', gap: 12 }}><span style={{ fontWeight: 600 }}>{tt.name}</span><span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>¥{tt.price}</span></div>} description={<span style={{ fontSize: 13 }}>库存 {tt.stock} · 限购 {tt.max_per_user}</span>} />
          </List.Item>
        )} />
      </Drawer>

      <Modal title={editingTicketType ? '编辑票种' : '添加票种'} open={ticketTypeModalVisible} onOk={handleTicketTypeSubmit} onCancel={() => setTicketTypeModalVisible(false)} width={500}>
        <Form form={ticketTypeForm} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input placeholder="如：VIP票" /></Form.Item>
          <Form.Item name="price" label="价格" rules={[{ required: true }]}><InputNumber min={0} step={0.01} prefix="¥" style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="stock" label="库存" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="max_per_user" label="每人限购"><InputNumber min={1} max={100} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="sort_order" label="排序"><InputNumber style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
