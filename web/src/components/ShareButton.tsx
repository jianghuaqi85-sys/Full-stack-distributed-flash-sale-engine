import { useState } from 'react'
import { Button, Modal, Input, message, Space, Typography } from 'antd'
import { ShareAltOutlined, CopyOutlined } from '@ant-design/icons'

const { Text, Paragraph } = Typography

interface ShareButtonProps {
  eventId: number
  eventTitle: string
  eventDescription?: string
}

export default function ShareButton({ eventId, eventTitle, eventDescription }: ShareButtonProps) {
  const [modalVisible, setModalVisible] = useState(false)
  const shareUrl = `${window.location.origin}/events/${eventId}`

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl)
      message.success('链接已复制到剪贴板')
    } catch {
      message.error('复制失败，请手动复制')
    }
  }

  const handleShare = (platform: string) => {
    const text = `${eventTitle} - 精彩活动等你来！`
    const url = shareUrl

    let shareUrlFinal = ''

    switch (platform) {
      case 'wechat':
        // 微信分享需要显示二维码
        Modal.info({
          title: '分享到微信',
          content: (
            <div style={{ textAlign: 'center' }}>
              <p>请使用微信扫描下方二维码分享</p>
              <div style={{ padding: 20, background: '#f5f5f5', borderRadius: 8 }}>
                <Text type="secondary">二维码占位</Text>
              </div>
            </div>
          ),
        })
        return
      case 'weibo':
        shareUrlFinal = `https://service.weibo.com/share/share.php?url=${encodeURIComponent(url)}&title=${encodeURIComponent(text)}`
        break
      case 'qq':
        shareUrlFinal = `https://connect.qq.com/widget/shareqq/index.html?url=${encodeURIComponent(url)}&title=${encodeURIComponent(text)}&desc=${encodeURIComponent(eventDescription || text)}`
        break
      case 'twitter':
        shareUrlFinal = `https://twitter.com/intent/tweet?url=${encodeURIComponent(url)}&text=${encodeURIComponent(text)}`
        break
      case 'facebook':
        shareUrlFinal = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`
        break
    }

    if (shareUrlFinal) {
      window.open(shareUrlFinal, '_blank', 'width=600,height=400')
    }
  }

  return (
    <>
      <Button icon={<ShareAltOutlined />} onClick={() => setModalVisible(true)}>
        分享
      </Button>

      <Modal
        title="分享活动"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Text type="secondary">活动链接</Text>
            <Input.Group compact style={{ marginTop: 8 }}>
              <Input
                style={{ width: 'calc(100% - 80px)' }}
                value={shareUrl}
                readOnly
              />
              <Button icon={<CopyOutlined />} onClick={handleCopyLink}>
                复制
              </Button>
            </Input.Group>
          </div>

          <div>
            <Text type="secondary">分享到</Text>
            <div style={{ marginTop: 8 }}>
              <Space wrap>
                <Button
                  style={{ background: '#07c160', color: '#fff', border: 'none' }}
                  onClick={() => handleShare('wechat')}
                >
                  微信
                </Button>
                <Button
                  style={{ background: '#e6162d', color: '#fff', border: 'none' }}
                  onClick={() => handleShare('weibo')}
                >
                  微博
                </Button>
                <Button
                  style={{ background: '#12b7f5', color: '#fff', border: 'none' }}
                  onClick={() => handleShare('qq')}
                >
                  QQ
                </Button>
                <Button
                  style={{ background: '#1da1f2', color: '#fff', border: 'none' }}
                  onClick={() => handleShare('twitter')}
                >
                  Twitter
                </Button>
                <Button
                  style={{ background: '#1877f2', color: '#fff', border: 'none' }}
                  onClick={() => handleShare('facebook')}
                >
                  Facebook
                </Button>
              </Space>
            </div>
          </div>

          <div>
            <Text type="secondary">活动简介</Text>
            <Paragraph
              copyable
              style={{ marginTop: 8, padding: 12, background: '#f5f5f5', borderRadius: 6 }}
            >
              {eventTitle}{eventDescription ? ` - ${eventDescription}` : ''} {shareUrl}
            </Paragraph>
          </div>
        </Space>
      </Modal>
    </>
  )
}
