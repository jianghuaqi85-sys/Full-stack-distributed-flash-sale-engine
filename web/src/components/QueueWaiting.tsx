import { useEffect, useState } from 'react'
import { Card, Progress, Typography, Space, Button, Alert } from 'antd'
import { ClockCircleOutlined, LoadingOutlined } from '@ant-design/icons'
import { getQueuePosition, leaveQueue, QueuePosition } from '../api/queue'

const { Title, Text } = Typography

interface QueueWaitingProps {
  eventId: number
  onReady: () => void
  onLeave: () => void
}

export default function QueueWaiting({ eventId, onReady, onLeave }: QueueWaitingProps) {
  const [position, setPosition] = useState<QueuePosition | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchPosition = async () => {
      try {
        const res = await getQueuePosition(eventId)
        setPosition(res.data)

        if (res.data.status === 'ready') {
          onReady()
        }
      } catch {
        // ignore
      } finally {
        setLoading(false)
      }
    }

    fetchPosition()
    const interval = setInterval(fetchPosition, 3000) // 每3秒刷新一次

    return () => clearInterval(interval)
  }, [eventId, onReady])

  const handleLeave = async () => {
    try {
      await leaveQueue(eventId)
      onLeave()
    } catch {
      // ignore
    }
  }

  if (loading) {
    return (
      <Card style={{ textAlign: 'center', padding: '40px 0' }}>
        <LoadingOutlined style={{ fontSize: 48 }} />
        <p style={{ marginTop: 16 }}>正在获取排队信息...</p>
      </Card>
    )
  }

  if (!position) {
    return (
      <Card style={{ textAlign: 'center', padding: '40px 0' }}>
        <Alert message="获取排队信息失败" type="error" showIcon />
      </Card>
    )
  }

  const estimatedMinutes = Math.ceil(position.estimated_wait / 60)
  const progressPercent = position.total_ahead > 0
    ? Math.max(0, 100 - (position.position / (position.position + position.total_ahead)) * 100)
    : 100

  return (
    <Card style={{ textAlign: 'center', padding: '20px 0' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <ClockCircleOutlined style={{ fontSize: 64, color: '#1890ff' }} />
        </div>

        <Title level={3}>排队等待中</Title>

        <div>
          <Text type="secondary">当前排队位置</Text>
          <Title level={1} style={{ margin: '8px 0' }}>{position.position}</Title>
          <Text type="secondary">前面还有 {position.total_ahead} 人</Text>
        </div>

        <Progress
          percent={Math.round(progressPercent)}
          status="active"
          strokeColor={{ from: '#108ee9', to: '#87d068' }}
        />

        <Alert
          message={`预计等待时间：约 ${estimatedMinutes} 分钟`}
          type="info"
          showIcon
        />

        <Alert
          message="请保持页面打开，系统将自动为您购票"
          type="warning"
          showIcon
        />

        <Button type="link" danger onClick={handleLeave}>
          离开队列
        </Button>
      </Space>
    </Card>
  )
}
