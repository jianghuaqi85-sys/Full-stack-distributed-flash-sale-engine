import { useState } from 'react'
import { Button, Form, Input, message, Avatar } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { changePassword } from '../../api/profile'

export default function Profile() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [showPasswordForm, setShowPasswordForm] = useState(false)
  const [form] = Form.useForm()

  const handleChangePassword = async () => {
    try {
      const values = await form.validateFields()
      await changePassword({ old_password: values.oldPassword, new_password: values.newPassword })
      message.success('密码修改成功')
      setShowPasswordForm(false)
      form.resetFields()
    } catch { /* handled */ }
  }

  return (
    <div className="page-enter">
      <div className="page-header">
        <h1>个人</h1>
        <div className="subtitle">账户设置</div>
      </div>

      {/* User info */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 20, marginBottom: 24 }}>
          <Avatar size={64} style={{ background: '#0066FF', fontSize: 24, fontWeight: 700 }}>
            {user?.username?.[0]?.toUpperCase()}
          </Avatar>
          <div>
            <div style={{ fontSize: 24, fontWeight: 700, letterSpacing: '-0.02em' }}>{user?.username}</div>
            <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase' }}>
              {user?.role || 'user'}
            </div>
          </div>
        </div>

        {/* Info grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
          gap: 1,
          background: 'var(--color-border)',
          border: '1px solid var(--color-border)',
        }}>
          <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>用户ID</div>
            <div style={{ fontSize: 15, fontWeight: 500, fontFamily: 'var(--font-mono)' }}>{user?.id || '-'}</div>
          </div>
          <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>用户名</div>
            <div style={{ fontSize: 15, fontWeight: 500 }}>{user?.username || '-'}</div>
          </div>
          <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>邮箱</div>
            <div style={{ fontSize: 15, fontWeight: 500 }}>{user?.email || '-'}</div>
          </div>
          <div style={{ padding: '20px 24px', background: 'var(--color-bg-container)' }}>
            <div style={{ fontSize: 12, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 4 }}>角色</div>
            <div style={{ fontSize: 15, fontWeight: 500 }}>{user?.role || 'user'}</div>
          </div>
        </div>
      </div>

      {/* Security */}
      <div style={{ marginBottom: 32 }}>
        <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 12 }}>安全</div>
        <div style={{ border: '1px solid var(--color-border)', padding: '20px', background: 'var(--color-bg-container)' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div style={{ fontWeight: 500 }}>登录密码</div>
              <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)' }}>已设置</div>
            </div>
            <Button onClick={() => setShowPasswordForm(!showPasswordForm)}>
              {showPasswordForm ? '取消' : '修改'}
            </Button>
          </div>
          {showPasswordForm && (
            <Form form={form} layout="vertical" onFinish={handleChangePassword} style={{ marginTop: 16, maxWidth: 400 }}>
              <Form.Item name="oldPassword" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item name="newPassword" label="新密码" rules={[{ required: true, min: 6, message: '密码至少6位' }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item
                name="confirmPassword"
                label="确认密码"
                dependencies={['newPassword']}
                rules={[
                  { required: true, message: '请确认密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('newPassword') === value) return Promise.resolve()
                      return Promise.reject(new Error('两次密码不一致'))
                    },
                  }),
                ]}
              >
                <Input.Password />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit">保存</Button>
              </Form.Item>
            </Form>
          )}
        </div>
      </div>

      {/* Quick nav */}
      <div>
        <div style={{ fontSize: 13, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', fontFamily: 'var(--font-mono)', marginBottom: 12 }}>导航</div>
        <div style={{ display: 'flex', gap: 1, background: 'var(--color-border)', border: '1px solid var(--color-border)' }}>
          {[
            { path: '/events', label: '活动' },
            { path: '/tickets', label: '票务' },
            { path: '/marketplace', label: '市场' },
            { path: '/transfer-records', label: '转让' },
          ].map(item => (
            <button
              key={item.path}
              onClick={() => navigate(item.path)}
              style={{
                flex: 1,
                padding: '16px',
                border: 'none',
                background: 'var(--color-bg-container)',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500,
                color: 'var(--color-text-primary)',
                transition: 'background 0.1s',
              }}
              onMouseEnter={e => e.currentTarget.style.background = 'var(--color-accent-soft)'}
              onMouseLeave={e => e.currentTarget.style.background = 'var(--color-bg-container)'}
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
