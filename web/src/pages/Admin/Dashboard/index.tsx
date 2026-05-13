import { useEffect, useState } from 'react'
import { Card, Row, Col, Statistic, Spin, Select, Typography } from 'antd'
import { CalendarOutlined, DollarOutlined, RiseOutlined } from '@ant-design/icons'
import { Line, Pie } from '@ant-design/charts'
import { getDashboardStats, getSalesTrend, getTicketTypeStats, DashboardStats, SalesTrend, TicketTypeStats } from '../../../api/stats'

const { Title } = Typography

export default function AdminDashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [salesTrend, setSalesTrend] = useState<SalesTrend[]>([])
  const [ticketTypeStats, setTicketTypeStats] = useState<TicketTypeStats[]>([])
  const [loading, setLoading] = useState(true)
  const [trendDays, setTrendDays] = useState(7)

  useEffect(() => {
    fetchData()
  }, [trendDays])

  const fetchData = async () => {
    setLoading(true)
    try {
      const [statsRes, trendRes, ticketRes] = await Promise.all([
        getDashboardStats(),
        getSalesTrend(trendDays),
        getTicketTypeStats(),
      ])
      setStats(statsRes.data)
      setSalesTrend(trendRes.data.data)
      setTicketTypeStats(ticketRes.data.data)
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  const salesTrendData = salesTrend.map(item => ({
    date: item.date,
    销量: item.count,
    收入: item.revenue,
  }))

  const ticketTypeData = ticketTypeStats.map(item => ({
    type: item.ticket_type_name,
    value: item.sold_count,
  }))

  const salesTrendConfig = {
    data: salesTrendData,
    xField: 'date',
    yField: '销量',
    smooth: true,
    point: { size: 3, shape: 'circle' },
    color: '#5B2FE8',
  }

  const ticketTypeConfig = {
    data: ticketTypeData,
    angleField: 'value',
    colorField: 'type',
    radius: 0.8,
    innerRadius: 0.6,
    label: {
      text: 'type',
      position: 'outside',
    },
    legend: {
      color: {
        position: 'bottom',
        layout: { justifyContent: 'center' },
      },
    },
  }

  if (loading) {
    return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>数据仪表盘</Title>
        <Select
          value={trendDays}
          onChange={setTrendDays}
          style={{ width: 120 }}
          options={[
            { value: 7, label: '近7天' },
            { value: 14, label: '近14天' },
            { value: 30, label: '近30天' },
          ]}
        />
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活动总数"
              value={stats?.total_events || 0}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="售票中活动"
              value={stats?.active_events || 0}
              prefix={<CalendarOutlined />}
              valueStyle={{ color: 'var(--color-success)' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="已售票数"
              value={stats?.sold_tickets || 0}
              suffix={`/ ${stats?.total_tickets || 0}`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总收入"
              value={stats?.total_revenue || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="元"
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日销量"
              value={stats?.today_sales || 0}
              prefix={<RiseOutlined />}
              valueStyle={{ color: 'var(--color-primary)' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日收入"
              value={stats?.today_revenue || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="元"
              valueStyle={{ color: 'var(--color-gold)' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待支付"
              value={stats?.reserved_tickets || 0}
              valueStyle={{ color: 'var(--color-gold)' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="转化率"
              value={stats?.total_tickets ? ((stats?.sold_tickets || 0) / stats.total_tickets * 100) : 0}
              precision={1}
              suffix="%"
              valueStyle={{ color: 'var(--color-success)' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card title="销售趋势">
            <div style={{ height: 300 }}>
              <Line {...salesTrendConfig} />
            </div>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="票种分布">
            <div style={{ height: 300 }}>
              {ticketTypeData.length > 0 ? (
                <Pie {...ticketTypeConfig} />
              ) : (
                <div style={{ textAlign: 'center', paddingTop: 100, color: 'var(--color-text-tertiary)' }}>
                  暂无数据
                </div>
              )}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
