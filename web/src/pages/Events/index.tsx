import { useEffect, useState } from 'react'
import { Card, Row, Col, Tag, Button, Empty, Pagination, Input, Select, Space } from 'antd'
import { CalendarOutlined, EnvironmentOutlined, SearchOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getEvents, Event } from '../../api/events'
import { getEventGradient } from '../../theme/gradients'
import SkeletonCard from '../../components/SkeletonCard'

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: 'default', text: '草稿' },
  on_sale: { color: 'success', text: '售票中' },
  off_sale: { color: 'warning', text: '已下架' },
  ended: { color: 'error', text: '已结束' },
}

export default function Events() {
  const [events, setEvents] = useState<Event[]>([])
  const [filteredEvents, setFilteredEvents] = useState<Event[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const navigate = useNavigate()

  useEffect(() => {
    setLoading(true)
    getEvents(page, 12)
      .then((res) => {
        setEvents(res.data.data)
        setTotal(res.data.total)
      })
      .finally(() => setLoading(false))
  }, [page])

  useEffect(() => {
    let result = events
    if (searchText) {
      result = result.filter(e =>
        e.title.toLowerCase().includes(searchText.toLowerCase()) ||
        e.location.toLowerCase().includes(searchText.toLowerCase())
      )
    }
    if (statusFilter !== 'all') {
      result = result.filter(e => e.status === statusFilter)
    }
    setFilteredEvents(result)
  }, [events, searchText, statusFilter])

  return (
    <div className="animate-slide-up">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
        <h2 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>活动列表</h2>
        <Space>
          <Input
            placeholder="搜索活动名称或地点"
            prefix={<SearchOutlined style={{ color: 'var(--color-text-tertiary)' }} />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250, borderRadius: 24 }}
            allowClear
          />
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'on_sale', label: '售票中' },
              { value: 'off_sale', label: '已下架' },
              { value: 'ended', label: '已结束' },
            ]}
          />
        </Space>
      </div>

      {loading ? (
        <SkeletonCard variant="card" count={8} />
      ) : filteredEvents.length === 0 ? (
        <Empty description="暂无活动" />
      ) : (
        <>
          <Row gutter={[16, 16]}>
            {filteredEvents.map((event) => (
              <Col key={event.id} xs={24} sm={12} lg={8} xl={6}>
                <Card
                  hoverable
                  onClick={() => navigate(`/events/${event.id}`)}
                  style={{ borderRadius: 16, overflow: 'hidden', cursor: 'pointer' }}
                  bodyStyle={{ padding: 16 }}
                  cover={
                    <div style={{
                      height: 160,
                      background: getEventGradient(event.id),
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      position: 'relative',
                      overflow: 'hidden',
                    }}>
                      <div style={{
                        position: 'absolute',
                        inset: 0,
                        background: 'radial-gradient(circle at 30% 50%, rgba(255,255,255,0.1) 0%, transparent 60%)',
                      }} />
                      <span style={{
                        color: '#fff',
                        fontSize: 28,
                        fontWeight: 700,
                        textShadow: '0 2px 8px rgba(0,0,0,0.2)',
                        position: 'relative',
                        zIndex: 1,
                      }}>
                        {event.title.slice(0, 4)}
                      </span>
                      <Tag
                        color={statusMap[event.status]?.color}
                        style={{
                          position: 'absolute',
                          top: 12,
                          right: 12,
                          borderRadius: 8,
                          fontWeight: 500,
                          zIndex: 1,
                        }}
                      >
                        {statusMap[event.status]?.text}
                      </Tag>
                    </div>
                  }
                >
                  <div style={{ fontWeight: 600, fontSize: 16, marginBottom: 8, lineHeight: 1.4 }}>
                    {event.title}
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--color-text-secondary)', fontSize: 13, marginBottom: 4 }}>
                    <CalendarOutlined />
                    {new Date(event.start_time).toLocaleDateString('zh-CN')}
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--color-text-secondary)', fontSize: 13, marginBottom: 12 }}>
                    <EnvironmentOutlined />
                    {event.location}
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <span style={{ color: 'var(--color-gold)', fontSize: 20, fontWeight: 700 }}>
                      {event.total_stock > 0 ? `${event.total_stock} 张` : '暂无票'}
                    </span>
                    <Button
                      type="primary"
                      size="small"
                      disabled={event.status !== 'on_sale'}
                      style={{ borderRadius: 8 }}
                    >
                      {event.status === 'on_sale' ? '立即抢票' : '查看'}
                    </Button>
                  </div>
                </Card>
              </Col>
            ))}
          </Row>
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Pagination
              current={page}
              total={total}
              pageSize={12}
              onChange={setPage}
              showTotal={(t) => `共 ${t} 个活动`}
            />
          </div>
        </>
      )}
    </div>
  )
}
