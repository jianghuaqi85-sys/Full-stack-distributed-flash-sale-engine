import { useState, useEffect } from 'react'
import { Button, Modal, message } from 'antd'
import { TeamOutlined } from '@ant-design/icons'
import { joinWaitlist, getWaitlistPosition, leaveWaitlist } from '../api/waitlist'

interface WaitlistButtonProps {
  eventId: number
  isSoldOut: boolean
  onJoinSuccess?: () => void
}

export default function WaitlistButton({ eventId, isSoldOut, onJoinSuccess }: WaitlistButtonProps) {
  const [loading, setLoading] = useState(false)
  const [inWaitlist, setInWaitlist] = useState(false)
  const [position, setPosition] = useState(0)

  useEffect(() => {
    checkWaitlistStatus()
  }, [eventId])

  const checkWaitlistStatus = async () => {
    try {
      const res = await getWaitlistPosition(eventId)
      setInWaitlist(true)
      setPosition(res.data.position)
    } catch {
      setInWaitlist(false)
    }
  }

  const handleJoin = async () => {
    Modal.confirm({
      title: '加入等候名单',
      content: '活动已售罄，加入等候名单后，有人退票时将优先通知您。',
      onOk: async () => {
        setLoading(true)
        try {
          const res = await joinWaitlist(eventId)
          message.success(res.data.message || '已加入等候名单')
          setInWaitlist(true)
          setPosition(res.data.position)
          onJoinSuccess?.()
        } catch (error: unknown) {
          const err = error as { response?: { data?: { error?: string } } }
          message.error(err.response?.data?.error || '加入失败')
        } finally {
          setLoading(false)
        }
      },
    })
  }

  const handleLeave = async () => {
    Modal.confirm({
      title: '离开等候名单',
      content: '确定要离开等候名单吗？',
      onOk: async () => {
        setLoading(true)
        try {
          await leaveWaitlist(eventId)
          message.success('已离开等候名单')
          setInWaitlist(false)
          setPosition(0)
        } catch {
          message.error('操作失败')
        } finally {
          setLoading(false)
        }
      },
    })
  }

  if (!isSoldOut) return null

  if (inWaitlist) {
    return (
      <Button
        icon={<TeamOutlined />}
        onClick={handleLeave}
        loading={loading}
      >
        等候名单中（第 {position} 位）
      </Button>
    )
  }

  return (
    <Button
      type="primary"
      icon={<TeamOutlined />}
      onClick={handleJoin}
      loading={loading}
    >
      加入等候名单
    </Button>
  )
}
