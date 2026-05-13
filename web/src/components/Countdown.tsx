import { useEffect, useState } from 'react'
import { Statistic, Space } from 'antd'
import { ClockCircleOutlined } from '@ant-design/icons'

const { Countdown: AntCountdown } = Statistic

interface CountdownProps {
  targetDate: string | Date
  onFinish?: () => void
  format?: string
  size?: 'small' | 'default' | 'large'
}

export default function Countdown({
  targetDate,
  onFinish,
  format = 'HH:mm:ss',
  size = 'default'
}: CountdownProps) {
  const [target, setTarget] = useState<number>(0)

  useEffect(() => {
    const date = new Date(targetDate)
    setTarget(date.getTime())
  }, [targetDate])

  const handleChange = () => {
    onFinish?.()
  }

  if (!target) return null

  return (
    <Space direction="vertical" align="center">
      <ClockCircleOutlined style={{ fontSize: size === 'large' ? 48 : 24, color: 'var(--color-gold)' }} />
      <AntCountdown
        value={target}
        format={format}
        onFinish={handleChange}
        valueStyle={{
          fontSize: size === 'large' ? 48 : size === 'small' ? 16 : 24,
          color: 'var(--color-gold)',
          fontWeight: 'bold',
          fontVariantNumeric: 'tabular-nums',
        }}
      />
    </Space>
  )
}
