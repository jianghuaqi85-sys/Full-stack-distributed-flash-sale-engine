import { useEffect, useState } from 'react'
import { Tag, Empty, Pagination, Input, Select } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getEvents, Event } from '../../api/events'

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
    getEvents(page, 20)
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
    <div className="page-enter">
      {/* Header */}
      <div className="page-header">
        <h1>活动</h1>
        <div className="subtitle">浏览所有票务活动</div>
      </div>

      {/* Filters */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 24,
        flexWrap: 'wrap',
        gap: 12,
      }}>
        <Input
          placeholder="搜索活动名称或地点"
          prefix={<SearchOutlined style={{ color: 'var(--color-text-tertiary)' }} />}
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          style={{ width: 280 }}
          allowClear
        />
        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          style={{ width: 120 }}
          options={[
            { value: 'all', label: '全部' },
            { value: 'on_sale', label: '售票中' },
            { value: 'off_sale', label: '已下架' },
            { value: 'ended', label: '已结束' },
          ]}
        />
      </div>

      {/* Content */}
      {loading ? (
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton" style={{ height: 72, marginBottom: 1 }} />
          ))}
        </div>
      ) : filteredEvents.length === 0 ? (
        <Empty description="暂无活动" />
      ) : (
        <>
          {/* Table-style list */}
          <div style={{ border: '1px solid var(--color-border)' }}>
            {/* Header */}
            <div style={{
              display: 'grid',
              gridTemplateColumns: '1fr 120px 160px 100px 100px',
              gap: 16,
              padding: '12px 20px',
              background: 'var(--color-bg-container)',
              borderBottom: '1px solid var(--color-border)',
              fontSize: 12,
              color: 'var(--color-text-tertiary)',
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontFamily: 'var(--font-mono)',
            }}>
              <div>活动</div>
              <div>日期</div>
              <div>地点</div>
              <div>库存</div>
              <div>状态</div>
            </div>

            {/* Rows */}
            {filteredEvents.map((event) => (
              <button
                key={event.id}
                onClick={() => navigate(`/events/${event.id}`)}
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 120px 160px 100px 100px',
                  gap: 16,
                  padding: '16px 20px',
                  border: 'none',
                  borderBottom: '1px solid var(--color-border)',
                  background: 'transparent',
                  cursor: 'pointer',
                  textAlign: 'left',
                  width: '100%',
                  transition: 'background 0.1s',
                  alignItems: 'center',
                }}
                onMouseEnter={e => e.currentTarget.style.background = 'var(--color-accent-soft)'}
                onMouseLeave={e => e.currentTarget.style.background = 'transparent'}
              >
                <div style={{ fontWeight: 600, fontSize: 15, color: 'var(--color-text-primary)' }}>
                  {event.title}
                </div>
                <div style={{ fontSize: 13, color: 'var(--color-text-secondary)', fontFamily: 'var(--font-mono)' }}>
                  {new Date(event.start_time).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })}
                </div>
                <div style={{ fontSize: 13, color: 'var(--color-text-secondary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {event.location}
                </div>
                <div style={{ fontSize: 13, fontFamily: 'var(--font-mono)', color: event.total_stock > 0 ? 'var(--color-text-primary)' : 'var(--color-text-tertiary)' }}>
                  {event.total_stock > 0 ? event.total_stock : '—'}
                </div>
                <div>
                  <Tag color={statusMap[event.status]?.color}>
                    {statusMap[event.status]?.text}
                  </Tag>
                </div>
              </button>
            ))}
          </div>

          {/* Pagination */}
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Pagination
              current={page}
              total={total}
              pageSize={20}
              onChange={setPage}
              showTotal={(t) => `共 ${t} 个活动`}
              size="small"
            />
          </div>
        </>
      )}
    </div>
  )
}
