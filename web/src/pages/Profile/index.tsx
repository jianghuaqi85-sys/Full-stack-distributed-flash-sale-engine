import { useState } from 'react'
import { Card, Descriptions, Button, Form, Input, Tag, Avatar, message, Tabs } from 'antd'
import { UserOutlined, LockOutlined, CalendarOutlined, TagOutlined, ShopOutlined, SwapOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { changePassword } from '../../api/profile'

export default function Profile() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [passwordModalOpen, setPasswordModalOpen] = useState(false)
  const [form] = Form.useForm()

  const handleChangePassword = async () => {
    try {
      const values = await form.validateFields()
      await changePassword({
        old_password: values.oldPassword,
        new_password: values.newPassword,
      })
      message.success('密码修改成功')
      setPasswordModalOpen(false)
      form.resetFields()
    } catch {
      // validation failed or API error handled by interceptor
    }
  }

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>个人中心</h2>

      <Tabs
        items={[
          {
            key: 'profile',
            label: '基本信息',
            children: (
              <Card>
                <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
                  <div style={{ textAlign: 'center' }}>
                    <Avatar size={100} icon={<UserOutlined />} style={{ background: 'linear-gradient(135deg, #5B2FE8, #D4A843)' }} />
                    <div style={{ marginTop: 16 }}>
                      <Tag color="blue">{user?.role || 'user'}</Tag>
                    </div>
                  </div>
                  <div style={{ flex: 1, minWidth: 300 }}>
                    <Descriptions column={1} bordered size="small">
                      <Descriptions.Item label="用户ID">{user?.id || '-'}</Descriptions.Item>
                      <Descriptions.Item label="用户名">{user?.username || '-'}</Descriptions.Item>
                      <Descriptions.Item label="邮箱">{user?.email || '-'}</Descriptions.Item>
                      <Descriptions.Item label="角色">
                        <Tag color="blue">{user?.role || 'user'}</Tag>
                      </Descriptions.Item>
                    </Descriptions>
                  </div>
                </div>
              </Card>
            ),
          },
          {
            key: 'security',
            label: '安全设置',
            children: (
              <Card>
                <Descriptions column={1} bordered size="small">
                  <Descriptions.Item label="登录密码">
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span>已设置</span>
                      <Button type="link" icon={<LockOutlined />} onClick={() => setPasswordModalOpen(true)}>
                        修改密码
                      </Button>
                    </div>
                  </Descriptions.Item>
                  <Descriptions.Item label="两步验证">
                    <span style={{ color: 'var(--color-text-tertiary)' }}>未开启</span>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            ),
          },
          {
            key: 'shortcut',
            label: '快捷操作',
            children: (
              <Card>
                <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
                  <Button icon={<CalendarOutlined />} onClick={() => navigate('/events')}>
                    活动中心
                  </Button>
                  <Button icon={<TagOutlined />} onClick={() => navigate('/tickets')}>
                    我的票务
                  </Button>
                  <Button icon={<ShopOutlined />} onClick={() => navigate('/marketplace')}>
                    二手市场
                  </Button>
                  <Button icon={<SwapOutlined />} onClick={() => navigate('/transfer-records')}>
                    转让记录
                  </Button>
                </div>
              </Card>
            ),
          },
        ]}
      />

      <Card title="修改密码" style={{ marginTop: 16 }} hidden={!passwordModalOpen}>
        <Form form={form} layout="vertical" onFinish={handleChangePassword}>
          <Form.Item name="oldPassword" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
            <Input.Password placeholder="请输入当前密码" />
          </Form.Item>
          <Form.Item name="newPassword" label="新密码" rules={[{ required: true, min: 6, message: '密码至少6位' }]}>
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password placeholder="请确认新密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit">保存</Button>
            <Button style={{ marginLeft: 8 }} onClick={() => { setPasswordModalOpen(false); form.resetFields() }}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
