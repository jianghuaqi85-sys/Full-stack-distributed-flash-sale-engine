import { useEffect, useState } from 'react'
import { Select, Spin } from 'antd'
import { Line, Pie } from '@ant-design/charts'
import { getDashboardStats, getSalesTrend, getTicketTypeStats, getConversionFunnel, DashboardStats, SalesTrend, TicketTypeStats, ConversionFunnel } from '../../../api/stats'
import { getEvents, Event } from '../../../api/events'

export default function AdminDashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [salesTrend, setSalesTrend] = useState<SalesTrend[]>([])
  const [ticketTypeStats, setTicketTypeStats] = useState<TicketTypeStats[]>([])
  const [funnel, setFunnel] = useState<ConversionFunnel | null>(null)
  const [loading, setLoading] = useState(true)
  const [trendDays, setTrendDays] = useState(7)
  const [funnelEventId, setFunnelEventId] = useState<number | undefined>(undefined)
  const [events, setEvents] = useState<Event[]>([])

  useEffect(() => { fetchData() }, [trendDays])

  useEffect(() => {
    if (funnelEventId) {
      getConversionFunnel(funnelEventId).then((res) => setFunnel(res.data)).catch(() => setFunnel(null))
    }
  }, [funnelEventId])

  const fetchData = async () => {
    setLoading(true)
    try {
      const [statsRes, trendRes, ticketRes, eventsRes] = await Promise.all([
        getDashboardStats(), getSalesTrend(trendDays), getTicketTypeStats(), getEvents(1, 100),
      ])
      setStats(statsRes.data)
      setSalesTrend(trendRes.data.data || [])
      setTicketTypeStats(ticketRes.data.data || [])
      setEvents(eventsRes.data.data || [])
    } catch { /* ignore */ } finally { setLoading(false) }
  }

  const salesTrendData = salesTrend.map(item => ({ date: item.date, 销量: item.count }))
  const ticketTypeData = ticketTypeStats.map(item => ({ type: item.ticket_type_name, value: item.sold_count }))

  const funnelData = funnel ? [
    { stage: '页面浏览', count: funnel.page_views },
    { stage: '加购', count: funnel.add_to_cart },
    { stage: '预留', count: funnel.reserved },
    { stage: '支付', count: funnel.paid },
    { stage: '使用', count: funnel.used },
  ] : []

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  const statItems = [
    { label: '活动总数', value: stats?.total_events || 0 },
    { label: '售票中', value: stats?.active_events || 0, accent: true },
    { label: '已售票数', value: stats?.sold_tickets || 0, suffix: `/ ${stats?.total_tickets || 0}` },
    { label: '总收入', value: `¥${(stats?.total_revenue || 0).toFixed(2)}` },
    { label: '今日销量', value: stats?.today_sales || 0, accent: true },
    { label: '今日收入', value: `¥${(stats?.today_revenue || 0).toFixed(2)}` },
    { label: '待支付', value: stats?.reserved_tickets || 0 },
    { label: '转化率', value: `${stats?.total_tickets ? ((stats?.sold_tickets || 0) / stats.total_tickets * 100).toFixed(1) : 0}%`, accent: true },
  ]

  return (
    <div className="page-enter">
      <div className="page-header">
        <h1>数据</h1>
        <div className="subtitle">管理后台数据仪表盘</div>
      </div>

      {/* Stats grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
        gap: 1,
        background: 'var(--color-border)',
        border: '1px solid var(--color-border)',
        marginBottom: 32,
      }}>
        {statItems.map(item => (
          <div key={item.label} style={{ padding: '24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 8 }}>
              {item.label}
            </div>
            <div style={{
              fontSize: 32,
              fontWeight: 700,
              fontFamily: 'var(--font-mono)',
              letterSpacing: '-0.04em',
              color: item.accent ? 'var(--color-accent)' : 'var(--color-text-primary)',
            }}>
              {item.value}
              {item.suffix && <span style={{ fontSize: 16, color: 'var(--color-text-tertiary)', fontWeight: 400 }}> {item.suffix}</span>}
            </div>
          </div>
        ))}
      </div>

      {/* Charts */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 1, background: 'var(--color-border)', border: '1px solid var(--color-border)', marginBottom: 32 }}>
        <div style={{ padding: '24px', background: 'var(--color-bg-container)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)' }}>销售趋势</div>
            <Select
              value={trendDays}
              onChange={setTrendDays}
              style={{ width: 100 }}
              size="small"
              options={[{ value: 7, label: '7天' }, { value: 14, label: '14天' }, { value: 30, label: '30天' }]}
            />
          </div>
          <div style={{ height: 280 }}>
            <Line
              data={salesTrendData}
              xField="date"
              yField="销量"
              smooth
              point={{ size: 3, shape: 'circle' }}
              color="#0066FF"
            />
          </div>
        </div>
        <div style={{ padding: '24px', background: 'var(--color-bg-container)' }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 16 }}>票种分布</div>
          <div style={{ height: 280 }}>
            {ticketTypeData.length > 0 ? (
              <Pie
                data={ticketTypeData}
                angleField="value"
                colorField="type"
                radius={0.8}
                innerRadius={0.6}
                label={{ text: 'type', position: 'outside' }}
                legend={{ color: { position: 'bottom', layout: { justifyContent: 'center' } } }}
              />
            ) : (
              <div style={{ textAlign: 'center', paddingTop: 100, color: 'var(--color-text-tertiary)' }}>暂无数据</div>
            )}
          </div>
        </div>
      </div>

      {/* Funnel */}
      <div style={{ border: '1px solid var(--color-border)', background: 'var(--color-bg-container)', padding: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)' }}>转化漏斗</div>
          <Select
            value={funnelEventId}
            onChange={setFunnelEventId}
            placeholder="选择活动"
            style={{ width: 200 }}
            size="small"
            allowClear
            options={events.map((e) => ({ value: e.id, label: e.title }))}
          />
        </div>
        {funnelEventId && funnel ? (
          <div style={{ display: 'flex', gap: 1, background: 'var(--color-border)', border: '1px solid var(--color-border)' }}>
            {funnelData.map((item, index) => {
              const maxVal = Math.max(...funnelData.map(d => d.count))
              const height = maxVal > 0 ? 60 + (item.count / maxVal) * 120 : 60
              return (
                <div key={item.stage} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                  <div style={{ flex: 1, display: 'flex', alignItems: 'flex-end' }}>
                    <div style={{
                      width: '100%',
                      height,
                      background: 'var(--color-accent)',
                      opacity: 1 - index * 0.15,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontFamily: 'var(--font-mono)',
                      fontSize: 18,
                      fontWeight: 700,
                      color: '#FFFFFF',
                    }}>
                      {item.count}
                    </div>
                  </div>
                  <div style={{ padding: '8px 0', fontSize: 12, color: 'var(--color-text-secondary)', textAlign: 'center' }}>
                    {item.stage}
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: 60, color: 'var(--color-text-tertiary)' }}>请选择活动查看转化漏斗</div>
        )}
      </div>
    </div>
  )
}
