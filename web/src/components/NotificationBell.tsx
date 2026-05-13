import { Badge, Button, Dropdown, List, Typography, Empty, Space } from 'antd'
import { BellOutlined, CheckOutlined, DeleteOutlined } from '@ant-design/icons'
import { useNotificationStore } from '../stores/notificationStore'
import type { Notification } from '../stores/notificationStore'

const { Text } = Typography

const typeColors: Record<string, string> = {
  success: 'var(--color-success)',
  warning: 'var(--color-warning)',
  info: 'var(--color-info)',
  error: 'var(--color-error)',
}

export default function NotificationBell() {
  const { notifications, unreadCount, markAsRead, markAllAsRead, clearAll } = useNotificationStore()

  const formatTime = (time: number) => {
    const diff = Date.now() - time
    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
    return new Date(time).toLocaleDateString('zh-CN')
  }

  const dropdownContent = (
    <div style={{ width: 360, maxHeight: 400, overflow: 'auto' }}>
      <div style={{ padding: '8px 16px', borderBottom: '1px solid var(--color-border)', display: 'flex', justifyContent: 'space-between' }}>
        <Text strong>消息中心</Text>
        <Space>
          {unreadCount > 0 && (
            <Button type="link" size="small" icon={<CheckOutlined />} onClick={markAllAsRead}>
              全部已读
            </Button>
          )}
          <Button type="link" size="small" icon={<DeleteOutlined />} onClick={clearAll}>
            清空
          </Button>
        </Space>
      </div>
      {notifications.length === 0 ? (
        <Empty description="暂无消息" style={{ padding: 40 }} />
      ) : (
        <List
          dataSource={notifications}
          renderItem={(item: Notification) => (
            <List.Item
              style={{
                padding: '12px 16px',
                cursor: 'pointer',
                background: item.read ? 'transparent' : 'var(--color-success-bg)',
                transition: 'background 0.2s',
              }}
              onClick={() => markAsRead(item.id)}
            >
              <List.Item.Meta
                avatar={
                  <div style={{ width: 8, height: 8, borderRadius: '50%', background: typeColors[item.type], marginTop: 8 }} />
                }
                title={<Text strong={!item.read}>{item.title}</Text>}
                description={
                  <div>
                    <Text>{item.message}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>{formatTime(item.time)}</Text>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}
    </div>
  )

  return (
    <Dropdown dropdownRender={() => dropdownContent} trigger={['click']} placement="bottomRight">
      <Badge count={unreadCount} size="small">
        <Button type="text" icon={<BellOutlined />} style={{ fontSize: 18 }} />
      </Badge>
    </Dropdown>
  )
}
